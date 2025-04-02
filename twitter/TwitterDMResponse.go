package twitter

import (
	"XDMArchiver/logger"
	"XDMArchiver/utils"
)

// Root structure for the entire response
type ConversationResponse struct {
	ConversationTimeline ConversationTimeline `json:"conversation_timeline"`
}

/*
Min/Max entry are disabled because of their weird behavior.
Example: decoding 1907055874461565247 yields timestamp 1743512526513
However, the messages contain 1743512526556, notice that the [1743512526556] is shared.
The messages differ in their timestamp by some microsecond precision, which introduces
weird behavior and bugs when depending on them. So disabling them is the best option.
*/
type ConversationTimeline struct {
	Status string `json:"status"`
	// MinEntryID string          `json:"min_entry_id"`
	// MaxEntryID string          `json:"max_entry_id"`
	Entries []Entry         `json:"entries"`
	Users   map[string]User `json:"users"`
}

type User struct {
	ID                     int64                  `json:"id"`
	IDStr                  string                 `json:"id_str"`
	Name                   string                 `json:"name"`
	ScreenName             string                 `json:"screen_name"`
	Location               *string                `json:"location"`
	Description            string                 `json:"description"`
	URL                    *string                `json:"url"`
	Entities               map[string]interface{} `json:"entities"`
	Protected              bool                   `json:"protected"`
	FollowersCount         int                    `json:"followers_count"`
	FriendsCount           int                    `json:"friends_count"`
	ListedCount            int                    `json:"listed_count"`
	CreatedAt              string                 `json:"created_at"`
	FavouritesCount        int                    `json:"favourites_count"`
	UTCOffset              *int                   `json:"utc_offset"`
	TimeZone               *string                `json:"time_zone"`
	GeoEnabled             bool                   `json:"geo_enabled"`
	Verified               bool                   `json:"verified"`
	StatusesCount          int                    `json:"statuses_count"`
	Lang                   *string                `json:"lang"`
	ContributorsEnabled    bool                   `json:"contributors_enabled"`
	IsTranslator           bool                   `json:"is_translator"`
	IsTranslationEnabled   bool                   `json:"is_translation_enabled"`
	ProfileBackgroundColor string                 `json:"profile_background_color"`
}

// Entry represents a message in the conversation
type Entry struct {
	Message Message `json:"message"`
}

// Message contains the actual message data
type Message struct {
	Time        string      `json:"time"`
	MessageData MessageData `json:"message_data"`
}

// MessageData contains the content of the message
type MessageData struct {
	Time       string      `json:"time"`
	SenderID   string      `json:"sender_id"`
	Text       string      `json:"text"`
	Attachment *Attachment `json:"attachment"`
}

// Attachment contains information about attached content
type Attachment struct {
	Photo Photo `json:"photo"`
	Video Video `json:"video"`
}

// Video contains video metadata
type Video struct {
	MediaURLHTTPS string `json:"media_url_https"`
	VideoInfo     struct {
		Variants []struct {
			ContentType string `json:"content_type"`
			URL         string `json:"url"`
			Bitrate     int    `json:"bitrate,omitempty"`
		} `json:"variants"`
	} `json:"video_info"`
}

// Photo contains photo metadata
type Photo struct {
	MediaURLHTTPS string `json:"media_url_https"`
}

func (res *ConversationResponse) GetEntries() []Entry {
	filteredEntries := make([]Entry, 0)
	for _, entry := range res.ConversationTimeline.Entries {
		if entry.Message.Time != "" {
			filteredEntries = append(filteredEntries, entry)
		}
	}
	return filteredEntries
}

func (res *ConversationResponse) GetMaxEntry() Entry {
	entries := res.GetEntries()
	if len(entries) == 0 {
		logger.MediaLogger.Fatalf("Trying to get max entry from a response with no entries.. %+v\n", res.ConversationTimeline.Entries)
	}
	return entries[0]
}

func (res *ConversationResponse) GetMinEntry() Entry {
	entries := res.GetEntries()
	if len(entries) == 0 {
		logger.MediaLogger.Fatal("Trying to get min entry from a response with no entries..")
	}
	return entries[len(entries)-1]
}

func (entry *Entry) GetEntryId() string {
	t, _ := utils.UnixTimestampStringToTime(entry.Message.Time, true)
	return EncodeFakeSnowflakeFromTimestamp(*t)
}
