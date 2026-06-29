package downloader

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"wget/internal/config"
)

func DownloadFromFile(opts *config.Options) error {
	content, err := os.ReadFile(opts.InputFile)
	if err != nil {
		return err
	}

	validInputs := []string{}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		uu, err := url.Parse(line)
		if err != nil || uu.Scheme == "" || uu.Host == "" {
			fmt.Println("invalid URL in input file", "url", line)
			continue
		}

		validInputs = append(validInputs, line)
	}

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	runErrs := []error{}

	for _, line := range validInputs {
		line := line
		wg.Add(1)
		go func() {
			defer wg.Done()

			child := *opts
			child.URL = line
			if err := DownloadOne(&child); err != nil {
				fmt.Println("failed to download URL from input file", "url", line, "err", err)
				mu.Lock()
				runErrs = append(runErrs, err)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(runErrs) == 0 {
		return nil
	}
	return errors.Join(runErrs...)
}
