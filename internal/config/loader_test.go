package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Theme != "zhenhai" {
		t.Errorf("expected theme 'zhenhai', got %q", cfg.Theme)
	}
	if cfg.ThemeConfig.PostsPerPage != 10 {
		t.Errorf("expected PostsPerPage 10, got %d", cfg.ThemeConfig.PostsPerPage)
	}
}

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("title: \"测试\"\nthemeConfig:\n  postsPerPage: 5\n"), 0644)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Title != "测试" {
		t.Errorf("expected Title '测试', got %q", cfg.Title)
	}
	if cfg.ThemeConfig.PostsPerPage != 5 {
		t.Errorf("PostsPerPage override failed: got %d", cfg.ThemeConfig.PostsPerPage)
	}
	if cfg.Theme != "zhenhai" {
		t.Error("default Theme should persist after merge")
	}
}

func TestLoad_CustomThemeParams(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	yaml := "title: \"Test\"\nthemeConfig:\n  colorMode: dark\n  primaryColor: \"#ff0000\"\n  fontSize: 16\n"
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Known fields should map correctly
	if cfg.ThemeConfig.ColorMode != "dark" {
		t.Errorf("expected ColorMode 'dark', got %q", cfg.ThemeConfig.ColorMode)
	}
	// Custom params should be captured in Params map via yaml:",inline"
	if cfg.ThemeConfig.Params["primaryColor"] != "#ff0000" {
		t.Errorf("expected Params[primaryColor] '#ff0000', got %v", cfg.ThemeConfig.Params["primaryColor"])
	}
	if cfg.ThemeConfig.Params["fontSize"] != 16 {
		t.Errorf("expected Params[fontSize] 16, got %v", cfg.ThemeConfig.Params["fontSize"])
	}
}
