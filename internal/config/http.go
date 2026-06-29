package config

import (
	"fmt"
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
		fmt.Println("err", err)
		return nil, err
	}
	if conf.Timeout <= 0 {
		fmt.Println("Timeout is not set or invalid, using default 30 seconds")
		conf.Timeout = 30 // default timeout in seconds
	}

	return &CustomHttpClient{
		Client: &http.Client{
			Timeout: time.Duration(conf.Timeout) * time.Second, // important: large downloads should not timeout
			Transport: &RateLimitTransport{
				Base: &http.Transport{
					DisableCompression: true,
				},
				BytesPerSec:  rateLimit,
				ShouldRender: conf.ShouldRender,
				OutputName:   conf.Output,
			},
		},
	}, nil
}
