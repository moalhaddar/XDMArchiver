package twitter

import (
	"XDMArchiver/logger"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	CONVERSATION_API_BASE_PATH = "https://x.com/i/api/1.1/dm/conversation/"
	MAX_ID_QUERY_PARAM         = "max_id"
)

type TwitterContext struct {
	conversationId string
	authHeaders    map[string]string
}

func InitTwitterContext(conversationId string, authHeaderPath string) TwitterContext {
	var context TwitterContext
	context.conversationId = conversationId
	context.authHeaders = make(map[string]string)
	context.loadAuthHeadersFromFile(authHeaderPath)
	return context
}

func deserializeEvent(response []byte) (*ConversationResponse, error) {
	var responseJson ConversationResponse
	reader := bytes.NewReader(response)
	err := json.NewDecoder(reader).Decode(&responseJson)

	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return &responseJson, nil
}

func (context *TwitterContext) loadAuthHeadersFromFile(authFilePath string) {
	file, err := os.Open(authFilePath)
	if err != nil {
		logger.MediaLogger.Fatal("Failed to open auth file.")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			logger.MediaLogger.Fatalf("Failed to parse header line:\n\t%s\n", line)
		}
		context.authHeaders[parts[0]] = parts[1]
	}
}

func (context *TwitterContext) GetConversation(maxId *string) (*ConversationResponse, *RateLimits, error) {
	req, err := http.NewRequest(http.MethodGet, CONVERSATION_API_BASE_PATH+context.conversationId+".json", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	query := req.URL.Query()

	if maxId != nil {
		query.Add(MAX_ID_QUERY_PARAM, *maxId)
	}

	query.Add("context", "FETCH_DM_CONVERSATION_HISTORY")
	query.Add("include_profile_interstitial_type", "1")
	query.Add("include_blocking", "1")
	query.Add("include_blocked_by", "1")
	query.Add("include_followed_by", "1")
	query.Add("include_want_retweets", "1")
	query.Add("include_mute_edge", "1")
	query.Add("include_can_dm", "1")
	query.Add("include_can_media_tag", "1")
	query.Add("include_ext_is_blue_verified", "1")
	query.Add("include_ext_verified_type", "1")
	query.Add("include_ext_profile_image_shape", "1")
	query.Add("skip_status", "1")
	query.Add("dm_secret_conversations_enabled", "false")
	query.Add("krs_registration_enabled", "true")
	query.Add("cards_platform", "Web-12")
	query.Add("include_cards", "1")
	query.Add("include_ext_alt_text", "true")
	query.Add("include_ext_limited_action_results", "true")
	query.Add("include_quote_count", "true")
	query.Add("include_reply_count", "1")
	query.Add("tweet_mode", "extended")
	query.Add("include_ext_views", "true")
	query.Add("dm_users", "false")
	query.Add("include_groups", "true")
	query.Add("include_inbox_timelines", "true")
	query.Add("include_ext_media_color", "true")
	query.Add("supports_reactions", "true")
	query.Add("supports_edit", "true")
	query.Add("include_conversation_info", "true")
	query.Add("ext", "mediaColor,altText,mediaStats,highlightedLabel,parodyCommentaryFanLabel,voiceInfo,birdwatchPivot,superFollowMetadata,unmentionInfo,editControl,article")
	req.URL.RawQuery = query.Encode()

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:127.0) Gecko/20100101 Firefox/127.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-US")
	req.Header.Add("x-twitter-auth-type", "OAuth2Session")
	req.Header.Add("x-twitter-client-language", "en")
	req.Header.Add("x-twitter-active-user", "yes")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("TE", "trailers")
	for key, value := range context.authHeaders {
		req.Header.Add(key, value)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	response, err := client.Do(req)

	if err != nil {
		return nil, nil, fmt.Errorf("failed making request %w", err)
	}

	rateLimits := RateLimit(response.Header)

	if response.StatusCode != 200 {
		return nil, rateLimits, fmt.Errorf("conversation api request failed: %w", &ErrNot200{
			StatusCode: response.StatusCode,
		})
	}

	defer response.Body.Close()
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, rateLimits, fmt.Errorf("failed to read http response body %w", err)
	}
	event, err := deserializeEvent(bodyBytes)
	if err != nil {
		bodyString := string(bodyBytes)
		return nil, rateLimits, fmt.Errorf("failed to deserialize event from response body\n response body: %s. %w", bodyString, err)
	}
	return event, rateLimits, nil
}

func (context *TwitterContext) GetFile(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request %w", err)
	}

	for key, value := range context.authHeaders {
		req.Header.Add(key, value)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed making request %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("download request failed: %w", &ErrNot200{
			StatusCode: response.StatusCode,
		})
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read http body from response %w", err)
	}

	return bytes, nil
}
