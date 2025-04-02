# XDMArchiver

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
git clone https://github.com/username/xdmarchiver.git
cd xdmarchiver
go build
```

### Or download the binaries from the releases.

## Usage

```sh
./XDMArchiver \
  --auth-headers ./auth.txt \
  --conversation-id "6322615781-2112126499297904" \
  --debug
```

### Parameters

- `--conversation-id` (required): The ID of the conversation to archive. You can find this in the URL when viewing DMs on X.com.
- `--auth-headers` (default: `./auth.txt`): Path to a file containing authorization headers.
- `--debug` (optional): Enable debug mode for additional logging and step-by-step execution.

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