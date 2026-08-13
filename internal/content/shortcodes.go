package content

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

// RegisterBuiltins registers the built-in shortcode components.
func RegisterBuiltins(reg *ShortcodeRegistry) {
	reg.Register("details", renderDetails)
	reg.Register("gallery", renderGallery)
	reg.Register("tabs", renderTabs)
}

// renderDetails renders {{< details "标题" >}}…{{< /details >}}.
func renderDetails(ctx *ShortcodeContext) (string, error) {
	title := ctx.Params["title"]
	if len(ctx.Positional) > 0 {
		title = ctx.Positional[0]
	}
	return fmt.Sprintf(
		`<details class="shortcode-details"><summary>%s</summary><div class="details-body">%s</div></details>`,
		html.EscapeString(title), ctx.ContentHTML,
	), nil
}

var mdImageRe = regexp.MustCompile(`!\[([^\]]*)\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)

// renderGallery renders {{< gallery >}} rows of ![alt](src) into a grid.
func renderGallery(ctx *ShortcodeContext) (string, error) {
	var b strings.Builder
	b.WriteString(`<div class="gallery">`)
	matches := mdImageRe.FindAllStringSubmatch(ctx.Content, -1)
	for _, m := range matches {
		fmt.Fprintf(&b, `<figure class="gallery-item"><img src="%s" alt="%s" loading="lazy"></figure>`,
			html.EscapeString(m[2]), html.EscapeString(m[1]))
	}
	b.WriteString(`</div>`)
	return b.String(), nil
}

// renderTabs renders {{< tabs "A" "B" >}} with === A === section separators.
func renderTabs(ctx *ShortcodeContext) (string, error) {
	if len(ctx.Positional) == 0 {
		return "", fmt.Errorf("tabs: 需要至少一个标签名，如 {{< tabs \"Go\" \"Java\" >}}")
	}
	sections := splitTabSections(ctx.Content)
	var b strings.Builder
	b.WriteString(`<div class="tabs" data-tabs>`)
	b.WriteString(`<div class="tab-nav" role="tablist">`)
	for i, name := range ctx.Positional {
		active := ""
		if i == 0 {
			active = " active"
		}
		fmt.Fprintf(&b, `<button type="button" class="tab-btn%s" data-tab="%d" role="tab">%s</button>`,
			active, i, html.EscapeString(name))
	}
	b.WriteString(`</div>`)
	for i, name := range ctx.Positional {
		active := ""
		if i == 0 {
			active = " active"
		}
		body := sections[name]
		if body == "" {
			body = sections[fmt.Sprintf("%d", i)] // fallback: unnamed sections by index
		}
		fmt.Fprintf(&b, `<div class="tab-panel%s" data-panel="%d" role="tabpanel">%s</div>`,
			active, i, ctx.reg.renderInnerHTML(body))
	}
	b.WriteString(`</div>`)
	return b.String(), nil
}

var tabSectionRe = regexp.MustCompile(`(?m)^=== (.+) ===\s*$`)

// splitTabSections splits content into sections keyed by their "=== name ==="
// headers. Unlabelled leading content is keyed by section index ("0", "1", …).
func splitTabSections(content string) map[string]string {
	sections := make(map[string]string)
	idx := tabSectionRe.FindAllStringSubmatchIndex(content, -1)
	if len(idx) == 0 {
		sections["0"] = content
		return sections
	}
	start := 0
	key := "0"
	for _, loc := range idx {
		if loc[0] > start {
			sections[key] = strings.TrimSpace(content[start:loc[0]])
		}
		key = content[loc[2]:loc[3]]
		start = loc[1]
	}
	if start < len(content) {
		sections[key] = strings.TrimSpace(content[start:])
	}
	return sections
}
