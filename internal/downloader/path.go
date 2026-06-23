package downloader

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"
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

	baseDir := opts.Directory
	if baseDir == "" {
		baseDir = "."
	}

	if opts.Mirror {
		baseDir = filepath.Join(baseDir, u.Host)
		return filepath.Join(baseDir, filepath.FromSlash(mirrorRelativePath(u)))
	}

	if baseDir == "." {
		return name
	}

	return filepath.Join(baseDir, name)
}

func mirrorRelativePath(u *url.URL) string {
	currentPath := u.Path
	if currentPath == "" || currentPath == "/" {
		return "index.html"
	}

	cleanPath := strings.TrimPrefix(path.Clean(currentPath), "/")
	if cleanPath == "." || cleanPath == "" {
		return "index.html"
	}

	if strings.HasSuffix(currentPath, "/") || path.Ext(cleanPath) == "" {
		return path.Join(cleanPath, "index.html")
	}

	return cleanPath
}
