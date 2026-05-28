package content

import (
	"strings"
	"time"
)

type Page struct {
	Title       string
	Date        time.Time
	LastMod     time.Time
	Draft       bool
	Categories  []string
	Tags        []string
	Slug        string
	URL         string
	Weight      int
	Description string
	Summary     string
	TOC         bool
	Math        bool

	Content     string
	RawContent  string
	WordCount   int
	ReadingTime int
	IsPage      bool

	FilePath string
	RelPath  string
	Section  string
}

func (p *Page) CalcReadingTime() {
	if p.WordCount == 0 {
		p.WordCount = len([]rune(p.RawContent)) / 3
	}
	p.ReadingTime = p.WordCount / 400
	if p.ReadingTime < 1 {
		p.ReadingTime = 1
	}
}

func (p *Page) Permalink() string {
	if p.URL != "" {
		return p.URL
	}
	if p.Slug != "" {
		return "/" + p.Section + "/" + p.Slug + "/"
	}
	rel := strings.TrimSuffix(p.RelPath, ".md")
	rel = strings.TrimSuffix(rel, "/index")
	return "/" + rel + "/"
}

func (p *Page) CategoryString() string { return strings.Join(p.Categories, "/") }
func (p *Page) HasCategory() bool       { return len(p.Categories) > 0 }
