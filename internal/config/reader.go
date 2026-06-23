package config

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
	"wget/internal/utils"

	"golang.org/x/time/rate"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴"}

type customReader struct {
	io.ReadCloser
	limiter        *rate.Limiter
	maxChunkBytes  int
	totalSizeBytes int64
	currentBytes   int64
	isUnknownSize  bool
	startedAt      time.Time
	spinnerIndex   int
	finishOnce     sync.Once
	shouldRender   bool
}

func (r *customReader) Read(p []byte) (int, error) {
	if r.maxChunkBytes > 0 && len(p) > r.maxChunkBytes {
		p = p[:r.maxChunkBytes]
	}

	if r.limiter != nil {
		if err := r.limiter.WaitN(context.Background(), len(p)); err != nil {
			return 0, err
		}
	}

	n, err := r.ReadCloser.Read(p)
	if n > 0 {
		r.currentBytes += int64(n)
		r.render(false)
	}
	if err == io.EOF {
		r.finish()
	}
	return n, err
}

func (r *customReader) Close() error {
	r.finish()
	return r.ReadCloser.Close()
}

func (r *customReader) finish() {
	r.finishOnce.Do(func() {
		r.render(true)
		if r.shouldRender {
			fmt.Fprintln(os.Stderr)
		}
	})
}

func (r *customReader) render(done bool) {
	if !r.shouldRender {
		return
	}
	if r.isUnknownSize {
		r.renderSpinner(done)
		return
	}
	r.renderBar(done)
}

func (r *customReader) renderBar(done bool) {
	const barWidth = 28
	total := r.totalSizeBytes
	if total <= 0 {
		total = r.currentBytes
	}

	percent := 0.0
	if total > 0 {
		percent = float64(r.currentBytes) / float64(total)
		if percent > 1 {
			percent = 1
		}
	}

	filled := int(percent * barWidth)
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	status := "Downloading"
	if done {
		status = "Done"
	}

	eta := "ETA --:--:--"
	if total > 0 && r.currentBytes > 0 && r.currentBytes < total {
		elapsed := time.Since(r.startedAt)
		if elapsed > 0 {
			bytesRemaining := total - r.currentBytes
			estimatedSeconds := float64(bytesRemaining) * float64(elapsed) / float64(r.currentBytes) / float64(time.Second)
			if estimatedSeconds > 0 {
				eta = fmt.Sprintf("ETA %s", utils.FormatDuration(time.Duration(estimatedSeconds)*time.Second))
			}
		}
	} else if done {
		eta = "ETA 00:00:00"
	}

	line := fmt.Sprintf("%s [%s] %6.2f%% (%s/%s) %s", status, bar, percent*100, utils.FormatBytes(r.currentBytes), utils.FormatBytes(total), eta)
	r.writeLine(line, done)
}

func (r *customReader) renderSpinner(done bool) {
	frame := spinnerFrames[r.spinnerIndex%len(spinnerFrames)]
	r.spinnerIndex++

	status := "Downloading"
	if done {
		status = "Done"
	}

	line := fmt.Sprintf("%s %s %s", status, frame, utils.FormatBytes(r.currentBytes))
	r.writeLine(line, done)
}

func (r *customReader) writeLine(line string, done bool) {
	if done {
		fmt.Fprintf(os.Stderr, "\r\x1b[2K%s\n", line)
		return
	}
	fmt.Fprintf(os.Stderr, "\r\x1b[2K%s", line)
}
