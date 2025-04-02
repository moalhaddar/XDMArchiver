package dlmanager

import (
	"bufio"
	"dmarchiver/logger"
	"dmarchiver/twitter"
	"dmarchiver/utils"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
)

type MediaUnit struct {
	URL       string
	Filename  string
	MediaType string
}

type DLManager struct {
	TwitterCtx     twitter.TwitterContext
	ConversationId string
	EventsPath     string
	PhotosPath     string
	VideosPath     string
	MaxEntryId     *string
	CurrentEvent   *twitter.ConversationResponse
	Events         []twitter.ConversationResponse
	Entries        []twitter.Entry
	EntriesContMap map[string]string
	MediaURLsQueue chan MediaUnit
	IsDebug        bool
}

const (
	CONVER_DIR = "conversations"
	EVENTS_DIR = "events"
	PHOTOS_DIR = "photos"
	VIDEOS_DIR = "videos"
	AT_END     = "AT_END"
)

func InitDLManager(conversationId string, twitterCtx twitter.TwitterContext, isDebug bool) (*DLManager, error) {
	queue := make(chan MediaUnit, 256)
	dlManager := DLManager{
		TwitterCtx:     twitterCtx,
		ConversationId: conversationId,
		MaxEntryId:     nil,
		CurrentEvent:   nil,
		MediaURLsQueue: queue,
		IsDebug:        isDebug,
		EventsPath:     filepath.Join(CONVER_DIR, conversationId, EVENTS_DIR),
		PhotosPath:     filepath.Join(CONVER_DIR, conversationId, PHOTOS_DIR),
		VideosPath:     filepath.Join(CONVER_DIR, conversationId, VIDEOS_DIR),
		Events:         nil,
		Entries:        nil,
		EntriesContMap: nil,
	}

	err := dlManager.loadEvents()
	if err != nil {
		return nil, err
	}

	err = dlManager.loadEntriesFromEvents()
	if err != nil {
		return nil, err
	}

	logger.EventsLogger.Printf("Total loaded events: %d\n", len(dlManager.Events))
	logger.EventsLogger.Printf("Total loaded entries: %d\n", len(dlManager.Entries))
	logger.EventsLogger.Printf("URLs to be downloaded: %d\n", len(dlManager.MediaURLsQueue))

	return &dlManager, nil
}

func (dlManager *DLManager) loadEvents() error {
	files, err := os.ReadDir(dlManager.EventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return fmt.Errorf("failed to load events from dir %s", dlManager.EventsPath)
		}
	}

	events := make([]twitter.ConversationResponse, 0, len(files))
	for _, file := range files {
		eventPath := filepath.Join(dlManager.EventsPath, file.Name())
		eventFile, err := os.Open(eventPath)
		if err != nil {
			return fmt.Errorf("failed to load events from file %s", eventPath)
		}
		var event twitter.ConversationResponse
		err = json.NewDecoder(eventFile).Decode(&event)
		if err != nil {
			return fmt.Errorf("failed to json decode from file %s", eventPath)
		}
		logger.EventsLogger.Printf("\tLoaded events from %s\n", file)
		dlManager.extractUrlsFromEvent(event)
		events = append(events, event)
	}

	dlManager.Events = events
	return nil
}

func (dlManager *DLManager) loadEntriesFromEvents() error {
	entries := make([]twitter.Entry, 0)
	entriesContMap := make(map[string]string)
	for _, event := range dlManager.Events {
		eventEntries := event.GetEntries()
		for i := range eventEntries {
			entries = append(entries, eventEntries[i])

			nextEntryTimestamp := ""
			if i+1 < len(eventEntries) {
				nextEntryTimestamp = eventEntries[i+1].Message.Time
			}
			if eventEntries[i].Message.Time == nextEntryTimestamp {
				logger.EventsLogger.Fatalf("Found a loop in the contiuation map with id %s\n", nextEntryTimestamp)
			}
			val, isSet := entriesContMap[eventEntries[i].Message.Time]
			if (isSet && val == "") || !isSet {
				entriesContMap[eventEntries[i].Message.Time] = nextEntryTimestamp
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		timeA, err := strconv.ParseInt(entries[i].Message.Time, 10, 64)
		if err != nil {
			logger.EventsLogger.Fatalf("Invalid timestamp when sorting the entries %s.", entries[i].Message.Time)
		}
		timeB, err := strconv.ParseInt(entries[j].Message.Time, 10, 64)
		if err != nil {
			logger.EventsLogger.Fatalf("Invalid timestamp when sorting the entries %s.", entries[i].Message.Time)
		}
		return time.UnixMilli(timeA).Before(time.UnixMilli(timeB))
	})

	dlManager.Entries = entries
	dlManager.EntriesContMap = entriesContMap

	return nil
}

func (dlManager *DLManager) saveCurrentEvent() error {
	err := os.MkdirAll(dlManager.EventsPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var maxId string
	if dlManager.MaxEntryId != nil {
		maxId = *dlManager.MaxEntryId
	} else {
		maxEntry := dlManager.CurrentEvent.GetMaxEntry()
		maxId = maxEntry.GetEntryId()
	}

	eventPath := filepath.Join(dlManager.EventsPath, maxId+".json")
	file, err := os.Create(eventPath)
	if err != nil {
		return fmt.Errorf("failed to create event file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(dlManager.CurrentEvent); err != nil {
		return fmt.Errorf("failed to encode conversation: %w", err)
	}
	logger.EventsLogger.Printf("\tSuccessfully saved event: %s", eventPath)

	return nil
}

func (dlManager *DLManager) printStats() {
	entries := dlManager.CurrentEvent.GetEntries()
	oldestEntry := entries[len(entries)-1]
	oldestEntryTimeFormatted, err := utils.FormatUnixTimestamp(oldestEntry.Message.Time, true)
	if dlManager.MaxEntryId != nil {
		logger.EventsLogger.Printf("\tResult for max_entry_id: %s\n", *dlManager.MaxEntryId)
	} else {
		logger.EventsLogger.Print("\tResult for max_entry_id: nil\n")
	}
	logger.EventsLogger.Printf("\tEvents count %d\n", len(entries))
	if err != nil {
		logger.EventsLogger.Printf("\tOldest Message Date & Time: %v\n", err)
	} else {
		logger.EventsLogger.Printf("\tOldest Message Date & Time: %s\n", oldestEntryTimeFormatted)
	}
}

func (dlManager *DLManager) followTimestampChain(timestamp *string) int {
	iterations := 0
	for {
		next, exists := dlManager.EntriesContMap[*timestamp]
		if !exists || next == "" {
			break
		}
		if dlManager.IsDebug {
			logger.EventsLogger.Printf("\tFound message with id %s in data. Skipping to %s", *timestamp, next)
		}
		*timestamp = next
		iterations++
	}
	return iterations
}

func (dlManager *DLManager) setMinEntryIdAsMax() {
	nextMaxEntry := dlManager.CurrentEvent.GetMinEntry().Message.Time
	dlManager.MaxEntryId = &nextMaxEntry
}

func (dlManager *DLManager) setNextMaxEntryId() {
	nextEntryTimestamp := dlManager.CurrentEvent.GetMaxEntry().Message.Time
	iterations := dlManager.followTimestampChain(&nextEntryTimestamp)
	if iterations == 0 {
		logger.EventsLogger.Printf("\tZero iterations for %s\n", nextEntryTimestamp)
		dlManager.setMinEntryIdAsMax()
		return
	}

	t, err := utils.UnixTimestampStringToTime(nextEntryTimestamp, true)
	if err != nil {
		logger.EventsLogger.Fatalf("Expected a non empty unix timestamp, got %s\n", nextEntryTimestamp)
	}

	if dlManager.IsDebug {
		logger.EventsLogger.Printf("\tLatest timestamp is %s: %s\n", nextEntryTimestamp, t.Local().Format(time.DateTime))
	}

	newMaxEntry := twitter.EncodeFakeSnowflakeFromTimestamp(*t)
	if dlManager.MaxEntryId != nil && newMaxEntry == *dlManager.MaxEntryId {
		logger.EventsLogger.Fatalf("Trying to get the same max entry id twice: %s", *dlManager.MaxEntryId)
	}

	dlManager.MaxEntryId = &newMaxEntry
}

func (dlManager *DLManager) extractUrlsFromEvent(event twitter.ConversationResponse) {
	urls := make([]MediaUnit, 0, 10)
	for _, entry := range event.GetEntries() {
		if entry.Message.MessageData.Attachment != nil {
			stamp, _ := utils.FormatUnixTimestamp(entry.Message.MessageData.Time, true)
			vars := entry.Message.MessageData.Attachment.Video.VideoInfo.Variants
			sort.Slice(vars, func(i, j int) bool {
				return vars[i].Bitrate > vars[j].Bitrate
			})
			for _, v := range vars {
				if v.ContentType == "video/mp4" {
					urls = append(urls, MediaUnit{
						URL:       v.URL,
						Filename:  fmt.Sprintf("%s-%d.mp4", stamp, v.Bitrate),
						MediaType: "Video",
					})
					break
				}
			}
			photoUrl := entry.Message.MessageData.Attachment.Photo.MediaURLHTTPS
			if photoUrl != "" {
				urls = append(urls, MediaUnit{
					URL:       photoUrl,
					Filename:  fmt.Sprintf("%s.jpg\n", stamp),
					MediaType: "Photo",
				})
			}
		}
	}

	for _, url := range urls {
		dlManager.MediaURLsQueue <- url
	}
}

func (dlManager *DLManager) downloadEvents() {
	reader := bufio.NewReader(os.Stdin)

	for iteration := 0; ; iteration++ {
		if dlManager.IsDebug {
			logger.EventsLogger.Printf("Press any key to continue")
			reader.ReadRune()
		}
		logger.EventsLogger.Printf("Iteration #%d\n", iteration)

		event, rateLimit, err := dlManager.TwitterCtx.GetConversation(dlManager.MaxEntryId)
		if rateLimit != nil {
			rateLimit.Print("\t")
		}
		if err != nil {
			var statusError *twitter.ErrNot200
			if errors.As(err, &statusError) {
				if statusError.StatusCode == http.StatusTooManyRequests {
					utils.SleepUntil(*rateLimit.RateLimitResetTime)
				}
				if statusError.StatusCode == http.StatusForbidden {
					logger.EventsLogger.Printf("The user is unauthorized.\n")
					logger.EventsLogger.Printf("The auth data must be provided explicitly through the CLI.\n")
					logger.EventsLogger.Printf("If provided, that means that the data is not supplied correctly. Exiting..\n")
					flag.PrintDefaults()
					os.Exit(1)
				}
			} else {
				logger.EventsLogger.Printf("Error while downloading conversation: %s\n", err)
			}
			continue
		}
		dlManager.CurrentEvent = event
		dlManager.extractUrlsFromEvent(*event)
		dlManager.saveCurrentEvent()
		dlManager.printStats()
		dlManager.setNextMaxEntryId()
		logger.EventsLogger.Printf("\tNext max entry is %s\n", *dlManager.MaxEntryId)
		logger.EventsLogger.Printf("\tNext max entry timestamp is %d\n", twitter.DecodeSnowflake(*dlManager.MaxEntryId).Timestamp.UnixMilli())

		if event.ConversationTimeline.Status == AT_END {
			logger.EventsLogger.Printf("\tDone.\n")
			break
		}
	}

	close(dlManager.MediaURLsQueue)
}

func (dlManager *DLManager) downloadMedia() {
	err := os.MkdirAll(dlManager.PhotosPath, 0755)
	if err != nil {
		logger.MediaLogger.Fatalf("failed to create directory: %+v\n", err)
	}
	err = os.MkdirAll(dlManager.VideosPath, 0755)
	if err != nil {
		logger.MediaLogger.Fatalf("failed to create directory: %+v\n", err)
	}
	for {
		unit, ok := <-dlManager.MediaURLsQueue
		if ok {
			var path string
			if unit.MediaType == "Photo" {
				path = filepath.Join(dlManager.PhotosPath, unit.Filename)
			} else {
				path = filepath.Join(dlManager.VideosPath, unit.Filename)
			}
			if utils.FileExists(path) {
				logger.MediaLogger.Printf("File %s exists. Skipping\n", unit.Filename)
				continue
			}
			logger.MediaLogger.Printf("Downloading URL: %s\n", unit.URL)
			bytes, err := dlManager.TwitterCtx.GetFile(unit.URL)
			if err != nil {
				logger.MediaLogger.Printf("Failed to download url %s: %+v\n", unit.URL, err)
				continue
			}
			err = os.WriteFile(path, bytes, 0644)
			if err != nil {
				logger.MediaLogger.Printf("Failed to write file %s to FS: %+v\n", unit.Filename, err)
			} else {
				logger.MediaLogger.Printf("Downloaded %s successfully\n", unit.Filename)
			}
		} else {
			logger.MediaLogger.Printf("Done downloading")
			return
		}
	}
}

func (dlManager *DLManager) Start() {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		dlManager.downloadEvents()
		wg.Done()
	}()
	go func() {
		dlManager.downloadMedia()
		wg.Done()
	}()
	wg.Wait()
}
