package config

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

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
	fileName       string
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
	const barWidth = 64
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

	filled := int(percent * float64(barWidth))
	if filled < 0 {
		filled = 0
	}
	if filled >= barWidth {
		filled = barWidth - 1
	}
	bar := strings.Repeat("=", filled) + ">" + strings.Repeat(" ", barWidth-filled-1)
	if done || percent >= 1 {
		bar = strings.Repeat("=", barWidth) + ">"
	}

	line := fmt.Sprintf("%-36s %3.0f%%[%s] %7s %9s/s    in %.1fs", r.fileName, percent*100, bar, formatProgressBytes(r.currentBytes), formatProgressSpeed(r.currentBytes, time.Since(r.startedAt)), time.Since(r.startedAt).Seconds())
	r.writeLine(line, done)
}

func (r *customReader) renderSpinner(done bool) {
	frame := spinnerFrames[r.spinnerIndex%len(spinnerFrames)]
	r.spinnerIndex++

	status := "Downloading"
	if done {
		status = "Done"
	}

	line := fmt.Sprintf("%s %s %s", status, frame, formatProgressBytes(r.currentBytes))
	r.writeLine(line, done)
}

func formatProgressBytes(value int64) string {
	if value < 1024 {
		return fmt.Sprintf("%d", value)
	}
	units := []string{"K", "M", "G", "T"}
	current := float64(value)
	for _, unit := range units {
		current /= 1024
		if current < 1024 || unit == units[len(units)-1] {
			return fmt.Sprintf("%.2f%s", current, unit)
		}
	}
	return fmt.Sprintf("%d", value)
}

func formatProgressSpeed(bytes int64, elapsed time.Duration) string {
	if elapsed <= 0 {
		return "0B"
	}
	bytesPerSecond := float64(bytes) / elapsed.Seconds()
	if bytesPerSecond < 1000 {
		return fmt.Sprintf("%.1fB", bytesPerSecond)
	}
	units := []string{"KB", "MB", "GB", "TB"}
	for _, unit := range units {
		bytesPerSecond /= 1000
		if bytesPerSecond < 1000 || unit == units[len(units)-1] {
			return fmt.Sprintf("%.1f%s", bytesPerSecond, unit)
		}
	}
	return fmt.Sprintf("%.1fB", bytesPerSecond)
}

func (r *customReader) writeLine(line string, done bool) {
	if done {
		fmt.Fprintf(os.Stderr, "\r\x1b[2K%s\n", line)
		return
	}
	fmt.Fprintf(os.Stderr, "\r\x1b[2K%s", line)
}
