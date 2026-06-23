package mirror

import (
	"log/slog"
	"net/url"
	"wget/internal/config"

	"github.com/PuerkitoBio/goquery"
)

type Crawler struct {
	Client  *config.CustomHttpClient
	Root    *url.URL
	Visited map[string]bool
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

	c := NewCrawler(u, client)
	errs := c.Crawl(u)
	if errs != nil {
		slog.Error("failed to crawl website", "err", errs)
		return errs
	}

	return nil
}

func NewCrawler(root *url.URL, client *config.CustomHttpClient) *Crawler {
	return &Crawler{
		Client:  client,
		Root:    root,
		Visited: make(map[string]bool),
	}
}

func (c *Crawler) Crawl(u *url.URL) error {
	u, err := c.normalizeURL(u.String())
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
	doc, err := goquery.NewDocumentFromReader(res.Body)
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
		linkURL, err := c.normalizeURL(link)
		if err != nil {
			slog.Warn("Failed to normalize URL", "err", err, "link", link)
			continue
		}
		if err := c.Crawl(linkURL); err != nil {
			slog.Error("Failed to crawl URL", "err", err, "url", linkURL.String())
		}
	}

	return nil

}

func (c *Crawler) normalizeURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	return c.Root.ResolveReference(u), nil
}
func (c *Crawler) sameDomain(u *url.URL) bool {
	return u.Host == c.Root.Host
}
