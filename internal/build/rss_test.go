package build

import (
	"strings"
	"testing"
	"time"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
)

func TestGenerateRSS_Basic(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Title = "Test Blog"
	cfg.BaseURL = "https://example.com"
	cfg.Author.Name = "Kurong"

	pages := []*content.Page{
		{Title: "Newest", Date: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), RelPath: "posts/new.md", Section: "posts", Content: "<p>new</p>", Description: "desc"},
		{Title: "Draft", Date: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), RelPath: "posts/draft.md", Section: "posts", Content: "<p>d</p>", Draft: true},
		{Title: "Old", Date: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), RelPath: "posts/old.md", Section: "posts", Content: "<p>old</p>"},
	}

	xml, err := generateRSS(cfg, pages)
	if err != nil {
		t.Fatal(err)
	}
	s := string(xml)
	if !strings.Contains(s, `<feed xmlns="http://www.w3.org/2005/Atom">`) {
		t.Errorf("missing feed root: %s", s)
	}
	if !strings.Contains(s, "<title>Test Blog</title>") {
		t.Errorf("missing title: %s", s)
	}
	if !strings.Contains(s, "https://example.com/posts/new/") {
		t.Errorf("missing entry link: %s", s)
	}
	if strings.Contains(s, "draft.md") || strings.Contains(s, "Draft</title>") {
		t.Errorf("draft leaked into feed: %s", s)
	}
	// newest first
	if strings.Index(s, "Newest") > strings.Index(s, "Old") {
		t.Errorf("entries not newest-first: %s", s)
	}
	// summary present（v0.8 统一摘要：Description 优先）
	if !strings.Contains(s, "desc") {
		t.Errorf("entry summary missing: %s", s)
	}
}

// Y2: Description 为空时回退正文截断摘要。
func TestGenerateRSS_SummaryFallback(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.BaseURL = "https://example.com"
	pages := []*content.Page{{
		Title: "x", Date: time.Now(),
		RelPath: "posts/x.md", Section: "posts",
		Content: "<p>这是正文内容，用于摘要截断测试</p>",
	}}
	xml, err := generateRSS(cfg, pages)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(xml), "摘要截断测试") {
		t.Errorf("summary fallback missing: %s", xml)
	}
}

// Y3: XML 1.0 非法控制字符必须被清洗，否则单篇文章即可让整站构建失败。
func TestGenerateRSS_XMLClean(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.BaseURL = "https://example.com"
	pages := []*content.Page{{
		Title: "x\x01y", Date: time.Now(),
		RelPath: "posts/x.md", Section: "posts",
		Description: "含\x0b非法字符",
	}}
	xml, err := generateRSS(cfg, pages)
	if err != nil {
		t.Fatal(err)
	}
	s := string(xml)
	if strings.ContainsRune(s, '\x01') || strings.ContainsRune(s, '\x0b') {
		t.Errorf("illegal XML characters leaked: %s", s)
	}
}

func TestGenerateRSS_NoBaseURL(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.BaseURL = ""
	pages := []*content.Page{{Title: "x", Date: time.Now(), RelPath: "posts/x.md", Section: "posts"}}
	xml, err := generateRSS(cfg, pages)
	if err != nil {
		t.Fatal(err)
	}
	if len(xml) != 0 {
		t.Errorf("expected empty output without baseURL, got %d bytes", len(xml))
	}
}

func TestGenerateRSS_Limit(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.BaseURL = "https://example.com"
	cfg.RSS.Limit = 2
	pages := []*content.Page{}
	for i := 1; i <= 5; i++ {
		pages = append(pages, &content.Page{
			Title: "p" + string(rune('0'+i)),
			Date:  time.Date(2026, 8, i, 0, 0, 0, 0, time.UTC),
			RelPath: "posts/p.md", Section: "posts",
		})
	}
	xml, _ := generateRSS(cfg, pages)
	if n := strings.Count(string(xml), "<entry>"); n != 2 {
		t.Errorf("entry count = %d, want 2", n)
	}
}
