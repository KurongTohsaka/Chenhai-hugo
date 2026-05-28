package build

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
	"github.com/KurongTohsaka/chenhai-hugo/internal/theme"
	"github.com/KurongTohsaka/chenhai-hugo/themes/zhenhai"
)

// Builder orchestrates the full site build pipeline.
type Builder struct {
	cfg      *config.Config
	root     string
	renderer *content.Renderer
	engine   *theme.Engine
}

// New creates a new Builder.
func New(cfg *config.Config, root string, r *content.Renderer, e *theme.Engine) *Builder {
	return &Builder{cfg: cfg, root: root, renderer: r, engine: e}
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

		page.CalcReadingTime()
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// renderPages renders the homepage and all non-draft pages.
func (b *Builder) renderPages(site *index.Site, public string) error {
	// Render homepage (index.html) with all published pages
	published := site.PublishedPages()
	homepagePage := &content.Page{Title: b.cfg.Title}
	homepageData := &theme.TemplateData{
		Site:   site,
		Page:   homepagePage,
		Config: b.cfg,
		Extra:  map[string]interface{}{"pages": published},
	}
	if err := b.renderToFile(homepageData, filepath.Join(public, "index.html"), "index.html"); err != nil {
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
		}
		if err := b.renderToFile(pageData, filepath.Join(outDir, "index.html"), "single.html"); err != nil {
			return fmt.Errorf("page %q: %w", page.Title, err)
		}
	}
	return nil
}

// renderTaxonomies renders category and tag index and individual pages.
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

	// Individual category pages: /categories/<cat>/index.html
	for cat, pages := range site.Categories {
		catPageData := &theme.TemplateData{
			Site:   site,
			Page:   &content.Page{Title: cat},
			Config: b.cfg,
			Extra:  map[string]interface{}{"title": cat, "pages": pages},
		}
		catPageDir := filepath.Join(catDir, cat)
		if err := b.renderToFile(catPageData, filepath.Join(catPageDir, "index.html"), "list.html"); err != nil {
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

	// Individual tag pages: /tags/<tag>/index.html
	for tag, pages := range site.Tags {
		tagPageData := &theme.TemplateData{
			Site:   site,
			Page:   &content.Page{Title: tag},
			Config: b.cfg,
			Extra:  map[string]interface{}{"title": tag, "pages": pages},
		}
		tagPageDir := filepath.Join(tagDir, tag)
		if err := b.renderToFile(tagPageData, filepath.Join(tagPageDir, "index.html"), "list.html"); err != nil {
			return fmt.Errorf("tag %q: %w", tag, err)
		}
	}

	return nil
}

// renderArchives renders /archives/index.html with all published pages.
func (b *Builder) renderArchives(site *index.Site, public string) error {
	archiveDir := filepath.Join(public, "archives")
	published := site.PublishedPages()
	archiveData := &theme.TemplateData{
		Site:   site,
		Page:   &content.Page{Title: "Archives"},
		Config: b.cfg,
		Extra:  map[string]interface{}{"title": "Archives", "pages": published},
	}
	return b.renderToFile(archiveData, filepath.Join(archiveDir, "index.html"), "list.html")
}

// renderSearchIndex writes public/search-index.json.
func (b *Builder) renderSearchIndex(site *index.Site, public string) error {
	data, err := index.BuildSearchIndex(site.Pages)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(public, "search-index.json"), data, 0644)
}

// copyStatic copies all files from static/ to public/ recursively.
func (b *Builder) copyStatic(public string) error {
	staticDir := filepath.Join(b.root, "static")
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		return nil
	}
	return filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(staticDir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		outPath := filepath.Join(public, relPath)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(outPath, data, 0644)
	})
}

// copyThemeAssets copies embedded theme CSS and JS to public/assets/.
func (b *Builder) copyThemeAssets(public string) error {
	return fs.WalkDir(zhenhai.FS, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := zhenhai.FS.ReadFile(path)
		if err != nil {
			return err
		}
		outPath := filepath.Join(public, path)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(outPath, data, 0644)
	})
}

// writeSitemap generates a minimal XML sitemap.
func (b *Builder) writeSitemap(site *index.Site, public string) error {
	var buf strings.Builder
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	baseURL := strings.TrimRight(b.cfg.BaseURL, "/")

	// Homepage entry
	buf.WriteString(fmt.Sprintf("  <url><loc>%s/</loc></url>\n", baseURL))

	// Page entries
	for _, page := range site.PublishedPages() {
		buf.WriteString(fmt.Sprintf("  <url><loc>%s%s</loc></url>\n", baseURL, page.Permalink()))
	}

	buf.WriteString("</urlset>\n")
	return os.WriteFile(filepath.Join(public, "sitemap.xml"), []byte(buf.String()), 0644)
}

// writeRobotsTXT generates a simple robots.txt allowing all crawlers.
func (b *Builder) writeRobotsTXT(public string) error {
	var buf strings.Builder
	buf.WriteString("User-agent: *\n")
	buf.WriteString("Allow: /\n\n")
	buf.WriteString(fmt.Sprintf("Sitemap: %s/sitemap.xml\n", strings.TrimRight(b.cfg.BaseURL, "/")))
	return os.WriteFile(filepath.Join(public, "robots.txt"), []byte(buf.String()), 0644)
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
