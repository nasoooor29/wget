package config

import (
	"log/slog"
	"net/http"
	"time"
	"wget/internal/utils"
)

type CustomHttpClient struct {
	*http.Client
}

func NewHTTPClient(conf *Options) (*CustomHttpClient, error) {
	rateLimit, err := utils.ParseRateLimit(conf.RateLimit)
	if err != nil {
		slog.Error("", "err", err)
		return nil, err
	}
	if rateLimit > 0 {
		slog.Debug("Rate limit set", "bytes/sec", rateLimit)
	}
	if conf.Timeout <= 0 {
		slog.Warn("Timeout is not set or invalid, using default 30 seconds")
		conf.Timeout = 30 // default timeout in seconds
	}
	slog.Debug("Timeout set", "seconds", conf.Timeout)

	return &CustomHttpClient{
		Client: &http.Client{
			Timeout:   time.Duration(conf.Timeout) * time.Second, // important: large downloads should not timeout
			Transport: NewRateLimitTransport(rateLimit),
		},
	}, nil
}
