package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/internal/imagehost"
	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
	"github.com/KurongTohsaka/chenhai-hugo/internal/theme"
)

// Builder orchestrates the full site build pipeline.
type Builder struct {
	cfg          *config.Config
	root         string
	renderer     *content.Renderer
	engine       *theme.Engine
	imageHost    *imagehost.Host
	cache        *BuildCache
	skippedPaths map[string]bool // content file paths skipped because unchanged
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
	fmt.Println("Chenhai 构建中...")

	public := filepath.Join(b.root, "public")

	// Load build cache
	b.cache, _ = loadCache(b.root)
	defer b.cache.save()

	// Check if config or templates changed -> full rebuild
	configPath := filepath.Join(b.root, "config.yaml")
	configChanged := b.cache.configChanged(configPath)
	templatesChanged := b.hasTemplatesChanged()

	fullRebuild := configChanged || templatesChanged

	if fullRebuild {
		if configChanged {
			fmt.Println("  (配置文件已变更，全量重建)")
		} else {
			fmt.Println("  (模板文件已变更，全量重建)")
		}
		os.RemoveAll(public)
		b.cache.Files = make(map[string]string)
		b.cache.Config = ""
	} else {
		fmt.Println("  (增量构建模式)")
	}

	// 1. Collect all pages from content/
	b.skippedPaths = make(map[string]bool)
	fmt.Print("  扫描 content/ ... ")
	pages, err := b.collectPages()
	if err != nil {
		return fmt.Errorf("collect pages: %w", err)
	}
	pubCount := 0
	for _, p := range pages {
		if !p.Draft {
			pubCount++
		}
	}
	skipped := len(b.skippedPaths)
	if skipped > 0 {
		fmt.Printf("发现 %d 篇文章（%d 已发布，%d 跳过未变更）\n", len(pages), pubCount, skipped)
	} else {
		fmt.Printf("发现 %d 篇文章（%d 已发布）\n", len(pages), pubCount)
	}

	// 2. Build site index
	site := index.BuildSite(b.cfg, pages)
	fmt.Printf("  标签: %d | 分类: %d\n", len(site.Tags), len(site.Categories))

	// 3. Render pages
	fmt.Print("  渲染页面 ... ")
	if err := b.renderPages(site, public); err != nil {
		return fmt.Errorf("render pages: %w", err)
	}
	fmt.Println("完成")

	// 4. Render taxonomies
	fmt.Print("  渲染归档与分类 ... ")
	if err := b.renderTaxonomies(site, public); err != nil {
		return fmt.Errorf("render taxonomies: %w", err)
	}
	fmt.Println("完成")

	// 5. Render archives
	if err := b.renderArchives(site, public); err != nil {
		return fmt.Errorf("render archives: %w", err)
	}

	// 6. Generate search index
	fmt.Print("  生成搜索索引 ... ")
	if err := b.renderSearchIndex(site, public); err != nil {
		return fmt.Errorf("search index: %w", err)
	}
	fmt.Println("完成")

	// 7. Copy static files
	fmt.Print("  复制静态资源 ... ")
	if err := b.copyStatic(public); err != nil {
		return fmt.Errorf("copy static: %w", err)
	}
	fmt.Println("完成")

	// 8. Copy theme assets
	fmt.Print("  复制主题资源 ... ")
	if err := b.copyThemeAssets(public); err != nil {
		return fmt.Errorf("copy theme assets: %w", err)
	}
	if err := b.copyThemeStatic(public); err != nil {
		return fmt.Errorf("copy theme static: %w", err)
	}
	fmt.Println("完成")

	// 9. SEO files
	if b.cfg.SEO.EnableSitemap {
		fmt.Print("  生成 Sitemap ... ")
		if err := b.writeSitemap(site, public); err != nil {
			return fmt.Errorf("write sitemap: %w", err)
		}
		fmt.Println("完成")
	}
	if b.cfg.SEO.EnableRobotsTXT {
		fmt.Print("  生成 robots.txt ... ")
		if err := b.writeRobotsTXT(public); err != nil {
			return fmt.Errorf("write robots.txt: %w", err)
		}
		fmt.Println("完成")
	}

	// 10. Update config hash in cache
	if _, err := os.Stat(configPath); err == nil {
		b.cache.updateConfig(configPath)
	}

	fmt.Printf("\n✓ 构建完成 → %s\n", public)
	return nil
}

// hasTemplatesChanged checks if any template file has changed (full rebuild trigger).
func (b *Builder) hasTemplatesChanged() bool {
	dirs := []string{filepath.Join(b.root, "layouts")}
	if b.cfg.Theme != "zhenhai" {
		dirs = append(dirs, filepath.Join(b.root, "themes", b.cfg.Theme, "layouts"))
	}
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		changed := false
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || changed {
				return nil
			}
			if strings.HasSuffix(path, ".html") {
				c, _ := b.cache.isChanged(path)
				if c {
					changed = true
				}
			}
			return nil
		})
		if changed {
			return true
		}
	}
	return false
}

// removeDeletedPage removes the public/ output for a content file that no longer exists.
func (b *Builder) removeDeletedPage(path string) {
	contentDir := filepath.Join(b.root, "content")
	relPath, err := filepath.Rel(contentDir, path)
	if err != nil {
		return
	}
	relPath = filepath.ToSlash(relPath)

	// Compute permalink from relPath (same logic as page.Permalink)
	rel := strings.TrimSuffix(relPath, ".md")
	rel = strings.TrimSuffix(rel, "/index")
	outDir := filepath.Join(b.root, "public", rel)

	if _, err := os.Stat(outDir); err == nil {
		os.RemoveAll(outDir)
		fmt.Printf("  (移除已删除: %s)\n", relPath)
	}

	// Clean up empty parent directories
	parent := filepath.Dir(outDir)
	if parent != filepath.Join(b.root, "public") {
		if entries, _ := os.ReadDir(parent); len(entries) == 0 {
			os.Remove(parent)
		}
	}
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
