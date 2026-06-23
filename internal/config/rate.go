package config

import (
	"log/slog"
	"net/http"
	"time"
	"wget/internal/utils"

	"golang.org/x/time/rate"
)

type RateLimitTransport struct {
	Base         http.RoundTripper
	BytesPerSec  int64
	ShouldRender bool
}

func (t *RateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	slog.Info("sending request, awaiting response...", "url", req.URL.String())

	resp, err := base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	slog.Info("", "status code", resp.Status, "url", req.URL.String())
	totalSize := resp.ContentLength

	// content size: 56370 [~0.06MB]
	if totalSize > 0 {
		slog.Info("content size", "size", utils.FormatBytes(totalSize))
	} else {
		slog.Info("content size", "size", "unknown")
	}
	isUnkownSize := totalSize <= 0

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
		startedAt:      time.Now(),
		shouldRender:   t.ShouldRender,
	}

	return resp, nil
}
