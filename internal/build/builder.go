package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/internal/imagehost"
	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
	"github.com/KurongTohsaka/chenhai-hugo/internal/theme"
)

// Builder orchestrates the full site build pipeline.
type Builder struct {
	cfg       *config.Config
	root      string
	renderer  *content.Renderer
	engine    *theme.Engine
	imageHost *imagehost.Host
}

// New creates a new Builder.
func New(cfg *config.Config, root string, r *content.Renderer, e *theme.Engine) *Builder {
	return &Builder{
		cfg:       cfg,
		root:      root,
		renderer:  r,
		engine:    e,
		imageHost: imagehost.New(&cfg.ImageHost),
	}
}

// Build executes the complete build pipeline.
func (b *Builder) Build() error {
	public := filepath.Join(b.root, "public")
	os.RemoveAll(public)

	// 1. Collect all pages from content/
	pages, err := b.collectPages()
	if err != nil {
		return fmt.Errorf("collect pages: %w", err)
	}

	// 2. Build site index
	site := index.BuildSite(b.cfg, pages)

	// 3. Render pages
	if err := b.renderPages(site, public); err != nil {
		return fmt.Errorf("render pages: %w", err)
	}

	// 4. Render taxonomies
	if err := b.renderTaxonomies(site, public); err != nil {
		return fmt.Errorf("render taxonomies: %w", err)
	}

	// 5. Render archives
	if err := b.renderArchives(site, public); err != nil {
		return fmt.Errorf("render archives: %w", err)
	}

	// 6. Generate search index
	if err := b.renderSearchIndex(site, public); err != nil {
		return fmt.Errorf("search index: %w", err)
	}

	// 7. Copy static files
	if err := b.copyStatic(public); err != nil {
		return fmt.Errorf("copy static: %w", err)
	}

	// 8. Copy theme assets
	if err := b.copyThemeAssets(public); err != nil {
		return fmt.Errorf("copy theme assets: %w", err)
	}
	if err := b.copyThemeStatic(public); err != nil {
		return fmt.Errorf("copy theme static: %w", err)
	}

	// 9. SEO files
	if b.cfg.SEO.EnableSitemap {
		if err := b.writeSitemap(site, public); err != nil {
			return fmt.Errorf("write sitemap: %w", err)
		}
	}
	if b.cfg.SEO.EnableRobotsTXT {
		if err := b.writeRobotsTXT(public); err != nil {
			return fmt.Errorf("write robots.txt: %w", err)
		}
	}

	return nil
}

// renderSearchIndex writes public/search-index.json.
func (b *Builder) renderSearchIndex(site *index.Site, public string) error {
	data, err := index.BuildSearchIndex(site.Pages)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(public, "search-index.json"), data, 0644)
}

// renderToFile renders a template to a file, creating parent directories as needed.
func (b *Builder) renderToFile(data *theme.TemplateData, path, tmpl string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = b.engine.RenderPage(f, tmpl, data)
	f.Close()
	return err
}
