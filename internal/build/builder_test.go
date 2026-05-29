package build_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KurongTohsaka/chenhai-hugo/internal/build"
	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/internal/theme"
)

func TestBuild(t *testing.T) {
	root := t.TempDir()

	// Create config.yaml
	cfgYAML := []byte(`title: "Test Blog"
baseURL: "https://example.com"
seo:
  enableSitemap: true
  enableRobotsTXT: true
`)
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), cfgYAML, 0644); err != nil {
		t.Fatal(err)
	}

	// Create content/posts/
	postsDir := filepath.Join(root, "content", "posts")
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a test post with full front matter
	postMD := []byte(`---
title: "Hello World"
date: "2024-01-15"
categories: ["Tech"]
tags: ["go", "test"]
slug: "test-post"
description: "A test post description"
summary: "This is a test post"
---

This is the post content.
`)
	if err := os.WriteFile(filepath.Join(postsDir, "test-post.md"), postMD, 0644); err != nil {
		t.Fatal(err)
	}

	// Create static/ directory with a favicon
	staticDir := filepath.Join(root, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, "favicon.ico"), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	// Load config
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	// Create renderer
	renderer := content.NewRenderer(cfg.Markup.Highlight.Style, cfg.Markup.Highlight.LineNumbers)

	// Create engine
	engine, err := theme.New(cfg, root)
	if err != nil {
		t.Fatal(err)
	}

	// Create builder and build
	builder := build.New(cfg, root, renderer, engine)
	if err := builder.Build(); err != nil {
		t.Fatal(err)
	}

	public := filepath.Join(root, "public")

	// --- Verify homepage ---
	homepagePath := filepath.Join(public, "index.html")
	homepageContent, err := os.ReadFile(homepagePath)
	if err != nil {
		t.Fatal("homepage (public/index.html) was not generated:", err)
	}
	if !strings.Contains(string(homepageContent), "Test Blog") {
		t.Error("homepage should contain the site title 'Test Blog'")
	}

	// --- Verify post page ---
	postPagePath := filepath.Join(public, "posts", "test-post", "index.html")
	postContent, err := os.ReadFile(postPagePath)
	if err != nil {
		t.Fatal("post page (public/posts/test-post/index.html) was not generated:", err)
	}
	if !strings.Contains(string(postContent), "Hello World") {
		t.Error("post page should contain the post title 'Hello World'")
	}

		// --- Verify KaTeX/Mermaid are NOT injected for non-math content ---
		if strings.Contains(string(postContent), "katex.min.css") {
			t.Error("non-math post should NOT include KaTeX CSS")
		}
		if strings.Contains(string(postContent), "katex.min.js") {
			t.Error("non-math post should NOT include KaTeX JS")
		}
		if strings.Contains(string(postContent), "mermaid.min.js") {
			t.Error("non-math post should NOT include Mermaid JS")
		}

	// --- Verify categories index ---
	categoriesIndexPath := filepath.Join(public, "categories", "index.html")
	if _, err := os.Stat(categoriesIndexPath); os.IsNotExist(err) {
		t.Error("categories index (public/categories/index.html) was not generated")
	}

	// --- Verify individual category page ---
	categoryPagePath := filepath.Join(public, "categories", "Tech", "index.html")
	if _, err := os.Stat(categoryPagePath); os.IsNotExist(err) {
		t.Error("category page (public/categories/Tech/index.html) was not generated")
	}

	// --- Verify tags index ---
	tagsIndexPath := filepath.Join(public, "tags", "index.html")
	if _, err := os.Stat(tagsIndexPath); os.IsNotExist(err) {
		t.Error("tags index (public/tags/index.html) was not generated")
	}

	// --- Verify individual tag pages ---
	tagGoPath := filepath.Join(public, "tags", "go", "index.html")
	if _, err := os.Stat(tagGoPath); os.IsNotExist(err) {
		t.Error("tag page (public/tags/go/index.html) was not generated")
	}
	tagTestPath := filepath.Join(public, "tags", "test", "index.html")
	if _, err := os.Stat(tagTestPath); os.IsNotExist(err) {
		t.Error("tag page (public/tags/test/index.html) was not generated")
	}

	// --- Verify archives ---
	archivesPath := filepath.Join(public, "archives", "index.html")
	if _, err := os.Stat(archivesPath); os.IsNotExist(err) {
		t.Error("archives (public/archives/index.html) was not generated")
	}

	// --- Verify search-index.json is valid JSON with correct data ---
	searchIndexPath := filepath.Join(public, "search-index.json")
	searchData, err := os.ReadFile(searchIndexPath)
	if err != nil {
		t.Fatal("search-index.json was not generated:", err)
	}
	var entries []map[string]interface{}
	if err := json.Unmarshal(searchData, &entries); err != nil {
		t.Fatal("search-index.json is not valid JSON:", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 search entry, got %d", len(entries))
	}
	if entries[0]["title"] != "Hello World" {
		t.Errorf("search entry title mismatch: expected 'Hello World', got %v", entries[0]["title"])
	}
	if entries[0]["category"] != "Tech" {
		t.Errorf("search entry category mismatch: expected 'Tech', got %v", entries[0]["category"])
	}
	if entries[0]["tags"] == nil {
		t.Error("search entry tags should not be nil")
	}
	if entries[0]["url"] != "/posts/test-post/" {
		t.Errorf("search entry url mismatch: expected '/posts/test-post/', got %v", entries[0]["url"])
	}
	if entries[0]["date"] != "2024-01-15" {
		t.Errorf("search entry date mismatch: expected '2024-01-15', got %v", entries[0]["date"])
	}

	// --- Verify sitemap ---
	sitemapPath := filepath.Join(public, "sitemap.xml")
	sitemapData, err := os.ReadFile(sitemapPath)
	if err != nil {
		t.Fatal("sitemap.xml was not generated:", err)
	}
	if !strings.Contains(string(sitemapData), "https://example.com/") {
		t.Error("sitemap should contain the base URL")
	}
	if !strings.Contains(string(sitemapData), "/posts/test-post/") {
		t.Error("sitemap should contain the post permalink")
	}

	// --- Verify robots.txt ---
	robotsPath := filepath.Join(public, "robots.txt")
	robotsData, err := os.ReadFile(robotsPath)
	if err != nil {
		t.Fatal("robots.txt was not generated:", err)
	}
	if !strings.Contains(string(robotsData), "User-agent: *") {
		t.Error("robots.txt should contain 'User-agent: *'")
	}
	if !strings.Contains(string(robotsData), "Sitemap:") {
		t.Error("robots.txt should contain a Sitemap directive")
	}

	// --- Verify static file was copied ---
	staticOutputPath := filepath.Join(public, "favicon.ico")
	if _, err := os.Stat(staticOutputPath); os.IsNotExist(err) {
		t.Error("static file (public/favicon.ico) was not copied")
	}

	// --- Verify theme assets were copied ---
	themeCSSPath := filepath.Join(public, "assets", "css", "style.css")
	if _, err := os.Stat(themeCSSPath); os.IsNotExist(err) {
		t.Error("theme CSS (public/assets/css/style.css) was not copied")
	}
	themeJSPath := filepath.Join(public, "assets", "js", "main.js")
	if _, err := os.Stat(themeJSPath); os.IsNotExist(err) {
		t.Error("theme JS (public/assets/js/main.js) was not copied")
	}
	searchJSPath := filepath.Join(public, "assets", "js", "search.js")
	if _, err := os.Stat(searchJSPath); os.IsNotExist(err) {
		t.Error("search JS (public/assets/js/search.js) was not copied")
	}
}

func TestBuild_EmptyContent(t *testing.T) {
	root := t.TempDir()

	// Create config without any content
	cfgYAML := []byte(`title: "Empty Site"
baseURL: "https://example.com"
`)
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), cfgYAML, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	renderer := content.NewRenderer(cfg.Markup.Highlight.Style, cfg.Markup.Highlight.LineNumbers)
	engine, err := theme.New(cfg, root)
	if err != nil {
		t.Fatal(err)
	}

	builder := build.New(cfg, root, renderer, engine)
	if err := builder.Build(); err != nil {
		t.Fatal(err)
	}

	public := filepath.Join(root, "public")

	// Homepage should still be generated
	if _, err := os.Stat(filepath.Join(public, "index.html")); os.IsNotExist(err) {
		t.Error("homepage should be generated even with empty content")
	}

	// Search index should be valid empty JSON
	searchData, err := os.ReadFile(filepath.Join(public, "search-index.json"))
	if err != nil {
		t.Fatal("search-index.json was not generated:", err)
	}
	var entries []map[string]interface{}
	if err := json.Unmarshal(searchData, &entries); err != nil {
		t.Fatal("search-index.json is not valid JSON:", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty search index, got %d entries", len(entries))
	}
}

func TestBuild_DraftSkipped(t *testing.T) {
	root := t.TempDir()

	// Create config
	cfgYAML := []byte(`title: "Draft Test"
baseURL: "https://example.com"
`)
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), cfgYAML, 0644); err != nil {
		t.Fatal(err)
	}

	// Create content with one draft and one published post
	postsDir := filepath.Join(root, "content", "posts")
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		t.Fatal(err)
	}

	draftMD := []byte(`---
title: "Draft Post"
date: "2024-06-01"
draft: true
---

This is a draft.
`)
	if err := os.WriteFile(filepath.Join(postsDir, "draft-post.md"), draftMD, 0644); err != nil {
		t.Fatal(err)
	}

	publishedMD := []byte(`---
title: "Published Post"
date: "2024-06-15"
---

This is published.
`)
	if err := os.WriteFile(filepath.Join(postsDir, "published-post.md"), publishedMD, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	renderer := content.NewRenderer(cfg.Markup.Highlight.Style, cfg.Markup.Highlight.LineNumbers)
	engine, err := theme.New(cfg, root)
	if err != nil {
		t.Fatal(err)
	}

	builder := build.New(cfg, root, renderer, engine)
	if err := builder.Build(); err != nil {
		t.Fatal(err)
	}

	public := filepath.Join(root, "public")

	// Draft page should NOT be rendered
	draftPagePath := filepath.Join(public, "posts", "draft-post", "index.html")
	if _, err := os.Stat(draftPagePath); !os.IsNotExist(err) {
		t.Error("draft page should NOT be generated as HTML")
	}

	// Published page SHOULD be rendered
	pubPagePath := filepath.Join(public, "posts", "published-post", "index.html")
	if _, err := os.Stat(pubPagePath); os.IsNotExist(err) {
		t.Error("published page should be generated as HTML")
	}

	// Search index should only contain published post
	searchData, err := os.ReadFile(filepath.Join(public, "search-index.json"))
	if err != nil {
		t.Fatal("search-index.json was not generated:", err)
	}
	var entries []map[string]interface{}
	if err := json.Unmarshal(searchData, &entries); err != nil {
		t.Fatal("search-index.json is not valid JSON:", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 search entry (draft excluded), got %d", len(entries))
	}
	if entries[0]["title"] != "Published Post" {
		t.Errorf("expected search entry 'Published Post', got %v", entries[0]["title"])
	}
}

func TestBuild_Pagination(t *testing.T) {
	root := t.TempDir()

	// Create config with postsPerPage=1
	cfgYAML := []byte(`title: "Pagination Test"
baseURL: "https://example.com"
themeConfig:
  postsPerPage: 1
`)
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), cfgYAML, 0644); err != nil {
		t.Fatal(err)
	}

	// Create content directory
	postsDir := filepath.Join(root, "content", "posts")
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create 3 published posts, all in same category/tag for taxonomy pagination test
	for i, title := range []string{"Post One", "Post Two", "Post Three"} {
		postMD := []byte(fmt.Sprintf(`---
title: "%s"
date: "2024-01-%02d"
categories: ["Demo"]
tags: ["pagination"]
---

Content for %s.
`, title, 15-i, title))
		if err := os.WriteFile(filepath.Join(postsDir, fmt.Sprintf("post-%d.md", i+1)), postMD, 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	renderer := content.NewRenderer(cfg.Markup.Highlight.Style, cfg.Markup.Highlight.LineNumbers)
	engine, err := theme.New(cfg, root)
	if err != nil {
		t.Fatal(err)
	}

	builder := build.New(cfg, root, renderer, engine)
	if err := builder.Build(); err != nil {
		t.Fatal(err)
	}

	public := filepath.Join(root, "public")

	// --- Homepage pagination ---
	// Page 1 should exist and show first page of paginated data
	page1Path := filepath.Join(public, "index.html")
	page1Content, err := os.ReadFile(page1Path)
	if err != nil {
		t.Fatal("homepage page 1 (public/index.html) was not generated:", err)
	}
	if !strings.Contains(string(page1Content), "1 / 3") {
		t.Error("homepage page 1 should show '1 / 3' (page number / total pages)")
	}
	// Page 1 should have "next" link but no "prev"
	if !strings.Contains(string(page1Content), "下一页") {
		t.Error("homepage page 1 should have next page link")
	}
	if strings.Contains(string(page1Content), "上一页") {
		t.Error("homepage page 1 should NOT have previous page link")
	}
	// Page 1 should show the most recent post (Post One with date 2024-01-14)
	// The post title appears in the post-cards list <a href="/posts/post-1/">Post One</a>
	if !strings.Contains(string(page1Content), "Post One") {
		t.Error("homepage page 1 should show 'Post One'")
	}
	// Check that Post Two does NOT appear in the post cards section
	// (it may appear in sidebar "Recent Posts" which is expected)
	postCardsSection := string(page1Content)
	if idx := strings.Index(postCardsSection, `<ul class="post-cards">`); idx >= 0 {
		if endIdx := strings.Index(postCardsSection[idx:], `</ul>`); endIdx >= 0 {
			postCardsSection = postCardsSection[idx : idx+endIdx+5]
		}
	}
	if strings.Contains(postCardsSection, "Post Two") {
		t.Error("homepage page 1 post cards should NOT contain 'Post Two'")
	}

	// Page 2 should exist
	page2Path := filepath.Join(public, "page", "2", "index.html")
	page2Content, err := os.ReadFile(page2Path)
	if err != nil {
		t.Fatal("homepage page 2 (public/page/2/index.html) was not generated:", err)
	}
	if !strings.Contains(string(page2Content), "2 / 3") {
		t.Error("homepage page 2 should show '2 / 3'")
	}
	if !strings.Contains(string(page2Content), "上一页") {
		t.Error("homepage page 2 should have previous page link")
	}
	if !strings.Contains(string(page2Content), "下一页") {
		t.Error("homepage page 2 should have next page link")
	}
	if !strings.Contains(string(page2Content), "Post Two") {
		t.Error("homepage page 2 should show 'Post Two'")
	}

	// Page 3 should exist
	page3Path := filepath.Join(public, "page", "3", "index.html")
	page3Content, err := os.ReadFile(page3Path)
	if err != nil {
		t.Fatal("homepage page 3 (public/page/3/index.html) was not generated:", err)
	}
	if !strings.Contains(string(page3Content), "3 / 3") {
		t.Error("homepage page 3 should show '3 / 3'")
	}
	if !strings.Contains(string(page3Content), "上一页") {
		t.Error("homepage page 3 should have previous page link")
	}
	if strings.Contains(string(page3Content), "下一页") {
		t.Error("homepage page 3 should NOT have next page link")
	}
	if !strings.Contains(string(page3Content), "Post Three") {
		t.Error("homepage page 3 should show 'Post Three'")
	}

	// --- Category pagination ---
	catPage1Path := filepath.Join(public, "categories", "Demo", "index.html")
	catPage1Content, err := os.ReadFile(catPage1Path)
	if err != nil {
		t.Fatal("category page 1 (categories/Demo/index.html) was not generated:", err)
	}
	if !strings.Contains(string(catPage1Content), "1 / 3") {
		t.Error("category page 1 should show '1 / 3'")
	}

	catPage2Path := filepath.Join(public, "categories", "Demo", "page", "2", "index.html")
	if _, err := os.Stat(catPage2Path); os.IsNotExist(err) {
		t.Error("category page 2 (categories/Demo/page/2/index.html) was not generated")
	}

	// --- Tag pagination ---
	tagPage1Path := filepath.Join(public, "tags", "pagination", "index.html")
	tagPage1Content, err := os.ReadFile(tagPage1Path)
	if err != nil {
		t.Fatal("tag page 1 (tags/pagination/index.html) was not generated:", err)
	}
	if !strings.Contains(string(tagPage1Content), "1 / 3") {
		t.Error("tag page 1 should show '1 / 3'")
	}

	tagPage2Path := filepath.Join(public, "tags", "pagination", "page", "2", "index.html")
	if _, err := os.Stat(tagPage2Path); os.IsNotExist(err) {
		t.Error("tag page 2 (tags/pagination/page/2/index.html) was not generated")
	}

	// --- Verify individual post pages still exist (not paginated) ---
	for _, slug := range []string{"post-1", "post-2", "post-3"} {
		postPath := filepath.Join(public, "posts", slug, "index.html")
		if _, err := os.Stat(postPath); os.IsNotExist(err) {
			t.Errorf("individual post page (posts/%s/index.html) was not generated", slug)
		}
	}
}

func TestBuild_WithoutStaticDir(t *testing.T) {
	root := t.TempDir()

	cfgYAML := []byte(`title: "No Static Test"
baseURL: "https://example.com"
`)
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), cfgYAML, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	renderer := content.NewRenderer(cfg.Markup.Highlight.Style, cfg.Markup.Highlight.LineNumbers)
	engine, err := theme.New(cfg, root)
	if err != nil {
		t.Fatal(err)
	}

	builder := build.New(cfg, root, renderer, engine)
	if err := builder.Build(); err != nil {
		t.Fatal("build should succeed even without static/ directory:", err)
	}

	public := filepath.Join(root, "public")
	if _, err := os.Stat(filepath.Join(public, "index.html")); os.IsNotExist(err) {
		t.Error("homepage should be generated")
	}
}

func TestBuild_WithExternalTheme(t *testing.T) {
	root := t.TempDir()

	// Create config.yaml
	cfgYAML := []byte(`title: "External Theme Blog"
baseURL: "https://example.com"
seo:
  enableSitemap: true
`)
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), cfgYAML, 0644); err != nil {
		t.Fatal(err)
	}

	// Create content/posts/
	postsDir := filepath.Join(root, "content", "posts")
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		t.Fatal(err)
	}
	postMD := []byte(`---
title: "Test Post"
date: "2024-01-15"
---
This is content.
`)
	if err := os.WriteFile(filepath.Join(postsDir, "test-post.md"), postMD, 0644); err != nil {
		t.Fatal(err)
	}

	// Create external theme layout that overrides index.html's content block
	themeLayouts := filepath.Join(root, "themes", "test-theme", "layouts")
	if err := os.MkdirAll(themeLayouts, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themeLayouts, "index.html"), []byte(`{{define "content"}}<h1>Custom Theme</h1>{{end}}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Create external theme asset file
	themeAssets := filepath.Join(root, "themes", "test-theme", "assets", "css")
	if err := os.MkdirAll(themeAssets, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themeAssets, "custom.css"), []byte("body { color: red; }"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create external theme static file
	themeStatic := filepath.Join(root, "themes", "test-theme", "static")
	if err := os.MkdirAll(themeStatic, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themeStatic, "theme-favicon.ico"), []byte("custom-icon"), 0644); err != nil {
		t.Fatal(err)
	}

	// Load config and override theme
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	cfg.Theme = "test-theme"

	renderer := content.NewRenderer(cfg.Markup.Highlight.Style, cfg.Markup.Highlight.LineNumbers)
	engine, err := theme.New(cfg, root)
	if err != nil {
		t.Fatal(err)
	}

	builder := build.New(cfg, root, renderer, engine)
	if err := builder.Build(); err != nil {
		t.Fatal(err)
	}

	public := filepath.Join(root, "public")

	// --- Verify external theme template overrides index.html content ---
	homepageContent, err := os.ReadFile(filepath.Join(public, "index.html"))
	if err != nil {
		t.Fatal("homepage was not generated:", err)
	}
	if !strings.Contains(string(homepageContent), "Custom Theme") {
		t.Error("homepage should contain external theme's 'Custom Theme' content")
	}

	// --- Verify zhenhai base template still works (fallback) ---
	if !strings.Contains(string(homepageContent), "External Theme Blog") {
		t.Error("homepage should contain site title from zhenhai base template")
	}

	// --- Verify external theme assets were copied ---
	if _, err := os.Stat(filepath.Join(public, "assets", "css", "custom.css")); os.IsNotExist(err) {
		t.Error("external theme CSS (public/assets/css/custom.css) was not copied")
	}
	customCSS, err := os.ReadFile(filepath.Join(public, "assets", "css", "custom.css"))
	if err != nil {
		t.Fatal(err)
	}
	if string(customCSS) != "body { color: red; }" {
		t.Errorf("custom CSS content mismatch: got %q", string(customCSS))
	}

	// --- Verify zhenhai assets were also copied ---
	if _, err := os.Stat(filepath.Join(public, "assets", "css", "style.css")); os.IsNotExist(err) {
		t.Error("zhenhai CSS (public/assets/css/style.css) should still be copied")
	}

	// --- Verify external theme static files were copied ---
	if _, err := os.Stat(filepath.Join(public, "theme-favicon.ico")); os.IsNotExist(err) {
		t.Error("external theme static file (public/theme-favicon.ico) was not copied")
	}

	// --- Verify post page still works with zhenhai templates ---
	postPagePath := filepath.Join(public, "posts", "test-post", "index.html")
	if _, err := os.Stat(postPagePath); os.IsNotExist(err) {
		t.Error("post page should still be generated using zhenhai template")
	}
	postContent, err := os.ReadFile(postPagePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(postContent), "Test Post") {
		t.Error("post page should contain the post title")
	}
}

func TestBuild_ExternalThemeDirectoryNotFound(t *testing.T) {
	root := t.TempDir()

	// Create config.yaml
	cfgYAML := []byte(`title: "Missing Theme"
baseURL: "https://example.com"
`)
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), cfgYAML, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	// Set theme to something that doesn't exist
	cfg.Theme = "nonexistent-theme"

	renderer := content.NewRenderer(cfg.Markup.Highlight.Style, cfg.Markup.Highlight.LineNumbers)
	engine, err := theme.New(cfg, root)
	if err != nil {
		t.Fatalf("should not error when external theme dir doesn't exist: %v", err)
	}

	builder := build.New(cfg, root, renderer, engine)
	if err := builder.Build(); err != nil {
		t.Fatal(err)
	}

	public := filepath.Join(root, "public")

	// Homepage should still be generated using zhenhai fallback
	homepageContent, err := os.ReadFile(filepath.Join(public, "index.html"))
	if err != nil {
		t.Fatal("homepage was not generated:", err)
	}
	if !strings.Contains(string(homepageContent), "Missing Theme") {
		t.Error("homepage should contain the site title from zhenhai fallback")
	}

	// Theme assets should still be copied from zhenhai
	if _, err := os.Stat(filepath.Join(public, "assets", "css", "style.css")); os.IsNotExist(err) {
		t.Error("zhenhai CSS should still be copied when external theme is missing")
	}
}

func TestBuild_Incremental(t *testing.T) {
	root := t.TempDir()

	// Create config.yaml
	cfgYAML := []byte(`title: "Incremental Test"
baseURL: "https://example.com"
`)
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), cfgYAML, 0644); err != nil {
		t.Fatal(err)
	}

	// Create content with two posts
	postsDir := filepath.Join(root, "content", "posts")
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		t.Fatal(err)
	}

	post1MD := []byte(`---
title: "Post One"
date: "2024-01-15"
---
Content of post one.
`)
	if err := os.WriteFile(filepath.Join(postsDir, "post-one.md"), post1MD, 0644); err != nil {
		t.Fatal(err)
	}

	post2MD := []byte(`---
title: "Post Two"
date: "2024-01-20"
---
Content of post two.
`)
	if err := os.WriteFile(filepath.Join(postsDir, "post-two.md"), post2MD, 0644); err != nil {
		t.Fatal(err)
	}

	createBuilder := func() *build.Builder {
		cfg, err := config.Load(filepath.Join(root, "config.yaml"))
		if err != nil {
			t.Fatal(err)
		}
		renderer := content.NewRenderer(cfg.Markup.Highlight.Style, cfg.Markup.Highlight.LineNumbers)
		engine, err := theme.New(cfg, root)
		if err != nil {
			t.Fatal(err)
		}
		return build.New(cfg, root, renderer, engine)
	}

	public := filepath.Join(root, "public")
	cachePath := filepath.Join(root, ".chenhai-cache.json")

	// --- First build: full build, creates cache ---
	b1 := createBuilder()
	if err := b1.Build(); err != nil {
		t.Fatal("first build failed:", err)
	}

	// Verify cache file was created
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("cache file (.chenhai-cache.json) was not created")
	}

	// Verify all pages exist
	for _, path := range []string{
		filepath.Join(public, "posts", "post-one", "index.html"),
		filepath.Join(public, "posts", "post-two", "index.html"),
		filepath.Join(public, "index.html"),
	} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file was not generated: %s", path)
		}
	}

	// --- Second build: incremental, all pages should be skipped ---
	b2 := createBuilder()
	if err := b2.Build(); err != nil {
		t.Fatal("second build failed:", err)
	}

	// Verify pages still exist (were not deleted)
	for _, path := range []string{
		filepath.Join(public, "posts", "post-one", "index.html"),
		filepath.Join(public, "posts", "post-two", "index.html"),
	} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("page disappeared after second build: %s", path)
		}
	}

	// --- Third build: modify post-two, only that page should be rebuilt ---
	post2Modified := []byte(`---
title: "Post Two (Modified)"
date: "2024-01-20"
---
Content of post two, now modified.
`)
	if err := os.WriteFile(filepath.Join(postsDir, "post-two.md"), post2Modified, 0644); err != nil {
		t.Fatal(err)
	}

	b3 := createBuilder()
	if err := b3.Build(); err != nil {
		t.Fatal("third build failed:", err)
	}

	// Verify post-two was updated with new title
	post2Content, err := os.ReadFile(filepath.Join(public, "posts", "post-two", "index.html"))
	if err != nil {
		t.Fatal("post-two was not regenerated:", err)
	}
	if !strings.Contains(string(post2Content), "Post Two (Modified)") {
		t.Error("modified post-two should contain new title after incremental build")
	}

	// --- Fourth build: delete post-one, should be removed from public/ ---
	if err := os.Remove(filepath.Join(postsDir, "post-one.md")); err != nil {
		t.Fatal(err)
	}

	b4 := createBuilder()
	if err := b4.Build(); err != nil {
		t.Fatal("fourth build failed:", err)
	}

	// Verify post-one output was removed
	post1Path := filepath.Join(public, "posts", "post-one", "index.html")
	if _, err := os.Stat(post1Path); !os.IsNotExist(err) {
		t.Error("deleted page output should be removed from public/")
	}

	// Verify cache no longer has post-one
	cacheData, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatal("cache file should still exist:", err)
	}
	if strings.Contains(string(cacheData), "post-one") {
		t.Error("cache should not contain deleted file entry")
	}
}
