package build

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
	"github.com/KurongTohsaka/chenhai-hugo/internal/theme"
)

// renderPages renders the paginated homepage and all individual non-draft pages
// concurrently using a goroutine pool limited by runtime.NumCPU().
func (b *Builder) renderPages(site *index.Site, public string) error {
	// Render paginated homepage (always rendered — it may include changed pages)
	published := site.PublishedPages()
	if err := b.renderPaginatedListPages(b.cfg.Title, published, public, "index.html", site, public); err != nil {
		return fmt.Errorf("homepage: %w", err)
	}

	// Collect non-draft, non-skipped pages
	type renderTask struct {
		pageData *theme.TemplateData
		outDir   string
		tmpl     string
		title    string
	}
	var tasks []renderTask
	for _, page := range site.Pages {
		if page.Draft {
			continue
		}
		if b.skippedPaths[page.FilePath] {
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
		if related := site.RelatedPosts(page, 3); len(related) > 0 {
			pageData.Extra["related"] = related
		}
		if page.Series != "" {
			if seriesPages, ok := site.Series[page.Series]; ok {
				pageData.Extra["seriesPages"] = seriesPages
				pageData.Extra["seriesName"] = page.Series
				for i, p := range seriesPages {
					if p.FilePath == page.FilePath {
						pageData.Extra["seriesIdx"] = i
						if i > 0 {
							pageData.Extra["seriesPrev"] = seriesPages[i-1]
						}
						if i < len(seriesPages)-1 {
							pageData.Extra["seriesNext"] = seriesPages[i+1]
						}
						break
					}
				}
			}
		}
		tmpl := "single.html"
		if page.Layout != "" {
			tmpl = page.Layout + ".html"
		}
		tasks = append(tasks, renderTask{
			pageData: pageData,
			outDir:   outDir,
			tmpl:     tmpl,
			title:    page.Title,
		})
	}

	if len(tasks) == 0 {
		return nil
	}

	// Concurrent rendering with semaphore
	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	rendered := 0

	for _, tsk := range tasks {
		wg.Add(1)
		go func(t renderTask) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			err := b.renderToFile(t.pageData, filepath.Join(t.outDir, "index.html"), t.tmpl)
			mu.Lock()
			if err != nil && firstErr == nil {
				firstErr = fmt.Errorf("page %q: %w", t.title, err)
			}
			rendered++
			fmt.Printf("\r  渲染 %d/%d 篇文章", rendered, len(tasks))
			mu.Unlock()
		}(tsk)
	}
	wg.Wait()

	fmt.Println()
	if firstErr != nil {
		return firstErr
	}
	return nil
}
