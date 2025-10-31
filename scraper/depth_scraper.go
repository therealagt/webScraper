package scraper

import (
	"bytes"
	"context"
	"database/sql"

	"github.com/PuerkitoBio/goquery"
)

func (s *Scraper) ScrapeWithDepth(ctx context.Context, startURL string, maxDepth int) {
	visited := make(map[string]bool)
	s.scrapeRecursive(ctx, startURL, 1, maxDepth, visited)
}

func (s *Scraper) scrapeRecursive(ctx context.Context, urlStr string, depth, maxDepth int, visited map[string]bool) {
	if depth > maxDepth || visited[urlStr] {
		return 
	}
	visited[urlStr] = true

	s.fetchPage(ctx, urlStr, 1, 0, sql.NullTime{Valid: false})

	links := extractLinksFromHTML(s.DB, urlStr)
	for _, link := range links {
		s.scrapeRecursive(ctx, link, depth+1, maxDepth, visited)
	}
}

func extractLinksFromHTML(db *sql.DB, urlStr string) []string {
    var html []byte
    err := db.QueryRow("SELECT html FROM raw_html WHERE url = $1", urlStr).Scan(&html)
    if err != nil {
        return nil
    }
    doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
    if err != nil {
        return nil
    }
    var links []string
    doc.Find("a").Each(func(i int, s *goquery.Selection) {
        if href, exists := s.Attr("href"); exists {
            links = append(links, href)
        }
    })
    return links
}
