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

// Y11: config.yaml 语法损坏时 doctor 不得 panic（baseURL 检查仅在 Load 成功分支内执行）。
func TestDoctor_BadConfigNoPanic(t *testing.T) {
	root := t.TempDir()

	badYAML := []byte("title: [unclosed\nbaseURL: \"https://example.com\"\n")
	os.WriteFile(filepath.Join(root, "config.yaml"), badYAML, 0644)

	origDir, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(origDir)

	err := cli.ExecuteDoctor(root)
	if err == nil {
		t.Error("expected error for bad config.yaml")
	}
}

// baseURL 缺失时 doctor 给出 RSS 警告（不视为错误）。
func TestDoctor_MissingBaseURLWarns(t *testing.T) {
	root := t.TempDir()

	cfgYAML := []byte(`title: "Test Site"
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
		t.Errorf("unexpected error: %v", err)
	}
}
