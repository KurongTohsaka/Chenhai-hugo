package build

import (
	"fmt"
	"path/filepath"

	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
	"github.com/KurongTohsaka/chenhai-hugo/internal/theme"
)

// renderTaxonomies renders category/tag index pages and paginated individual pages.
func (b *Builder) renderTaxonomies(site *index.Site, public string) error {
	// Categories index: /categories/index.html
	catDir := filepath.Join(public, "categories")
	catIndexData := &theme.TemplateData{
		Site:   site,
		Page:   &content.Page{Title: "Categories"},
		Config: b.cfg,
		Extra:  map[string]interface{}{"title": "Categories"},
	}
	if err := b.renderToFile(catIndexData, filepath.Join(catDir, "index.html"), "taxonomy.html"); err != nil {
		return fmt.Errorf("categories index: %w", err)
	}

	// Individual category pages with pagination: /categories/<cat>/
	for cat, pages := range site.Categories {
		catPageDir := filepath.Join(catDir, cat)
		if err := b.renderPaginatedListPages(cat, pages, catPageDir, "list.html", site, public); err != nil {
			return fmt.Errorf("category %q: %w", cat, err)
		}
	}

	// Tags index: /tags/index.html
	tagDir := filepath.Join(public, "tags")
	tagIndexData := &theme.TemplateData{
		Site:   site,
		Page:   &content.Page{Title: "Tags"},
		Config: b.cfg,
		Extra:  map[string]interface{}{"title": "Tags"},
	}
	if err := b.renderToFile(tagIndexData, filepath.Join(tagDir, "index.html"), "taxonomy.html"); err != nil {
		return fmt.Errorf("tags index: %w", err)
	}

	// Individual tag pages with pagination: /tags/<tag>/
	for tag, pages := range site.Tags {
		tagPageDir := filepath.Join(tagDir, tag)
		if err := b.renderPaginatedListPages(tag, pages, tagPageDir, "list.html", site, public); err != nil {
			return fmt.Errorf("tag %q: %w", tag, err)
		}
	}

	return nil
}

// renderSeries renders series index and per-series listing pages.
func (b *Builder) renderSeries(site *index.Site, public string) error {
	if len(site.Series) == 0 {
		return nil
	}
	seriesDir := filepath.Join(public, "series")

	// Series index: /series/index.html
	seriesIndexData := &theme.TemplateData{
		Site:   site,
		Page:   &content.Page{Title: "Series"},
		Config: b.cfg,
		Extra:  map[string]interface{}{"title": "Series"},
	}
	if err := b.renderToFile(seriesIndexData, filepath.Join(seriesDir, "index.html"), "taxonomy.html"); err != nil {
		return fmt.Errorf("series index: %w", err)
	}

	// Individual series pages with pagination: /series/<name>/
	for name, pages := range site.Series {
		spd := filepath.Join(seriesDir, name)
		if err := b.renderPaginatedListPages(name, pages, spd, "list.html", site, public); err != nil {
			return fmt.Errorf("series %q: %w", name, err)
		}
	}
	return nil
}
