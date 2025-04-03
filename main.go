package main

import (
	"XDMArchiver/dlmanager"
	"XDMArchiver/logger"
	"XDMArchiver/twitter"
	"flag"
	"fmt"
	"os"
)

const (
	version = "v1.1.0"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s (version %s):\n", os.Args[0], version)
		fmt.Printf("\t%s --conversation-id [--auth-headers FILE] [--download-videos] [--download-photos] [--debug]\n", os.Args[0])
		flag.PrintDefaults()
	}
	showVersion := flag.Bool("version", false, "Display version information")
	isDebug := flag.Bool("debug", false, "Enable debugging mode")
	conversationId := flag.String("conversation-id", "", "ID for the conversation to be downloaded")
	downloadVideos := flag.Bool("download-videos", false, "To download videos in the conversation")
	downloadPhotos := flag.Bool("download-photos", false, "To download photos in the conversation")
	authHeaderPath := flag.String("auth-headers", "./auth.txt", "File path to authorization headers to be passed to each request\n"+
		"Headers are newline seperated, each header key value are colon seperated\n"+
		"Example file:\n\tCookie: ABCD\n\tContent-Type: application/json")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s version %s\n", os.Args[0], version)
		os.Exit(0)
	}

	if *conversationId == "" {
		fmt.Print("Missing --conversation-id argument.\n")
		flag.Usage()
		os.Exit(1)
	}

	twitterContext := twitter.InitTwitterContext(*conversationId, *authHeaderPath)
	dlManager, err := dlmanager.InitDLManager(*conversationId, twitterContext, dlmanager.Options{
		IsDebug:        *isDebug,
		DownloadVideos: *downloadVideos,
		DownloadPhotos: *downloadPhotos,
	})
	if err != nil {
		logger.MediaLogger.Fatalf("Failed to init DLManager %+v", dlManager)
	}
	dlManager.Start()
	logger.MediaLogger.Printf("Done\n")
}
