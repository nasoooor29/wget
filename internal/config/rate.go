package config

import (
	"context"
	"io"
	"golang.org/x/time/rate"
)

type RateLimitedReader struct {
	io.ReadCloser
	limiter *rate.Limiter
}

func (r *RateLimitedReader) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if n > 0 {
		if err := r.limiter.WaitN(context.Background(), n); err != nil {
			return n, err
		}
	}
	return n, err
}
