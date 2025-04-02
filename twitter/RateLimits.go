package twitter

import (
	"XDMArchiver/logger"
	"XDMArchiver/utils"
	"net/http"
	"time"
)

type RateLimits struct {
	RateLimit          string
	RateLimitRemaining string
	RateLimitResetTime *time.Time
}

const (
	RATE_LIMIT_RESET_TIME = "X-Rate-Limit-Reset"
	RATE_LIMIT_REMAINING  = "X-Rate-Limit-Remaining"
	RATE_LIMIT            = "X-Rate-Limit-Limit"
)

func RateLimit(headers http.Header) *RateLimits {
	var r RateLimits
	r.RateLimit = headers.Get(RATE_LIMIT)
	r.RateLimitRemaining = headers.Get(RATE_LIMIT_REMAINING)
	resetTime, err := utils.UnixTimestampStringToTime(headers.Get(RATE_LIMIT_RESET_TIME), false)
	if err != nil {
		r.RateLimitResetTime = nil
	}
	r.RateLimitResetTime = resetTime

	return &r
}

func (r *RateLimits) Print(indent string) {
	if r.RateLimit != "" {
		logger.EventsLogger.Printf("%sRate Limit %s\n", indent, r.RateLimit)
	}
	if r.RateLimitRemaining != "" {
		logger.EventsLogger.Printf("%sRemaining Rate Limit %s\n", indent, r.RateLimitRemaining)
	}
	if r.RateLimitResetTime != nil {
		logger.EventsLogger.Printf("%sRate Limit Reset %s", indent, r.RateLimitResetTime.Local().Format(time.DateTime))
	}
}
