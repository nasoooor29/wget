package config

import (
	"fmt"
	"net"
	"net/http"
	"path"
	"time"

	"golang.org/x/time/rate"
)

type RateLimitTransport struct {
	Base         http.RoundTripper
	BytesPerSec  int64
	ShouldRender bool
	OutputName   string
}

func (t *RateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("--%s--  %s\n", time.Now().Format("2006-01-02 15:04:05"), req.URL.String())
	host := req.URL.Hostname()
	port := req.URL.Port()
	if port == "" {
		if req.URL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	ips, lookupErr := net.LookupIP(host)
	if lookupErr == nil && len(ips) > 0 {
		fmt.Printf("Connecting to %s (%s)|%s|:%s... ", host, host, ips[0].String(), port)
	}

	resp, err := t.Base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if lookupErr == nil && len(ips) > 0 {
		fmt.Println("connected.")
	}
	fmt.Printf("HTTP request sent, awaiting response... %s\n", resp.Status)
	totalSize := resp.ContentLength
	isRedirect := resp.StatusCode >= 300 && resp.StatusCode < 400
	if isRedirect {
		location := resp.Header.Get("Location")
		if location != "" {
			fmt.Printf("Location: %s [following]\n", location)
		}
	}

	isUnkownSize := totalSize <= 0
	fileName := t.OutputName
	if fileName == "" {
		fileName = path.Base(resp.Request.URL.Path)
		if fileName == "." || fileName == "/" || fileName == "" {
			fileName = "index.html"
		}
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
		startedAt:      time.Now(),
		shouldRender:   t.ShouldRender && !isRedirect,
		fileName:       fileName,
	}

	return resp, nil
}
