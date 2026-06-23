package config

import (
	"log/slog"
	"net/http"

	"golang.org/x/time/rate"
)

type RateLimitTransport struct {
	Base        http.RoundTripper
	BytesPerSec int64
}

func NewRateLimitTransport(bytesPerSec int64) *RateLimitTransport {
	return &RateLimitTransport{
		BytesPerSec: bytesPerSec,
	}
}

func (t *RateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	resp, err := base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	totalSize := resp.ContentLength
	isUnkownSize := totalSize <= 0
	if totalSize <= 0 {
		slog.Debug("Content length is unknown, will use spinner for progress indication")
	}

	var limiter *rate.Limiter
	if t.BytesPerSec > 0 {
		burst := int(t.BytesPerSec)
		if burst < 1 {
			burst = 1
		}
		limiter = rate.NewLimiter(rate.Limit(t.BytesPerSec), burst)
	}

	resp.Body = &customReader{
		ReadCloser:     resp.Body,
		limiter:        limiter,
		maxChunkBytes:  int(t.BytesPerSec),
		totalSizeBytes: totalSize,
		currentBytes:   0,
		isUnknownSize:  isUnkownSize,
	}

	return resp, nil
}
