package downloader

import (
	"net/url"
	"path/filepath"
	"wget/internal/config"
)

func resolveOutputPath(opts *config.Options, u *url.URL) string {
	name := opts.Output
	if name == "" {
		name = filepath.Base(u.Path)
		if name == "." || name == "/" || name == "" {
			name = "index.html"
		}
	}

	if opts.Directory == "" || opts.Directory == "." {
		return name
	}

	return filepath.Join(opts.Directory, name)
}
