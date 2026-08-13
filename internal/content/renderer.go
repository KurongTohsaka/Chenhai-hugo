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
	md  goldmark.Markdown
	reg *ShortcodeRegistry
}

func NewRenderer(style string, lineNumbers bool) *Renderer {
	if style == "" {
		style = "github"
	}
	formatOptions := []chromahtml.Option{chromahtml.WithClasses(true)}
	if lineNumbers {
		formatOptions = append(formatOptions, chromahtml.WithLineNumbers(true))
	}
	exts := []goldmark.Extender{
		extension.GFM,         // tables, strikethrough, task lists, autolinks
		extension.Footnote,    // footnotes
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
	}
	// Build inner markdown (full extensions minus shortcode) for shortcode content.
	innerMD := goldmark.New(
		goldmark.WithExtensions(exts...),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)

	reg := NewShortcodeRegistry(innerMD)
	RegisterBuiltins(reg) // 内置组件 details/gallery/tabs（Task 5）
	// goldmark v1.8.2 的 Markdown 接口无 AddOptions——shortcode 扩展随主实例
	// 构造时一并注册。
	mdExts := make([]goldmark.Extender, 0, len(exts)+1)
	mdExts = append(mdExts, exts...)
	mdExts = append(mdExts, &shortcodeExt{reg: reg})
	md := goldmark.New(
		goldmark.WithExtensions(mdExts...),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // allow raw HTML in markdown
		),
	)
	return &Renderer{md: md, reg: reg}
}

func (r *Renderer) RenderHTML(source []byte) (string, error) {
	// Y1: shortcode 块从提取源中剔除（保留行数），其内部代码块不参与
	// 顶层 lang/hl 后处理；Convert 仍用完整源（goldmark 负责 shortcode 解析）。
	stripped := stripShortcodeBlocks(source)
	langs := extractLangs(stripped)
	_, hlInfo := extractHLLines(stripped) // 只提取正文标记
	cleaned, _ := extractHLLines(source)  // 仅删除 {hl_lines} 标记供 Convert

	r.reg.BeginRender()
	var buf bytes.Buffer
	if err := r.md.Convert(cleaned, &buf); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	result := injectHLLines(buf.String(), hlInfo)
	result = injectLangLabels(result, langs)
	result = splitLineNumbers(result)
	return r.reg.ReplacePlaceholders(result), nil
}

type TOCItem struct {
	ID    string
	Title string
	Level int
}

func (r *Renderer) RenderHTMLWithTOC(source []byte) (string, []TOCItem, error) {
	// Y1: 与 RenderHTML 同款占位符管线（见 RenderHTML 注释）。
	stripped := stripShortcodeBlocks(source)
	langs := extractLangs(stripped)
	_, hlInfo := extractHLLines(stripped) // 只提取正文标记
	cleaned, _ := extractHLLines(source)  // 仅删除 {hl_lines} 标记供 Convert

	r.reg.BeginRender()
	var buf bytes.Buffer
	if err := r.md.Convert(cleaned, &buf); err != nil {
		return "", nil, fmt.Errorf("render markdown: %w", err)
	}
	result := injectHLLines(buf.String(), hlInfo)
	result = injectLangLabels(result, langs)
	result = splitLineNumbers(result)
	result = r.reg.ReplacePlaceholders(result)
	return result, extractTOC(buf.Bytes()), nil
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
