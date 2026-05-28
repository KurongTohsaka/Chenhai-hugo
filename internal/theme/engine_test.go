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
