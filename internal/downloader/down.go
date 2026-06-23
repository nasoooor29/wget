package downloader

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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

	_, err = io.Copy(out, resp.Body)
	return err
}
