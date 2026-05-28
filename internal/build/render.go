package build

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
	"github.com/KurongTohsaka/chenhai-hugo/internal/theme"
)

// renderPages renders the paginated homepage and all individual non-draft pages.
func (b *Builder) renderPages(site *index.Site, public string) error {
	// Render paginated homepage
	published := site.PublishedPages()
	if err := b.renderPaginatedListPages(b.cfg.Title, published, public, "index.html", site, public); err != nil {
		return fmt.Errorf("homepage: %w", err)
	}

	// Render each non-draft page using single.html template
	for _, page := range site.Pages {
		if page.Draft {
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
				"tagCloud":   site.BuildTagCloud(),
			},
		}
		if err := b.renderToFile(pageData, filepath.Join(outDir, "index.html"), "single.html"); err != nil {
			return fmt.Errorf("page %q: %w", page.Title, err)
		}
	}
	return nil
}
