package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCmd_Basic(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "content"), 0755)

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(dir)

	newCategory = ""
	newTags = ""
	if err := newCmd.RunE(newCmd, []string{"posts/hello-world.md"}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "content", "posts", "hello-world.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `title: "hello-world"`) {
		t.Errorf("expected title in front matter, got:\n%s", content)
	}
	if !strings.Contains(content, "draft: false") {
		t.Errorf("expected draft: false")
	}
}

func TestNewCmd_WithCategoryAndTags(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "content"), 0755)

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(dir)

	newCategory = "技术, Go"
	newTags = "hugo, blog"
	defer func() { newCategory = ""; newTags = "" }()

	if err := newCmd.RunE(newCmd, []string{"posts/test.md"}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "content", "posts", "test.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `categories: ["技术", "Go"]`) {
		t.Errorf("expected categories, got:\n%s", content)
	}
	if !strings.Contains(content, `tags: ["hugo", "blog"]`) {
		t.Errorf("expected tags, got:\n%s", content)
	}
}

func TestNewCmd_DuplicateFile(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "content", "posts"), 0755)
	os.WriteFile(filepath.Join(dir, "content", "posts", "dup.md"), []byte("x"), 0644)

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(dir)

	err := newCmd.RunE(newCmd, []string{"posts/dup.md"})
	if err == nil {
		t.Fatal("expected error for duplicate file")
	}
}
