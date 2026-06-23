package downloader

import (
	"errors"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"wget/internal/config"
)

func DownloadFromFile(opts *config.Options) error {
	accumlatedErrors := []error{}
	content, err := os.ReadFile(opts.InputFile)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		uu, err := url.Parse(line)
		if err != nil || uu.Scheme == "" || uu.Host == "" {
			slog.Warn("invalid URL in input file", "url", line)
			continue
		}

		child := *opts
		child.URL = line
		if err := DownloadOne(&child); err != nil {
			slog.Error("failed to download URL from input file", "url", line, "err", err)
			accumlatedErrors = append(accumlatedErrors, err)
		}
	}

	if len(accumlatedErrors) == 0 {
		return nil
	}
	return errors.Join(accumlatedErrors...)
}
