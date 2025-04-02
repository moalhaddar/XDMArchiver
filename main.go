package main

import (
	"dmarchiver/dlmanager"
	"dmarchiver/logger"
	"dmarchiver/twitter"
	"flag"
)

func main() {
	var conversationId string
	var authHeaderPath string
	var isDebug bool
	flag.StringVar(&conversationId, "conversation-id", "", "ID for the conversation to be downloaded")
	flag.StringVar(&authHeaderPath, "auth-headers", "./auth.txt", "File path to authorization headers to be passed to each request\n"+
		"Headers are newline seperated, each header key value are colon seperated\n"+
		"Example file:\n\tCookie: ABCD\n\tContent-Type: application/json")
	flag.BoolVar(&isDebug, "debug", false, "Enable debugging mode")
	flag.Parse()
	if conversationId == "" {
		logger.MediaLogger.Fatal("Missing conversation-id argument. use --help flag to find out more.")
	}

	twitterContext := twitter.InitTwitterContext(conversationId, authHeaderPath)
	dlManager, err := dlmanager.InitDLManager(conversationId, twitterContext, isDebug)
	if err != nil {
		logger.MediaLogger.Fatalf("Failed to init DLManager %+v", dlManager)
	}
	dlManager.Start()
	logger.MediaLogger.Printf("Done\n")
}
