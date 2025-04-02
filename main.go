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
	version = "v1.0.0"
)

func main() {
	var conversationId string
	var authHeaderPath string
	var isDebug bool
	var showVersion bool
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s (version %s):\n", os.Args[0], version)
		flag.PrintDefaults()
	}
	flag.BoolVar(&showVersion, "version", false, "Display version information")
	flag.StringVar(&conversationId, "conversation-id", "", "ID for the conversation to be downloaded")
	flag.StringVar(&authHeaderPath, "auth-headers", "./auth.txt", "File path to authorization headers to be passed to each request\n"+
		"Headers are newline seperated, each header key value are colon seperated\n"+
		"Example file:\n\tCookie: ABCD\n\tContent-Type: application/json")
	flag.BoolVar(&isDebug, "debug", false, "Enable debugging mode")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s version %s\n", os.Args[0], version)
		os.Exit(0)
	}

	if conversationId == "" {
		logger.MediaLogger.Fatal("Missing --conversation-id argument. use --help flag to find out more.")
	}

	twitterContext := twitter.InitTwitterContext(conversationId, authHeaderPath)
	dlManager, err := dlmanager.InitDLManager(conversationId, twitterContext, isDebug)
	if err != nil {
		logger.MediaLogger.Fatalf("Failed to init DLManager %+v", dlManager)
	}
	dlManager.Start()
	logger.MediaLogger.Printf("Done\n")
}
