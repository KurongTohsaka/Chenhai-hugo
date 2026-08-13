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
	// Unknown shortcodes pass through verbatim (HTML-commented) for visibility.
	if !strings.Contains(html, "nosuch") {
		t.Errorf("unknown shortcode swallowed: %s", html)
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
