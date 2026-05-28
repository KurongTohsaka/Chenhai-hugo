package build

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
	"github.com/KurongTohsaka/chenhai-hugo/internal/theme"
)

// renderPages renders the paginated homepage and all individual non-draft pages.
// Pages that are unchanged (skipped) are not re-rendered — their output already exists in public/.
func (b *Builder) renderPages(site *index.Site, public string) error {
	// Render paginated homepage (always rendered — it may include changed pages)
	published := site.PublishedPages()
	if err := b.renderPaginatedListPages(b.cfg.Title, published, public, "index.html", site, public); err != nil {
		return fmt.Errorf("homepage: %w", err)
	}

	// Render each non-draft page using single.html template
	total := 0
	for _, page := range site.Pages {
		if page.Draft {
			continue
		}
		if b.skippedPaths[page.FilePath] {
			continue // skip unchanged pages from count
		}
		total++
	}
	rendered := 0
	skipped := 0
	for _, page := range site.Pages {
		if page.Draft {
			continue
		}

		// Skip template rendering for unchanged pages — output already exists in public/
		if b.skippedPaths[page.FilePath] {
			skipped++
			continue
		}

		permalink := page.Permalink()
		outDir := filepath.Join(public, strings.Trim(permalink, "/"))
		pageData := &theme.TemplateData{
			Site:   site,
			Page:   page,
			Config: b.cfg,
			Extra: map[string]interface{}{
				"hasMath":    strings.Contains(page.Content, "$$") || strings.Contains(page.Content, "\\("),
				"hasMermaid": strings.Contains(page.Content, "mermaid"),
			},
		}
		tmpl := "single.html"
		if page.Layout != "" {
			tmpl = page.Layout + ".html"
		}
		if err := b.renderToFile(pageData, filepath.Join(outDir, "index.html"), tmpl); err != nil {
			return fmt.Errorf("page %q: %w", page.Title, err)
		}
		rendered++
		fmt.Printf("\r  渲染 %d/%d 篇文章（跳过 %d）", rendered, total+skipped, skipped)
	}
	if total+skipped > 0 {
		fmt.Println()
	}
	return nil
}
