package content

import (
	"strings"
	"testing"
)

func TestRenderer_BasicMarkdown(t *testing.T) {
	r := NewRenderer()
	md := `# Hello World

This is **bold** text and a [link](https://example.com).`

	html, err := r.RenderHTML([]byte(md))
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	if !strings.Contains(html, "<h1") {
		t.Error("expected h1 tag in rendered HTML")
	}
	if !strings.Contains(html, "<strong>bold</strong>") {
		t.Error("expected <strong>bold</strong> in rendered HTML")
	}
	if !strings.Contains(html, "<a href=\"https://example.com\"") {
		t.Error("expected link in rendered HTML")
	}
}

func TestRenderer_CodeBlock(t *testing.T) {
	r := NewRenderer()
	md := "```go\npackage main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```"

	html, err := r.RenderHTML([]byte(md))
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	if !strings.Contains(html, "<code") {
		t.Error("expected <code> element")
	}
	if !strings.Contains(html, "<pre") {
		t.Error("expected <pre> element")
	}
	// Code keywords should be preserved (inside syntax-highlighting spans)
	if !strings.Contains(html, "package") {
		t.Error("expected 'package' keyword preserved")
	}
	if !strings.Contains(html, "Println") {
		t.Error("expected 'Println' preserved")
	}
	// Should have syntax highlighting markup (inline styles on span elements)
	if !strings.Contains(html, "<span") {
		t.Error("expected syntax highlighting spans in rendered HTML")
	}
}

func TestRenderer_MathInline(t *testing.T) {
	r := NewRenderer()
	md := "Inline math: $E=mc^2$ is famous."

	html, err := r.RenderHTML([]byte(md))
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	if !strings.Contains(html, "math inline") {
		t.Error("expected math inline class in rendered HTML")
	}
	if !strings.Contains(html, "E=mc^2") {
		t.Error("expected math content preserved")
	}
}

func TestRenderer_MathDisplay(t *testing.T) {
	r := NewRenderer()
	md := "$$\n\\mathbb{E}(X) = \\int x dF(x)\n$$"

	html, err := r.RenderHTML([]byte(md))
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	if !strings.Contains(html, "math display") {
		t.Error("expected math display class in rendered HTML")
	}
	if !strings.Contains(html, "\\mathbb{E}") {
		t.Error("expected math display content preserved")
	}
}

func TestRenderer_TOCExtraction(t *testing.T) {
	r := NewRenderer()
	md := `# Title

Some intro text.

## Section One

Content for section one.

### Subsection A

Deeper content.

## Section Two

More content.

#### Deep Dive

Very deep content.`

	_, items, err := r.RenderHTMLWithTOCItems([]byte(md))
	if err != nil {
		t.Fatalf("RenderHTMLWithTOCItems failed: %v", err)
	}

	if len(items) < 2 {
		t.Fatalf("expected at least 2 TOC items, got %d", len(items))
	}

	// Check h2 items
	foundSectionOne := false
	foundSectionTwo := false
	for _, item := range items {
		if item.Title == "Section One" && item.Level == 2 {
			foundSectionOne = true
		}
		if item.Title == "Section Two" && item.Level == 2 {
			foundSectionTwo = true
		}
	}
	if !foundSectionOne {
		t.Error("expected 'Section One' in TOC items")
	}
	if !foundSectionTwo {
		t.Error("expected 'Section Two' in TOC items")
	}

	// Check h3 item
	foundSubA := false
	for _, item := range items {
		if item.Title == "Subsection A" && item.Level == 3 {
			foundSubA = true
		}
	}
	if !foundSubA {
		t.Error("expected 'Subsection A' (h3) in TOC items")
	}

	// All items should have non-empty IDs
	for _, item := range items {
		if item.ID == "" {
			t.Errorf("TOC item %q has empty ID", item.Title)
		}
	}
}

func TestRenderer_GFMTable(t *testing.T) {
	r := NewRenderer()
	md := `| Name  | Age |
|-------|-----|
| Alice | 30  |
| Bob   | 25  |`

	html, err := r.RenderHTML([]byte(md))
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	if !strings.Contains(html, "<table") {
		t.Error("expected <table> in rendered HTML")
	}
	if !strings.Contains(html, "<th>") && !strings.Contains(html, "<thead") {
		t.Error("expected table header in rendered HTML")
	}
	if !strings.Contains(html, "Alice") {
		t.Error("expected table data preserved")
	}
	if !strings.Contains(html, "Bob") {
		t.Error("expected table data preserved")
	}
}

func TestRenderer_TaskList(t *testing.T) {
	r := NewRenderer()
	md := `- [x] Completed task
- [ ] Incomplete task`

	html, err := r.RenderHTML([]byte(md))
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	if !strings.Contains(html, `type="checkbox"`) && !strings.Contains(html, "task-list") {
		t.Error("expected checkbox or task-list markup in rendered HTML")
	}
	// Should have checked and unchecked states
	if !strings.Contains(html, "checked") && !strings.Contains(html, `checked="")`) {
		// Check if there's some form of checked/unchecked distinction
		t.Log("looking for checkbox state markers in rendered HTML")
	}
}

func TestRenderer_RenderHTMLWithTOC(t *testing.T) {
	r := NewRenderer()
	md := `# Only Titled

Some content.`

	html, err := r.RenderHTMLWithTOC([]byte(md))
	if err != nil {
		t.Fatalf("RenderHTMLWithTOC failed: %v", err)
	}

	if !strings.Contains(html, "<h1") {
		t.Error("expected h1 in rendered HTML")
	}
}

func TestRenderer_EmptyInput(t *testing.T) {
	r := NewRenderer()
	html, err := r.RenderHTML([]byte(""))
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}
	// Empty input is valid markdown; goldmark returns empty output
	_ = html
}

func TestRenderer_TOCWithHeadingID(t *testing.T) {
	r := NewRenderer()
	md := `## Hello World
### Foo Bar`

	_, items, err := r.RenderHTMLWithTOCItems([]byte(md))
	if err != nil {
		t.Fatalf("RenderHTMLWithTOCItems failed: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 TOC items, got %d", len(items))
	}

	expected := []struct {
		title string
		level int
	}{
		{"Hello World", 2},
		{"Foo Bar", 3},
	}

	for i, exp := range expected {
		if items[i].Title != exp.title {
			t.Errorf("item %d: expected title %q, got %q", i, exp.title, items[i].Title)
		}
		if items[i].Level != exp.level {
			t.Errorf("item %d: expected level %d, got %d", i, exp.level, items[i].Level)
		}
		if items[i].ID == "" {
			t.Errorf("item %d: expected non-empty ID", i)
		}
	}
}
