package build

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
)

// collectPages walks content/ and parses all .md files.
func (b *Builder) collectPages() ([]*content.Page, error) {
	contentDir := filepath.Join(b.root, "content")
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return nil, nil
	}

	var pages []*content.Page
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		relPath, err := filepath.Rel(contentDir, path)
		if err != nil {
			return fmt.Errorf("rel path %s: %w", path, err)
		}
		relPath = filepath.ToSlash(relPath)

		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		// Process images through image host (before front matter parsing)
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
		page.FilePath = path
		page.RelPath = relPath

		// Section is the first path segment under content/
		parts := strings.SplitN(relPath, "/", 2)
		if len(parts) > 0 {
			page.Section = parts[0]
		}

		pages = append(pages, page)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pages, nil
}
