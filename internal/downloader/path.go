package downloader

import (
	"fmt"
	"net/url"
	"os"
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
	baseDir = expandHomeDir(baseDir)

	if opts.Mirror {
		baseDir = filepath.Join(baseDir, u.Host)
		finalPath := filepath.Join(baseDir, filepath.FromSlash(mirrorRelativePath(u)))
		fmt.Printf("Saving to: %s\n\n", name)
		return finalPath
	}

	if baseDir == "." {
		fmt.Printf("Saving to: %s\n\n", name)
		return name
	}

	finalPath := filepath.Join(baseDir, name)
	fmt.Printf("Saving to: %s\n\n", name)
	return finalPath
}

func expandHomeDir(dir string) string {
	if dir == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	}
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(dir, "~/"))
		}
	}
	return dir
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
