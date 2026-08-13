package content

import (
	"strings"
	"testing"
)

func TestRenderShortcode_Unknown(t *testing.T) {
	r := NewRenderer("github", false)
	src := "{{< nosuch >}}x{{< /nosuch >}}\n"
	html, err := r.RenderHTML([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	// Unknown shortcodes pass through verbatim, marked with an HTML comment
	// so the author can spot it (a bare "nosuch" check cannot distinguish
	// recognized+passthrough from unrecognized).
	if !strings.Contains(html, "<!-- unknown shortcode: nosuch -->") {
		t.Errorf("unknown shortcode missing comment marker: %s", html)
	}
}

func TestRenderShortcode_Unclosed(t *testing.T) {
	r := NewRenderer("github", false)
	src := "{{< details \"答案\" >}}没有闭合\n"
	html, err := r.RenderHTML([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "{{<") {
		t.Errorf("unclosed shortcode should pass through: %s", html)
	}
}

func TestParseShortcodeParams(t *testing.T) {
	params, positional := parseShortcodeParams(`key="v 1" bare k2=v2`)
	if params["key"] != "v 1" || params["k2"] != "v2" {
		t.Errorf("params = %v", params)
	}
	if len(positional) != 1 || positional[0] != "bare" {
		t.Errorf("positional = %v", positional)
	}
}

func TestRenderShortcode_InsideCodeBlock(t *testing.T) {
	r := NewRenderer("github", false)
	src := "```text\n{{< details \"x\" >}}\n```\n"
	html, err := r.RenderHTML([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	// Code fence content must NOT be parsed as shortcode.
	if strings.Contains(html, `<details class=`) {
		t.Errorf("shortcode parsed inside code block: %s", html)
	}
}

func TestRenderShortcode_Gallery(t *testing.T) {
	r := NewRenderer("github", false)
	src := "{{< gallery >}}\n![图一](/img/a.png)\n![图二](/img/b.png \"B\")\n{{< /gallery >}}\n"
	html, err := r.RenderHTML([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `<div class="gallery">`) {
		t.Errorf("missing gallery wrapper: %s", html)
	}
	if !strings.Contains(html, `/img/a.png`) || !strings.Contains(html, `/img/b.png`) {
		t.Errorf("gallery images missing: %s", html)
	}
}

func TestRenderShortcode_Tabs(t *testing.T) {
	r := NewRenderer("github", false)
	src := "{{< tabs \"Go\" \"Java\" >}}\n=== Go ===\n```go\nfmt.Println(\"hi\")\n```\n=== Java ===\n```java\nSystem.out.println(\"hi\");\n```\n{{< /tabs >}}\n"
	html, err := r.RenderHTML([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `class="tabs"`) {
		t.Errorf("missing tabs wrapper: %s", html)
	}
	if !strings.Contains(html, `data-tab="0"`) || !strings.Contains(html, `data-tab="1"`) {
		t.Errorf("tab buttons missing: %s", html)
	}
	if !strings.Contains(html, "<span class=\"language-go\">") || !strings.Contains(html, "language-go") {
		// chroma emits language class on <div class="highlight">...<code class="language-go">
		if !strings.Contains(html, "language-go") {
			t.Errorf("inner code not rendered with language: %s", html)
		}
	}
	if !strings.Contains(html, `class="tab-panel active"`) {
		t.Errorf("first panel not active: %s", html)
	}
}

func TestRenderShortcode_Details(t *testing.T) {
	r := NewRenderer("github", false)
	src := "前言\n\n{{< details \"答案\" >}}折叠的**内容**{{< /details >}}\n\n结尾\n"
	html, err := r.RenderHTML([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `<details class="shortcode-details">`) {
		t.Errorf("missing details wrapper: %s", html)
	}
	if !strings.Contains(html, "<summary>答案</summary>") {
		t.Errorf("missing summary: %s", html)
	}
	if !strings.Contains(html, "<strong>内容</strong>") {
		t.Errorf("inner markdown not rendered: %s", html)
	}
	// Regression (verify 🔴): the input is the single-line inline form with
	// following content — "结尾" must stay OUTSIDE the details block. Before
	// the Closed fix it was swallowed into Content and rendered inside.
	if d := strings.Index(html, "</details>"); d < 0 {
		t.Errorf("missing details close tag: %s", html)
	} else if j := strings.Index(html, "结尾"); j < 0 || j < d {
		t.Errorf("content after inline shortcode swallowed into details: %s", html)
	}
}

// Regression (verify 🔴): single-line inline form {{< name >}}rest{{< /name >}}
// closes the block on the opening line; following document content must not
// be swallowed into Content.
func TestRenderShortcode_InlineFollowedByContent(t *testing.T) {
	r := NewRenderer("github", false)
	src := "{{< details \"答案\" >}}折叠**内容**{{< /details >}}\n\n结尾\n"
	html, err := r.RenderHTML([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `<details class="shortcode-details">`) {
		t.Errorf("inline details not rendered: %s", html)
	}
	if !strings.Contains(html, "<strong>内容</strong>") {
		t.Errorf("inline inner markdown not rendered: %s", html)
	}
	// The details body must contain ONLY the inline content; "结尾" must be
	// outside the block (after </details>).
	d := strings.Index(html, "</details>")
	j := strings.Index(html, "结尾")
	if d < 0 {
		t.Errorf("missing details close tag: %s", html)
	} else if j < 0 || j < d {
		t.Errorf("following content swallowed into inline shortcode: %s", html)
	}
}

// Regression (verify 🟡): v0.8 has no nested shortcodes — an unclosed block
// meeting a new {{< open >}} line terminates there (with the unclosed marker,
// not silently); the new open starts its own block and must not absorb the
// earlier block's content.
func TestRenderShortcode_UnclosedBeforeNewOpen(t *testing.T) {
	r := NewRenderer("github", false)
	src := "{{< details \"甲\" >}}\n内容甲\n{{< details \"乙\" >}}\n内容乙\n{{< /details >}}\n"
	html, err := r.RenderHTML([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	// First block is unclosed → HTML comment marker + verbatim passthrough.
	if !strings.Contains(html, "<!-- unclosed shortcode: details -->") {
		t.Errorf("missing unclosed marker: %s", html)
	}
	// Second block renders normally with its own title.
	if !strings.Contains(html, "<summary>乙</summary>") {
		t.Errorf("second details not rendered: %s", html)
	}
	// 内容甲 must not leak into the second details body.
	bodyStart := strings.Index(html, "<summary>乙</summary>")
	bodyEnd := strings.Index(html, "</details>")
	if bodyStart < 0 || bodyEnd < 0 {
		t.Fatalf("missing second details markers: %s", html)
	}
	if body := html[bodyStart:bodyEnd]; strings.Contains(body, "内容甲") {
		t.Errorf("first block content leaked into second details: %s", html)
	}
}

// Y1 regression: shortcode-internal code blocks (before top-level ones in
// source order) must not steal lang labels or hl_lines markers.
func TestRenderShortcode_TabsAndHLLines(t *testing.T) {
	r := NewRenderer("github", false)
	src := "{{< tabs \"Go\" >}}\n=== Go ===\n```go\nfmt.Println(1)\n```\n{{< /tabs >}}\n\n正文\n\n```python {hl_lines=[1]}\nx = 1\ny = 2\n```\n"
	html, err := r.RenderHTML([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	// top-level python block keeps its hl_lines highlight
	// (static form is the data-hl-lines attribute injected by injectHLLines;
	// the .hl class is added client-side by main.js initHighlightLines)
	if !strings.Contains(html, `data-hl-lines="1"`) {
		t.Errorf("top-level hl_lines lost: %s", html)
	}
	// tabs structure intact and go block inside
	if !strings.Contains(html, `class="tabs"`) || !strings.Contains(html, "language-go") {
		t.Errorf("tabs/go block missing: %s", html)
	}
}
