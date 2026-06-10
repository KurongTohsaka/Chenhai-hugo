package build

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
)

// collectPages walks content/ and parses all .md files.
// For files that haven't changed (per build cache), it parses front matter
// only and skips expensive HTML rendering to speed up incremental builds.
func (b *Builder) collectPages() ([]*content.Page, error) {
	contentDir := filepath.Join(b.root, "content")
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return nil, nil
	}

	// Track seen files for deletion detection
	seenPaths := make(map[string]bool)

	var pages []*content.Page
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		seenPaths[path] = true

		relPath, err := filepath.Rel(contentDir, path)
		if err != nil {
			return fmt.Errorf("rel path %s: %w", path, err)
		}
		relPath = filepath.ToSlash(relPath)

		// Check if file is unchanged (cached and hash matches)
		isUnchanged := false
		if len(b.cache.Files) > 0 {
			changed, _ := b.cache.isChanged(path)
			isUnchanged = !changed
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		// For unchanged files: parse front matter only (metadata + raw content),
		// skip expensive Markdown -> HTML rendering.
		if isUnchanged {
			page, _, err := content.ParseFrontMatter(raw)
			if err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}

			page.Content = "" // marker: skip template rendering, output already exists
			page.FilePath = path
			page.RelPath = relPath

			// Section is the first path segment under content/
			parts := strings.SplitN(relPath, "/", 2)
			if len(parts) > 0 {
				page.Section = parts[0]
				// If the file is directly under content/ (e.g., "about.md"), strip the .md extension
				if len(parts) == 1 {
					page.Section = strings.TrimSuffix(parts[0], ".md")
				}
			}

			pages = append(pages, page)
			b.skippedPaths[path] = true

			if page.Title != "" {
				fmt.Printf("  (跳过未变更: %s)\n", page.Title)
			} else {
				fmt.Printf("  (跳过未变更: %s)\n", relPath)
			}
			return nil
		}

		// Full processing for changed / new files
		if b.imageHost != nil {
			processed, err := b.imageHost.Process(raw, filepath.Dir(path))
			if err != nil {
				log.Printf("image host warning: %v", err)
			} else {
				raw = processed
			}
		}

		page, body, err := content.ParseFrontMatter(raw)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		html, err := b.renderer.RenderHTML(body)
		if err != nil {
			return fmt.Errorf("render %s: %w", path, err)
		}

		page.Content = html
		// Transform mermaid code blocks for Mermaid.js rendering
		page.Content = transformMermaidBlocks(page.Content)
		page.FilePath = path
		page.RelPath = relPath

		// Section is the first path segment under content/
		parts := strings.SplitN(relPath, "/", 2)
		if len(parts) > 0 {
			page.Section = parts[0]
			// If the file is directly under content/ (e.g., "about.md"), strip the .md extension
			if len(parts) == 1 {
				page.Section = strings.TrimSuffix(parts[0], ".md")
			}
		}

		pages = append(pages, page)
		b.cache.updateFile(path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Handle deleted files: remove from public/ and cache
	for cachedPath := range b.cache.Files {
		if !seenPaths[cachedPath] {
			b.removeDeletedPage(cachedPath)
			b.cache.deleteFile(cachedPath)
		}
	}

	return pages, nil
}

var mermaidBlockRe = regexp.MustCompile(`<pre><code class="language-mermaid">([\s\S]*?)</code></pre>`)

func transformMermaidBlocks(html string) string {
	return mermaidBlockRe.ReplaceAllString(html, `<pre class="mermaid">$1</pre>`)
}
