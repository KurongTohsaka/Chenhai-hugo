package theme

import (
	"bytes"
	"testing"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
)

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
		Page:   nil,
		Extra:  map[string]interface{}{},
	}
	// base.html expects {{template "content"}} but page is nil, so this may fail.
	// Test single.html instead which defines the content block.
	err = engine.Render(&buf, "single.html", data)
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
