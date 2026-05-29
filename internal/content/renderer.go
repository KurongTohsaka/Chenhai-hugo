package content

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
)

type Renderer struct {
	md goldmark.Markdown
}

func NewRenderer(style string, lineNumbers bool) *Renderer {
	if style == "" {
		style = "github"
	}
	formatOptions := []chromahtml.Option{chromahtml.WithClasses(true)}
	if lineNumbers {
		formatOptions = append(formatOptions, chromahtml.WithLineNumbers(true))
	}
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,        // tables, strikethrough, task lists, autolinks
			extension.Footnote,   // footnotes
			extension.Typographer, // smart quotes, dashes, ellipses
			highlighting.NewHighlighting(
				highlighting.WithStyle(style),
				highlighting.WithFormatOptions(formatOptions...),
			),
			mathjax.NewMathJax(
				mathjax.WithInlineDelim("$", "$"),
				mathjax.WithBlockDelim("$$", "$$"),
			),
			Admonition,
			ImageEnhancer,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // allow raw HTML in markdown
		),
	)
	return &Renderer{md: md}
}

func (r *Renderer) RenderHTML(source []byte) (string, error) {
	cleaned, hlInfo := extractHLLines(source)
	var buf bytes.Buffer
	if err := r.md.Convert(cleaned, &buf); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return injectHLLines(buf.String(), hlInfo), nil
}

type TOCItem struct {
	ID    string
	Title string
	Level int
}

func (r *Renderer) RenderHTMLWithTOC(source []byte) (string, []TOCItem, error) {
	cleaned, hlInfo := extractHLLines(source)
	var buf bytes.Buffer
	if err := r.md.Convert(cleaned, &buf); err != nil {
		return "", nil, fmt.Errorf("render markdown: %w", err)
	}
	return injectHLLines(buf.String(), hlInfo), extractTOC(buf.Bytes()), nil
}

// extractTOC scans HTML for h2/h3/h4 headings with id attributes.
func extractTOC(htmlContent []byte) []TOCItem {
	s := string(htmlContent)
	var items []TOCItem
	for i := 0; i < len(s)-4; {
		if s[i] == '<' && s[i+1] == 'h' && s[i+2] >= '2' && s[i+2] <= '4' && s[i+3] == ' ' {
			level := int(s[i+2] - '0')
			j := i + 4
			id := ""
			for j < len(s) && j < i+300 {
				if strings.HasPrefix(s[j:], "id=\"") {
					j += 4
					start := j
					for j < len(s) && s[j] != '"' {
						j++
					}
					id = s[start:j]
					continue
				}
				if s[j] == '>' {
					textStart := j + 1
					endTag := fmt.Sprintf("</h%d>", level)
					if idx := strings.Index(s[textStart:], endTag); idx >= 0 {
						title := stripTags(s[textStart : textStart+idx])
						if id != "" {
							items = append(items, TOCItem{ID: id, Title: title, Level: level})
						}
					}
					break
				}
				j++
			}
			i = j
			continue
		}
		i++
	}
	return items
}

func stripTags(s string) string {
	var b []byte
	in := false
	for i := 0; i < len(s); i++ {
		if s[i] == '<' {
			in = true
			continue
		}
		if s[i] == '>' {
			in = false
			continue
		}
		if !in {
			b = append(b, s[i])
		}
	}
	return string(b)
}
