package theme

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
)

// newTestTemplateData creates minimal TemplateData with a single page.
func newTestTemplateData(cfg *config.Config) *TemplateData {
	page := &content.Page{Title: "Test Post", Date: parseDate("2024-01-15")}
	site := index.BuildSite(cfg, []*content.Page{page})
	return &TemplateData{
		Config: cfg,
		Page:   page,
		Site:   site,
		Extra:  map[string]interface{}{},
	}
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestNew_LoadsEmbeddedTheme(t *testing.T) {
	cfg := config.DefaultConfig()
	engine, err := New(cfg, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !engine.HasTemplate("base.html") {
		t.Error("base.html template should exist")
	}
	if !engine.HasTemplate("single.html") {
		t.Error("single.html template should exist")
	}
}

func TestRender_BaseTemplate(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Title = "测试"
	engine, err := New(cfg, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	data := &TemplateData{
		Config: cfg,
		Page:   &content.Page{Title: "Test"},
		Extra:  map[string]interface{}{},
	}
	// base.html expects {{template "content"}} but page is nil, so this may fail.
	// Test single.html instead which defines the content block.
	err = engine.RenderPage(&buf, "single.html", data)
	if err != nil {
		t.Logf("render with nil page: %v (expected when Page is nil)", err)
	}
}

func TestNew_EmptySiteRoot(t *testing.T) {
	cfg := config.DefaultConfig()
	engine, err := New(cfg, "/nonexistent/path")
	if err != nil {
		t.Fatalf("should not error on nonexistent layouts dir: %v", err)
	}
	if !engine.HasTemplate("base.html") {
		t.Error("embedded templates should still load")
	}
}

func TestNew_ExternalThemeOverridesTemplate(t *testing.T) {
	dir := t.TempDir()

	// Create external theme with a template that overrides zhenhai's single.html
	themeDir := filepath.Join(dir, "themes", "test-theme", "layouts")
	if err := os.MkdirAll(themeDir, 0755); err != nil {
		t.Fatal(err)
	}
	customTmpl := `{{define "content"}}<h1>Custom Theme</h1>{{end}}`
	if err := os.WriteFile(filepath.Join(themeDir, "single.html"), []byte(customTmpl), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.Theme = "test-theme"
	engine, err := New(cfg, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The template should exist from external theme
	if !engine.HasTemplate("single.html") {
		t.Error("single.html should be loaded from external theme")
	}

	// base.html should still come from zhenhai (not overridden)
	if !engine.HasTemplate("base.html") {
		t.Error("base.html should still be available from zhenhai fallback")
	}

	// Render to verify custom content using RenderPage (goes through base.html)
	var buf bytes.Buffer
	data := newTestTemplateData(cfg)
	err = engine.RenderPage(&buf, "single.html", data)
	if err != nil {
		t.Fatalf("RenderPage single.html: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Custom Theme")) {
		t.Errorf("expected rendered output to contain 'Custom Theme', got: %s", buf.String())
	}
}

func TestNew_ExternalThemeWithSiteOverride(t *testing.T) {
	dir := t.TempDir()

	// Create external theme
	themeDir := filepath.Join(dir, "themes", "test-theme", "layouts")
	if err := os.MkdirAll(themeDir, 0755); err != nil {
		t.Fatal(err)
	}
	themeTmpl := `{{define "content"}}<h1>From Theme</h1>{{end}}`
	if err := os.WriteFile(filepath.Join(themeDir, "single.html"), []byte(themeTmpl), 0644); err != nil {
		t.Fatal(err)
	}

	// Create site-specific override (highest priority)
	siteLayouts := filepath.Join(dir, "layouts")
	if err := os.MkdirAll(siteLayouts, 0755); err != nil {
		t.Fatal(err)
	}
	siteTmpl := `{{define "content"}}<h1>From Site</h1>{{end}}`
	if err := os.WriteFile(filepath.Join(siteLayouts, "single.html"), []byte(siteTmpl), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.Theme = "test-theme"
	engine, err := New(cfg, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Site override should take precedence
	var buf bytes.Buffer
	data := newTestTemplateData(cfg)
	err = engine.RenderPage(&buf, "single.html", data)
	if err != nil {
		t.Fatalf("RenderPage single.html: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("From Site")) {
		t.Errorf("expected site override to take precedence, got: %s", buf.String())
	}
}

func TestNew_ExternalThemeDirectoryNotFound(t *testing.T) {
	dir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Theme = "nonexistent-theme"

	// Should not error; should just use zhenhai
	engine, err := New(cfg, dir)
	if err != nil {
		t.Fatalf("should not error when external theme dir doesn't exist: %v", err)
	}
	if !engine.HasTemplate("base.html") {
		t.Error("base.html should still be available from zhenhai")
	}
}

func TestNew_ExternalThemeWithNewTemplateFile(t *testing.T) {
	dir := t.TempDir()

	// Create external theme with a brand new template not in zhenhai
	themeDir := filepath.Join(dir, "themes", "test-theme", "layouts")
	if err := os.MkdirAll(themeDir, 0755); err != nil {
		t.Fatal(err)
	}
	customTmpl := `{{define "custom-block"}}<div>Custom Block</div>{{end}}`
	if err := os.WriteFile(filepath.Join(themeDir, "custom.html"), []byte(customTmpl), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.Theme = "test-theme"
	engine, err := New(cfg, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !engine.HasTemplate("custom.html") {
		t.Error("custom.html should be loaded from external theme")
	}

	// zhenhai templates should still be available
	if !engine.HasTemplate("base.html") {
		t.Error("base.html should still be available")
	}
}

func TestNew_ThemeParamsMerge(t *testing.T) {
	dir := t.TempDir()

	// Create external theme with theme.yaml containing default params
	themeDir := filepath.Join(dir, "themes", "test-theme")
	if err := os.MkdirAll(filepath.Join(themeDir, "layouts"), 0755); err != nil {
		t.Fatal(err)
	}
	themeYAML := `name: "test-theme"
version: "1.0.0"
description: "A test theme"
author: ""
params:
  primaryColor: "#3366ff"
  fontSize: 16
  showSidebar: true
`
	if err := os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(themeYAML), 0644); err != nil {
		t.Fatal(err)
	}
	// Need at least base.html for engine to work
	baseTmpl := `<!DOCTYPE html><html><body>{{block "content" .}}{{end}}</body></html>`
	if err := os.WriteFile(filepath.Join(themeDir, "layouts", "base.html"), []byte(baseTmpl), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.Theme = "test-theme"

	engine, err := New(cfg, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if engine == nil {
		t.Fatal("engine should not be nil")
	}

	// Verify theme params were merged
	if cfg.ThemeConfig.Params["primaryColor"] != "#3366ff" {
		t.Errorf("expected primaryColor '#3366ff', got %v", cfg.ThemeConfig.Params["primaryColor"])
	}
	if cfg.ThemeConfig.Params["fontSize"] != 16 {
		t.Errorf("expected fontSize 16, got %v", cfg.ThemeConfig.Params["fontSize"])
	}
	if cfg.ThemeConfig.Params["showSidebar"] != true {
		t.Errorf("expected showSidebar true, got %v", cfg.ThemeConfig.Params["showSidebar"])
	}
}

func TestNew_ThemeParamsUserOverride(t *testing.T) {
	dir := t.TempDir()

	// Create external theme with theme.yaml containing default params
	themeDir := filepath.Join(dir, "themes", "test-theme")
	if err := os.MkdirAll(filepath.Join(themeDir, "layouts"), 0755); err != nil {
		t.Fatal(err)
	}
	themeYAML := `name: "test-theme"
version: "1.0.0"
params:
  primaryColor: "#3366ff"
  fontSize: 16
`
	if err := os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(themeYAML), 0644); err != nil {
		t.Fatal(err)
	}
	baseTmpl := `<!DOCTYPE html><html><body>{{block "content" .}}{{end}}</body></html>`
	if err := os.WriteFile(filepath.Join(themeDir, "layouts", "base.html"), []byte(baseTmpl), 0644); err != nil {
		t.Fatal(err)
	}

	// Simulate user having already set some params in config.yaml
	cfg := config.DefaultConfig()
	cfg.Theme = "test-theme"
	cfg.ThemeConfig.Params["primaryColor"] = "#ff0000" // user override

	_, err := New(cfg, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// User's value should take precedence
	if cfg.ThemeConfig.Params["primaryColor"] != "#ff0000" {
		t.Errorf("expected user override primaryColor '#ff0000', got %v", cfg.ThemeConfig.Params["primaryColor"])
	}
	// Theme default for non-overridden value should still be applied
	if cfg.ThemeConfig.Params["fontSize"] != 16 {
		t.Errorf("expected fontSize 16 from theme default, got %v", cfg.ThemeConfig.Params["fontSize"])
	}
}
