package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KurongTohsaka/chenhai-hugo/internal/cli"
)

func TestDoctor_NoConfigFile(t *testing.T) {
	root := t.TempDir()

	origDir, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(origDir)

	err := cli.ExecuteDoctor(root)
	if err == nil {
		t.Error("expected error for missing config.yaml")
	}
}

func TestDoctor_ValidSite(t *testing.T) {
	root := t.TempDir()

	// Create minimal valid site
	cfgYAML := []byte(`title: "Test Site"
baseURL: "https://example.com"
`)
	os.WriteFile(filepath.Join(root, "config.yaml"), cfgYAML, 0644)

	contentDir := filepath.Join(root, "content", "posts")
	os.MkdirAll(contentDir, 0755)
	postMD := []byte(`---
title: "Test Post"
date: "2024-01-15"
---
Body content.
`)
	os.WriteFile(filepath.Join(contentDir, "test.md"), postMD, 0644)

	origDir, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(origDir)

	err := cli.ExecuteDoctor(root)
	if err != nil {
		t.Errorf("unexpected error for valid site: %v", err)
	}
}

func TestDoctor_BadFrontMatter(t *testing.T) {
	root := t.TempDir()

	cfgYAML := []byte(`title: "Test Site"`)
	os.WriteFile(filepath.Join(root, "config.yaml"), cfgYAML, 0644)

	contentDir := filepath.Join(root, "content")
	os.MkdirAll(contentDir, 0755)
	badMD := []byte(`---
title: Broken
date: not-a-valid-date
---
Body.
`)
	os.WriteFile(filepath.Join(contentDir, "bad.md"), badMD, 0644)

	origDir, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(origDir)

	err := cli.ExecuteDoctor(root)
	if err == nil {
		t.Error("expected error for bad front matter")
	}
}
