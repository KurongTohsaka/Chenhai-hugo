package content

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type imageExt struct{}

// ImageEnhancer is a Goldmark extension that wraps <img> in <figure> with
// figcaption, supports alignment markers in alt text, and adds lazy loading.
var ImageEnhancer = &imageExt{}

func (e *imageExt) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&imageRenderer{}, 100),
	))
}

type imageRenderer struct{}

func (r *imageRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, r.renderImage)
}

func (r *imageRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	img := node.(*ast.Image)

	alt := extractAlt(img, source)
	title := string(img.Title)
	src := string(img.Destination)

	// Parse alignment from alt text: "text |center" or "text |right"
	align := ""
	cleanAlt := alt
	if idx := strings.LastIndex(alt, "|"); idx >= 0 {
		suffix := strings.TrimSpace(alt[idx+1:])
		if suffix == "center" || suffix == "right" {
			align = suffix
			cleanAlt = strings.TrimSpace(alt[:idx])
		}
	}

	// Determine caption: use title if present, fallback to cleaned alt
	caption := title
	if caption == "" {
		caption = cleanAlt
	}

	// Build figure wrapper
	figClass := "image-figure"
	if align != "" {
		figClass += " image-" + align
	}

	_, _ = w.WriteString(`<figure class="` + figClass + `">`)

	// Build <img> tag
	imgTag := `<img src="` + src + `" alt="` + escapeHTML(cleanAlt) + `"`
	if title != "" {
		imgTag += ` title="` + escapeHTML(title) + `"`
	}
	imgTag += ` loading="lazy">`
	_, _ = w.WriteString(imgTag)

	// Caption
	if caption != "" {
		_, _ = w.WriteString(`<figcaption>` + escapeHTML(caption) + `</figcaption>`)
	}

	_, _ = w.WriteString(`</figure>`)

	return ast.WalkSkipChildren, nil
}

// extractAlt collects the text content from an Image node's children.
func extractAlt(img *ast.Image, source []byte) string {
	var buf bytes.Buffer
	collectText(img, source, &buf)
	return buf.String()
}

// collectText recursively collects text from inline nodes.
func collectText(n ast.Node, source []byte, buf *bytes.Buffer) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch t := c.(type) {
		case *ast.Text:
			buf.Write(t.Segment.Value(source))
		case *ast.String:
			buf.Write(t.Value)
		default:
			collectText(c, source, buf)
		}
	}
}

// escapeHTML escapes HTML special characters: & < > " '
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
