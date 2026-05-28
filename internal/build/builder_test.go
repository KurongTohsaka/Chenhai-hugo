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
	renderer := content.NewRenderer()

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

	renderer := content.NewRenderer()
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

	renderer := content.NewRenderer()
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

	renderer := content.NewRenderer()
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

	renderer := content.NewRenderer()
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
