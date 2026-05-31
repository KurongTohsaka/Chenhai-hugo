package index

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
)

// Site holds the complete site data used during rendering.
type Site struct {
	Config     *config.Config
	Pages      []*content.Page
	Categories map[string][]*content.Page // key = category path like "技术/Go"
	Tags       map[string][]*content.Page // key = tag name
	Series     map[string][]*content.Page
	Archives   *Archive
}

// ArchiveMonth holds pages for a specific month.
type ArchiveMonth struct {
	Month time.Month
	Pages []*content.Page
}

// Archive holds pages grouped by year then month (months sorted 12→1).
type Archive struct {
	Years []int
	Items map[int][]ArchiveMonth
}

// SearchEntry is a single entry in the search-index.json.
type SearchEntry struct {
	Title    string   `json:"title"`
	URL      string   `json:"url"`
	Content  string   `json:"content"`
	Summary  string   `json:"summary"`
	Tags     []string `json:"tags"`
	Category string   `json:"category"`
	Date     string   `json:"date"`
}

// TagCloudEntry holds computed tag cloud data.
type TagCloudEntry struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Size  string `json:"size"` // xs|sm|md|lg|xl
}

// BuildSite processes all pages and builds the complete Site index.
// Pages are sorted by date descending. Draft pages are excluded from
// Categories, Tags, and Archives indices.
func BuildSite(cfg *config.Config, pages []*content.Page) *Site {
	// Sort by date descending
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Date.After(pages[j].Date)
	})

	site := &Site{
		Config:     cfg,
		Pages:      pages,
		Categories: make(map[string][]*content.Page),
		Tags:       make(map[string][]*content.Page),
		Series:     make(map[string][]*content.Page),
		Archives: &Archive{
			Items: make(map[int][]ArchiveMonth),
		},
	}

	for _, page := range pages {
		if page.Draft || page.Layout != "" {
			continue
		}

		// Categories
		if page.HasCategory() {
			cat := page.CategoryString()
			site.Categories[cat] = append(site.Categories[cat], page)
		}

		// Tags
		for _, tag := range page.Tags {
			site.Tags[tag] = append(site.Tags[tag], page)
		}

		// Archives
		year := page.Date.Year()
		month := page.Date.Month()
		if _, ok := site.Archives.Items[year]; !ok {
			site.Archives.Years = append(site.Archives.Years, year)
		}
		site.Archives.Items[year] = appendArchiveMonth(site.Archives.Items[year], month, page)

		// Series
		if page.Series != "" {
			site.Series[page.Series] = append(site.Series[page.Series], page)
		}
	}

	// Sort years descending
	sort.Slice(site.Archives.Years, func(i, j int) bool {
		return site.Archives.Years[i] > site.Archives.Years[j]
	})

	// Sort months within each year descending (12→1)
	for _, months := range site.Archives.Items {
		sort.Slice(months, func(i, j int) bool {
			return months[i].Month > months[j].Month
		})
	}

	// Sort each series by date ascending
	for _, pages := range site.Series {
		sort.Slice(pages, func(i, j int) bool {
			return pages[i].Date.Before(pages[j].Date)
		})
	}

	return site
}

// BuildSearchIndex generates the search JSON from all non-draft pages.
// Content is truncated to 500 runes.
func BuildSearchIndex(pages []*content.Page) ([]byte, error) {
	var entries []SearchEntry
	for _, page := range pages {
		if page.Draft || page.Layout != "" {
			continue
		}
		preview := page.RawContent
		runes := []rune(preview)
		if len(runes) > 500 {
			preview = string(runes[:500]) + "..."
		}
		entries = append(entries, SearchEntry{
			Title:    page.Title,
			URL:      page.Permalink(),
			Content:  preview,
			Summary:  page.Summary,
			Tags:     page.Tags,
			Category: page.CategoryString(),
			Date:     page.Date.Format("2006-01-02"),
		})
	}
	return json.MarshalIndent(entries, "", "  ")
}

// BuildTagCloud computes tag cloud data with 5 size levels
// based on min/max count distribution.
func (s *Site) BuildTagCloud() []TagCloudEntry {
	if len(s.Tags) == 0 {
		return nil
	}

	minCount, maxCount := -1, 0
	for _, pages := range s.Tags {
		n := len(pages)
		if minCount == -1 || n < minCount {
			minCount = n
		}
		if n > maxCount {
			maxCount = n
		}
	}

	sizes := []string{"xs", "sm", "md", "lg", "xl"}
	var entries []TagCloudEntry
	for tag, pages := range s.Tags {
		idx := 0
		if maxCount > minCount {
			idx = (len(pages) - minCount) * (len(sizes) - 1) / (maxCount - minCount)
		}
		entries = append(entries, TagCloudEntry{
			Name:  tag,
			Count: len(pages),
			Size:  sizes[idx],
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	return entries
}

// PublishedPages returns all non-draft pages.
func (s *Site) PublishedPages() []*content.Page {
	var result []*content.Page
	for _, page := range s.Pages {
		if !page.Draft && page.Layout == "" {
			result = append(result, page)
		}
	}
	return result
}

// PagesByCategory returns pages filtered by category path.
func (s *Site) PagesByCategory(category string) []*content.Page {
	return s.Categories[category]
}

// PagesByTag returns pages filtered by tag.
func (s *Site) PagesByTag(tag string) []*content.Page {
	return s.Tags[tag]
}

// RelatedPosts returns up to n most related non-draft posts for the given page,
// based on tag overlap (Jaccard similarity).
func (s *Site) RelatedPosts(page *content.Page, n int) []*content.Page {
	if len(page.Tags) == 0 {
		return nil
	}
	type scored struct {
		p     *content.Page
		score float64
	}
	var candidates []scored
	pageTagSet := make(map[string]bool)
	for _, t := range page.Tags {
		pageTagSet[t] = true
	}
	for _, p := range s.PublishedPages() {
		if p.FilePath == page.FilePath {
			continue
		}
		intersection := 0
		for _, t := range p.Tags {
			if pageTagSet[t] {
				intersection++
			}
		}
		if intersection == 0 {
			continue
		}
		union := len(pageTagSet)
		for _, t := range p.Tags {
			if !pageTagSet[t] {
				union++
			}
		}
		score := float64(intersection) / float64(union)
		candidates = append(candidates, scored{p, score})
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	if len(candidates) > n {
		candidates = candidates[:n]
	}
	result := make([]*content.Page, len(candidates))
	for i, c := range candidates {
		result[i] = c.p
	}
	return result
}

// appendArchiveMonth appends a page to the correct month slice.
func appendArchiveMonth(months []ArchiveMonth, month time.Month, page *content.Page) []ArchiveMonth {
	for i := range months {
		if months[i].Month == month {
			months[i].Pages = append(months[i].Pages, page)
			return months
		}
	}
	return append(months, ArchiveMonth{Month: month, Pages: []*content.Page{page}})
}
