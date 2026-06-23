package downloader

import (
	"log/slog"
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
		finalPath := filepath.Join(baseDir, filepath.FromSlash(mirrorRelativePath(u)))
		slog.Info("saving file to:", "file_name", finalPath)
		return finalPath
	}

	if baseDir == "." {
		slog.Info("saving file to:", "file_name", name)
		return name
	}

	finalPath := filepath.Join(baseDir, name)
	slog.Info("saving file to:", "file_name", finalPath)
	return finalPath
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

func matchesFileSuffixes(currentPath string, suffixes []string) bool {
	if len(suffixes) == 0 {
		return false
	}

	currentExt := strings.TrimPrefix(strings.ToLower(path.Ext(currentPath)), ".")
	if currentExt == "" {
		return false
	}

	for _, suffix := range suffixes {
		normalized := strings.TrimSpace(strings.ToLower(strings.TrimPrefix(suffix, ".")))
		if normalized != "" && currentExt == normalized {
			return true
		}
	}

	return false
}

func matchesPathPrefixes(currentPath string, prefixes []string) bool {
	if len(prefixes) == 0 {
		return false
	}

	currentPath = normalizeMirrorPath(currentPath)
	for _, prefix := range prefixes {
		normalized := normalizeMirrorPath(prefix)
		if normalized == "" || normalized == "/" {
			continue
		}
		if currentPath == normalized || strings.HasPrefix(currentPath, normalized+"/") {
			return true
		}
	}

	return false
}

func normalizeMirrorPath(value string) string {
	if value == "" {
		return "/"
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	cleaned := path.Clean(value)
	if cleaned == "." {
		return "/"
	}
	return cleaned
}
