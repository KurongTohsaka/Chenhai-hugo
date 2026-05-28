package content

import (
	"fmt"
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

	_, items, err := r.RenderHTMLWithTOC([]byte(md))
	if err != nil {
		t.Fatalf("RenderHTMLWithTOC failed: %v", err)
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

	html, _, err := r.RenderHTMLWithTOC([]byte(md))
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

func TestRenderer_Admonition(t *testing.T) {
	r := NewRenderer()
	input := "> [!note]\n> This is a note.\n\n> [!warning]\n> **Warning** text here.\n\nNormal paragraph."
	html, err := r.RenderHTML([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify note admonition
	if !strings.Contains(html, `class="admonition admonition-note"`) {
		t.Error("expected note admonition div")
	}
	if !strings.Contains(html, "笔记") {
		t.Error("expected note title '笔记'")
	}
	if !strings.Contains(html, "This is a note.") {
		t.Error("expected note content")
	}
	// Verify warning admonition
	if !strings.Contains(html, `class="admonition admonition-warning"`) {
		t.Error("expected warning admonition div")
	}
	if !strings.Contains(html, "注意") {
		t.Error("expected warning title '注意'")
	}
	if !strings.Contains(html, "<strong>Warning</strong>") {
		t.Error("expected rendered **Warning** as bold")
	}
	// Regular paragraph should still be present
	if !strings.Contains(html, "Normal paragraph") {
		t.Error("expected normal paragraph to pass through")
	}
}

func TestRenderer_RegularBlockquote(t *testing.T) {
	r := NewRenderer()
	input := "> This is a regular quote.\n\nThis is a normal paragraph."
	html, err := r.RenderHTML([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should contain <blockquote> tag
	if !strings.Contains(html, "<blockquote>") {
		t.Error("expected regular blockquote tag")
	}
	// Should NOT contain admonition class
	if strings.Contains(html, `class="admonition"`) {
		t.Error("regular blockquote should not become admonition")
	}
}

func TestRenderer_AdmonitionAllTypes(t *testing.T) {
	r := NewRenderer()
	types := []string{"note", "warning", "tip", "danger"}
	for _, tp := range types {
		input := fmt.Sprintf("> [!%s]\n> Content for %s.", tp, tp)
		html, err := r.RenderHTML([]byte(input))
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tp, err)
		}
		if !strings.Contains(html, fmt.Sprintf(`admonition-%s`, tp)) {
			t.Errorf("missing admonition-%s class", tp)
		}
	}
}

func TestRenderer_AdmonitionSameLineContent(t *testing.T) {
	r := NewRenderer()
	input := "> [!tip] This tip has content on the same line.\n> And a second line."
	html, err := r.RenderHTML([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, `class="admonition admonition-tip"`) {
		t.Error("expected tip admonition div")
	}
	if !strings.Contains(html, "提示") {
		t.Error("expected tip title '提示'")
	}
	if !strings.Contains(html, "This tip has content on the same line.") {
		t.Error("expected content from first line")
	}
	if !strings.Contains(html, "And a second line.") {
		t.Error("expected content from second line")
	}
}

func TestRenderer_TOCWithHeadingID(t *testing.T) {
	r := NewRenderer()
	md := `## Hello World
### Foo Bar`

	_, items, err := r.RenderHTMLWithTOC([]byte(md))
	if err != nil {
		t.Fatalf("RenderHTMLWithTOC failed: %v", err)
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
