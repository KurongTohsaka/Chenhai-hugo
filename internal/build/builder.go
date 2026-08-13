package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sync"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/internal/imagehost"
	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
	"github.com/KurongTohsaka/chenhai-hugo/internal/theme"
)

// cachedEngine persists across Builder instances for incremental build engine reuse.
var (
	cachedEngine   *theme.Engine
	cachedEngineMu sync.Mutex
)

// Builder orchestrates the full site build pipeline.
type Builder struct {
	cfg          *config.Config
	root         string
	renderer     *content.Renderer
	engine       *theme.Engine
	imageHost    *imagehost.Host
	showDrafts   bool
	cache        *BuildCache
	skippedPaths map[string]bool // content file paths skipped because unchanged
}

// New creates a new Builder.
func New(cfg *config.Config, root string, r *content.Renderer, e *theme.Engine, showDrafts bool) *Builder {
	return &Builder{
		cfg:        cfg,
		root:       root,
		renderer:   r,
		engine:     e,
		imageHost:  imagehost.New(&cfg.ImageHost),
		showDrafts: showDrafts,
	}
}

// Build executes the complete build pipeline.
func (b *Builder) Build() error {
	fmt.Println("Chenhai 构建中...")

	var totalStart = time.Now()
	var t time.Time

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
	// P2: reuse cached engine when templates haven't changed
	cachedEngineMu.Lock()
	if !fullRebuild && cachedEngine != nil {
		b.engine = cachedEngine
		cachedEngineMu.Unlock()
		fmt.Println("  (复用模板缓存)")
	} else {
		cachedEngine = b.engine
		cachedEngineMu.Unlock()
	}

	b.skippedPaths = make(map[string]bool)
	t = time.Now()
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
		fmt.Printf("发现 %d 篇文章（%d 已发布，%d 跳过未变更）(%s)\n", len(pages), pubCount, skipped, time.Since(t).Round(time.Millisecond))
	} else {
		fmt.Printf("发现 %d 篇文章（%d 已发布）(%s)\n", len(pages), pubCount, time.Since(t).Round(time.Millisecond))
	}

	// 2. Build site index
	t = time.Now()
	site := index.BuildSite(b.cfg, pages, b.showDrafts)
	fmt.Printf("  标签: %d | 分类: %d (%s)\n", len(site.Tags), len(site.Categories), time.Since(t).Round(time.Millisecond))

	// 3. Render pages
	t = time.Now()
	fmt.Print("  渲染页面 ... ")
	if err := b.renderPages(site, public); err != nil {
		return fmt.Errorf("render pages: %w", err)
	}
	fmt.Printf("完成 (%s)\n", time.Since(t).Round(time.Millisecond))

	// 4. Render taxonomies
	t = time.Now()
	fmt.Print("  渲染归档与分类 ... ")
	if err := b.renderTaxonomies(site, public); err != nil {
		return fmt.Errorf("render taxonomies: %w", err)
	}
	if err := b.renderSeries(site, public); err != nil { return fmt.Errorf("render series: %w", err) }
	if err := b.renderArchives(site, public); err != nil {
		return fmt.Errorf("render archives: %w", err)
	}
	fmt.Printf("完成 (%s)\n", time.Since(t).Round(time.Millisecond))

	// 6. Generate search index
	t = time.Now()
	fmt.Print("  生成搜索索引 ... ")
	if err := b.renderSearchIndex(site, public); err != nil {
		return fmt.Errorf("search index: %w", err)
	}
	fmt.Printf("完成 (%s)\n", time.Since(t).Round(time.Millisecond))

	// 7. Copy static files
	t = time.Now()
	fmt.Print("  复制静态资源 ... ")
	if err := b.copyStatic(public); err != nil {
		return fmt.Errorf("copy static: %w", err)
	}
	fmt.Printf("完成 (%s)\n", time.Since(t).Round(time.Millisecond))

	// 8. Copy theme assets
	t = time.Now()
	fmt.Print("  复制主题资源 ... ")
	if err := b.copyThemeAssets(public); err != nil {
		return fmt.Errorf("copy theme assets: %w", err)
	}
	if err := b.copyThemeStatic(public); err != nil {
		return fmt.Errorf("copy theme static: %w", err)
	}
	fmt.Printf("完成 (%s)\n", time.Since(t).Round(time.Millisecond))

	// 9. Render 404 page
	t = time.Now()
	fmt.Print("  生成 404 页面 ... ")
	if err := b.render404(site, public); err != nil {
		return fmt.Errorf("render 404: %w", err)
	}
	fmt.Printf("完成 (%s)\n", time.Since(t).Round(time.Millisecond))

	// 10. SEO files
	if b.cfg.SEO.EnableSitemap {
		t = time.Now()
		fmt.Print("  生成 Sitemap ... ")
		if err := b.writeSitemap(site, public); err != nil {
			return fmt.Errorf("write sitemap: %w", err)
		}
		fmt.Printf("完成 (%s)\n", time.Since(t).Round(time.Millisecond))
	}
	if b.cfg.SEO.EnableRobotsTXT {
		t = time.Now()
		fmt.Print("  生成 robots.txt ... ")
		if err := b.writeRobotsTXT(public); err != nil {
			return fmt.Errorf("write robots.txt: %w", err)
		}
		fmt.Printf("完成 (%s)\n", time.Since(t).Round(time.Millisecond))
	}

	// 11. RSS/Atom feed
	t = time.Now()
	fmt.Print("  生成 RSS 订阅 ... ")
	feed, err := generateRSS(b.cfg, site.Pages)
	if err != nil {
		return fmt.Errorf("generate rss: %w", err)
	}
	if len(feed) > 0 {
		if err := os.WriteFile(filepath.Join(public, "atom.xml"), feed, 0644); err != nil {
			return fmt.Errorf("write atom.xml: %w", err)
		}
		fmt.Printf("完成 (%s)\n", time.Since(t).Round(time.Millisecond))
	} else if b.cfg.RSS.Enabled && b.cfg.BaseURL == "" {
		fmt.Println("  ⚠ 未配置 baseURL，跳过 RSS 生成（config.yaml 添加 baseURL 即可）")
	}

	// 12. Update config hash in cache
	if _, err := os.Stat(configPath); err == nil {
		b.cache.updateConfig(configPath)
	}

	fmt.Printf("\n✓ 构建完成 → %s (总耗时 %s)\n", public, time.Since(totalStart).Round(time.Millisecond))
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

// render404 generates the 404.html page.
func (b *Builder) render404(site *index.Site, public string) error {
	data := &theme.TemplateData{
		Site:   site,
		Page:   &content.Page{Title: "404"},
		Config: b.cfg,
	}
	return b.renderToFile(data, filepath.Join(public, "404.html"), "404.html")
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
