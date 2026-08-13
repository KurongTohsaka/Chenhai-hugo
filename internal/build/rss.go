package build

import (
	"encoding/xml"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
)

// Atom feed structures (RFC 4287).

type atomFeed struct {
	XMLName  xml.Name    `xml:"feed"`
	Xmlns    string      `xml:"xmlns,attr"`
	Title    string      `xml:"title"`
	Subtitle string      `xml:"subtitle,omitempty"`
	ID       string      `xml:"id"`
	Updated  string      `xml:"updated"`
	Links    []atomLink  `xml:"link"`
	Author   *atomAuthor `xml:"author,omitempty"`
	Entries  []atomEntry `xml:"entry"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
	Href string `xml:"href,attr"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

type atomEntry struct {
	Title   string     `xml:"title"`
	ID      string     `xml:"id"`
	Updated string     `xml:"updated"`
	Links   []atomLink `xml:"link"`
	Summary string     `xml:"summary,omitempty"`
}

// generateRSS builds the Atom feed XML for published pages (newest first,
// capped at cfg.RSS.Limit). Returns empty bytes when BaseURL is empty or RSS
// is disabled.
func generateRSS(cfg *config.Config, pages []*content.Page) ([]byte, error) {
	if !cfg.RSS.Enabled || cfg.BaseURL == "" {
		return nil, nil
	}
	// Y12: limit 边界校验——非法值（<1）直接报错，避免负值裸 panic
	// （slice bounds out of range）与零值静默空 feed。
	if cfg.RSS.Limit < 1 {
		return nil, fmt.Errorf("rss.limit 必须 ≥ 1（当前 %d）", cfg.RSS.Limit)
	}
	base := strings.TrimRight(cfg.BaseURL, "/")

	var published []*content.Page
	for _, p := range pages {
		if !p.Draft {
			published = append(published, p)
		}
	}
	sort.Slice(published, func(i, j int) bool {
		return published[i].Date.After(published[j].Date)
	})
	if len(published) > cfg.RSS.Limit {
		published = published[:cfg.RSS.Limit]
	}

	feed := atomFeed{
		Xmlns:    "http://www.w3.org/2005/Atom",
		Title:    cfg.Title,
		Subtitle: cfg.Description,
		ID:       base + "/",
		Updated:  time.Now().UTC().Format(time.RFC3339),
		Links: []atomLink{
			{Rel: "self", Type: "application/atom+xml", Href: base + "/atom.xml"},
			{Href: base + "/"},
		},
	}
	if cfg.Author.Name != "" {
		feed.Author = &atomAuthor{Name: cfg.Author.Name}
	}
	for _, p := range published {
		updated := p.LastMod
		if updated.IsZero() {
			updated = p.Date
		}
		entry := atomEntry{
			Title:   cleanXMLText(p.Title),
			ID:      base + p.Permalink(),
			Updated: updated.UTC().Format(time.RFC3339),
			Links:   []atomLink{{Href: base + p.Permalink()}},
		}
		// Y2: v0.8 统一摘要——增量构建下旧文章无全文缓存（collect.go 置空
		// Content），全文/摘要混排会致 feed 前后不一致；Description 优先，
		// 无则正文截断约 300 字。
		summary := p.Description
		if summary == "" {
			summary = summarizeContent(p.Content, 300)
		}
		entry.Summary = cleanXMLText(summary)
		feed.Entries = append(feed.Entries, entry)
	}

	out, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal atom feed: %w", err)
	}
	return append([]byte(xml.Header), out...), nil
}

// htmlTagRegex matches HTML tags for summary extraction (compiled once,
// not per page per build).
var htmlTagRegex = regexp.MustCompile(`<[^>]+>`)

// summarizeContent strips HTML tags and truncates plain text to n runes.
func summarizeContent(content string, n int) string {
	text := strings.TrimSpace(htmlTagRegex.ReplaceAllString(content, ""))
	text = html.UnescapeString(text)
	runes := []rune(text)
	if len(runes) > n {
		return string(runes[:n]) + "…"
	}
	return text
}

// cleanXMLText removes XML 1.0 illegal control characters (\x00-\x08, \x0B,
// \x0C, \x0E-\x1F), keeping \t \n \r. Prevents a single bad character
// (common when pasting from Word/web pages) from failing the whole build.
func cleanXMLText(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' || r >= 0x20 {
			return r
		}
		return -1
	}, s)
}
