# X DM Archiver

A CLI tool to archive direct messages from X (formerly Twitter), including all conversation history, media attachments (photos and videos), and metadata.

## Features

- Archives complete DM conversation history
- Downloads all photos and videos in original quality
- Preserves timestamps and message order
- Handles rate limiting automatically
- Saves message data in structured JSON format
- Media files are organized by conversation ID and type

## Installation

### Prerequisites

- Go 1.16 or later

### Building from source

```sh
git clone https://github.com/moalhaddar/XDMArchiver.git
cd XDMArchiver
go build
```

### Or download the binaries from the releases.

## Usage

```sh
./XDMArchiver --conversation-id "154687269-1223525587627004904" --download-photos --download-videos
```

## Parameters

```sh
Usage of XDMArchiver (version v1.0.0):
        XDMArchiver --conversation-id [--auth-headers FILE] [--download-videos] [--download-photos] [--debug]
  -auth-headers string
        File path to authorization headers to be passed to each request
        Headers are newline seperated, each header key value are colon seperated
        Example file:
                Cookie: ABCD
                Content-Type: application/json (default "./auth.txt")
  -conversation-id string
        ID for the conversation to be downloaded
  -debug
        Enable debugging mode
  -download-photos
        To download photos in the conversation
  -download-videos
        To download videos in the conversation
  -version
        Display version information
```

## Creating auth file

The auth file contains the headers needed to authenticate with X's API. Create a text file with the following format:

```
Cookie: your_cookie_value
Authorization: Bearer your_token_value
X-csrf-token: your_csrf_token
```

### How to get authentication values

1. Log in to your X (Twitter) account in a web browser
2. Open Developer Tools (F12 or right-click > Inspect)
3. Go to the Network tab
4. Navigate to your DMs on X
5. Look for API requests to `https://x.com/i/api/1.1/dm/conversation/` 
6. From the request headers, copy the values for:
   - Cookie
   - Authorization
   - X-csrf-token

Paste these into your auth.txt file in the format shown above.

## How It Works

XDMArchiver operates by fetching conversation data, processing each message, extracting the media urls from the conversation, downloading the media, saving the media and the messages to the file system. 

All data is saved to the local filesystem in the following structure:

```
conversations/
  {conversation_id}/
    events/
      {event_id}.json  # Raw message data
    photos/
      {timestamp}.jpg  # Photos from the conversation
    videos/
      {timestamp}-{bitrate}.mp4  # Videos from the conversation
```
## License

MIT

## Author

Mohammed Alhaddar