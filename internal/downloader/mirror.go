package downloader

import (
	"bytes"
	"io"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"
	"wget/internal/config"

	"github.com/PuerkitoBio/goquery"
)

type Crawler struct {
	Client  *config.CustomHttpClient
	Root    *url.URL
	Seed    string
	Visited map[string]bool
	Opts    *config.Options
}

func MirrorWebsite(opts *config.Options) error {
	u, err := url.Parse(opts.URL)
	if err != nil {
		slog.Error("failed to parse URL", "err", err, "url", opts.URL)
		return err
	}
	u.Path = "/"

	client, err := config.NewHTTPClient(opts)
	if err != nil {
		slog.Error("failed to create HTTP client", "err", err)
		return err
	}

	c := NewCrawler(u, client, opts)
	errs := c.Crawl(u)
	if errs != nil {
		slog.Error("failed to crawl website", "err", errs)
		return errs
	}

	return nil
}

func NewCrawler(root *url.URL, client *config.CustomHttpClient, opts *config.Options) *Crawler {
	return &Crawler{
		Client:  client,
		Root:    root,
		Seed:    root.String(),
		Visited: make(map[string]bool),
		Opts:    opts,
	}
}

func (c *Crawler) Crawl(u *url.URL) error {
	u, err := c.normalizeURL(c.Root, u.String())
	if err != nil {
		return err
	}

	key := u.String()
	if c.Visited[key] {
		return nil
	}
	c.Visited[key] = true

	if !c.sameDomain(u) {
		return nil
	}
	if key != c.Seed && c.shouldSkipURL(u) {
		slog.Info("Skipping mirrored URL", "url", u.String())
		return nil
	}

	slog.Info("Crawling URL", "url", u.String())
	res, err := c.Client.Get(u.String())
	if err != nil {
		slog.Error("Could not fetch URL", "err", err, "url", u.String())
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		slog.Warn("Non-OK HTTP status", "status", res.Status, "url", u.String())
		return nil
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	isHTML := isHTMLResponse(res.Header.Get("Content-Type"), res.Request.URL.Path)
	if err := saveMirroredResponse(c.Opts, res.Request.URL, body, c, isHTML); err != nil {
		return err
	}

	if !isHTML {
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	var links []string

	doc.Find("a[href], link[href]").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if ok {
			links = append(links, href)
		}
	})

	doc.Find("img[src], script[src]").Each(func(_ int, s *goquery.Selection) {
		src, ok := s.Attr("src")
		if ok {
			links = append(links, src)
		}
	})

	for _, link := range links {
		linkURL, err := c.normalizeURL(u, link)
		if err != nil {
			slog.Warn("Failed to normalize URL", "err", err, "link", link)
			continue
		}
		if c.shouldSkipURL(linkURL) {
			continue
		}
		if err := c.Crawl(linkURL); err != nil {
			slog.Error("Failed to crawl URL", "err", err, "url", linkURL.String())
		}
	}

	return nil

}
func isHTMLResponse(contentType string, currentPath string) bool {
	if strings.Contains(strings.ToLower(contentType), "text/html") {
		return true
	}
	return currentPath == "" || currentPath == "/" || strings.HasSuffix(currentPath, ".html") || strings.HasSuffix(currentPath, ".htm")
}

func (c *Crawler) normalizeURL(base *url.URL, raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	u.Fragment = ""

	return base.ResolveReference(u), nil
}
func (c *Crawler) sameDomain(u *url.URL) bool {
	return u.Host == c.Root.Host
}

func (c *Crawler) shouldSkipURL(u *url.URL) bool {
	if c == nil || c.Opts == nil {
		return false
	}
	if matchesPathPrefixes(u.Path, c.Opts.Exclude) {
		return true
	}
	return matchesFileSuffixes(u.Path, c.Opts.Reject)
}

func (c *Crawler) convertMirroredLinks(currentURL *url.URL, body []byte) ([]byte, error) {
	if c == nil || c.Opts == nil || !c.Opts.ConvertLinks {
		return body, nil
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	currentPath := resolveOutputPath(c.Opts, currentURL)
	currentDir := filepath.Dir(currentPath)

	convertAttr := func(selection *goquery.Selection, attr string) {
		value, ok := selection.Attr(attr)
		if !ok || strings.TrimSpace(value) == "" {
			return
		}

		linkURL, err := url.Parse(value)
		if err != nil {
			return
		}
		resolved := currentURL.ResolveReference(linkURL)
		fragment := resolved.Fragment
		resolved.Fragment = ""
		if !c.sameDomain(resolved) || c.shouldSkipURL(resolved) {
			return
		}

		targetPath := resolveOutputPath(c.Opts, resolved)
		relativePath, err := filepath.Rel(currentDir, targetPath)
		if err != nil {
			return
		}

		if fragment != "" {
			relativePath += "#" + fragment
		}

		selection.SetAttr(attr, filepath.ToSlash(relativePath))
	}

	doc.Find("a[href], link[href]").Each(func(_ int, s *goquery.Selection) {
		convertAttr(s, "href")
	})

	doc.Find("img[src], script[src]").Each(func(_ int, s *goquery.Selection) {
		convertAttr(s, "src")
	})

	html, err := doc.Html()
	if err != nil {
		return nil, err
	}

	return []byte(html), nil
}
