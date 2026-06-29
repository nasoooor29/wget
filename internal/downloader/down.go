package downloader

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"wget/internal/config"
)

func DownloadOne(opts *config.Options) error {
	client, err := config.NewHTTPClient(opts)
	if err != nil {
		return err
	}

	resp, err := client.Get(opts.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		slog.Error("download failed", "status", resp.Status, "url", opts.URL)
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	targetPath := resolveOutputPath(opts, resp.Request.URL)

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		slog.Error("failed to create directories", "err", err, "path", targetPath)
		return err
	}

	out, err := os.Create(targetPath)
	if err != nil {
		slog.Error("failed to create output file", "err", err, "path", targetPath)
		return err
	}
	defer out.Close()

	startedAt := time.Now()
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		slog.Error("failed to write response body to file", "err", err, "path", targetPath)
		return err
	}
	totalSize := resp.ContentLength
	if totalSize <= 0 {
		totalSize = written
	}
	fmt.Printf("%s (%s/s) - ‘%s’ saved [%d/%d]\n", time.Now().Format("2006-01-02 15:04:05"), formatDownloadSpeed(written, time.Since(startedAt)), filepath.Base(targetPath), written, totalSize)
	return err
}

func formatDownloadSpeed(bytes int64, elapsed time.Duration) string {
	if elapsed <= 0 {
		return "0 B"
	}
	bytesPerSecond := float64(bytes) / elapsed.Seconds()
	if bytesPerSecond < 1000 {
		return fmt.Sprintf("%.1f B", bytesPerSecond)
	}
	units := []string{"KB", "MB", "GB", "TB"}
	for _, unit := range units {
		bytesPerSecond /= 1000
		if bytesPerSecond < 1000 || unit == units[len(units)-1] {
			return fmt.Sprintf("%.1f %s", bytesPerSecond, unit)
		}
	}
	return fmt.Sprintf("%.1f B", bytesPerSecond)
}

func saveMirroredResponse(opts *config.Options, targetURL *url.URL, body []byte, crawler *Crawler, isHTML bool) error {
	if crawler != nil && isHTML {
		convertedBody, err := crawler.convertMirroredLinks(targetURL, body)
		if err != nil {
			return err
		}
		body = convertedBody
	}

	targetPath := resolveOutputPath(opts, targetURL)
	slog.Debug("resolved output path", "path", targetPath)

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		slog.Error("failed to create directories", "err", err, "path", targetPath)
		return err
	}

	out, err := os.Create(targetPath)
	if err != nil {
		slog.Error("failed to create output file", "err", err, "path", targetPath)
		return err
	}
	defer out.Close()

	_, err = out.Write(body)
	return err
}
