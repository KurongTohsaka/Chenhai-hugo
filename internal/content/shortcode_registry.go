package content

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
)

// ShortcodeTemplateProvider lets a theme supply HTML templates for shortcodes
// via layouts/shortcodes/<name>.html. Implemented by theme.Engine.
type ShortcodeTemplateProvider interface {
	LookupShortcodeTemplate(name string) ([]byte, bool)
}

// ShortcodeContext is passed to registered shortcode render functions.
type ShortcodeContext struct {
	Name        string
	Params      map[string]string
	Positional  []string
	Content     string // raw inner markdown (before inner rendering)
	ContentHTML string // inner markdown rendered to HTML
	Page        *Page  // reserved for v0.9+; nil in v0.8
	reg         *ShortcodeRegistry // for innerHTML rendering (renderInnerHTML)
}

// ShortcodeFunc renders a shortcode to HTML.
type ShortcodeFunc func(ctx *ShortcodeContext) (string, error)

// ShortcodeRegistry maps shortcode names to render functions.
type ShortcodeRegistry struct {
	funcs        map[string]ShortcodeFunc
	innerMD      goldmark.Markdown
	provider     ShortcodeTemplateProvider
	placeholders []string // rendered shortcode HTML buffered during one RenderHTML call
}

// 注：RenderHTML 在内容 collect 阶段串行调用（collectPages 单线程遍历），
// placeholders 无并发竞争；若未来渲染调用并发化，需加锁保护。

// NewShortcodeRegistry creates a registry; inner is the goldmark instance used
// to render shortcode inner content (must NOT contain the shortcode parser).
func NewShortcodeRegistry(inner goldmark.Markdown) *ShortcodeRegistry {
	reg := &ShortcodeRegistry{
		funcs:   make(map[string]ShortcodeFunc),
		innerMD: inner,
	}
	// 内置组件（details/gallery/tabs）由 shortcodes.go 的 RegisterBuiltins 注册，
	// 该文件随 Task 5 落地；Task 4 阶段 registry 为空表（未知名称走透传兜底）。
	return reg
}

// Register adds a shortcode render function.
func (r *ShortcodeRegistry) Register(name string, fn ShortcodeFunc) {
	r.funcs[name] = fn
}

// SetTemplateProvider installs the theme template provider (theme.Engine).
func (r *ShortcodeRegistry) SetTemplateProvider(p ShortcodeTemplateProvider) {
	r.provider = p
}

// BeginRender resets the placeholder buffer; call at the start of each
// RenderHTML / RenderHTMLWithTOC invocation.
func (r *ShortcodeRegistry) BeginRender() { r.placeholders = r.placeholders[:0] }

// WritePlaceholder writes a \x00sc<N>\x00 token into the HTML stream and
// buffers the real HTML for replacement after post-processing.
func (r *ShortcodeRegistry) WritePlaceholder(w io.Writer, html string) {
	r.placeholders = append(r.placeholders, html)
	_, _ = fmt.Fprintf(w, "\x00sc%d\x00", len(r.placeholders)-1)
}

// ReplacePlaceholders swaps buffered shortcode HTML back in.
func (r *ShortcodeRegistry) ReplacePlaceholders(s string) string {
	for i, h := range r.placeholders {
		s = strings.ReplaceAll(s, fmt.Sprintf("\x00sc%d\x00", i), h)
	}
	return s
}

// Render renders a shortcode; returns ok=false when the name is unknown.
func (r *ShortcodeRegistry) Render(name, rawParams, content string, inner goldmark.Markdown) (string, bool) {
	fn, ok := r.funcs[name]
	if !ok {
		return "", false
	}
	params, positional := parseShortcodeParams(rawParams)
	innerHTML := ""
	if inner != nil {
		var buf strings.Builder
		if err := inner.Convert([]byte(content), &buf); err == nil {
			innerHTML = buf.String()
		} else {
			innerHTML = "<p>" + html.EscapeString(content) + "</p>"
		}
	}
	ctx := &ShortcodeContext{
		Name:        name,
		Params:      params,
		Positional:  positional,
		Content:     content,
		ContentHTML: innerHTML,
		reg:         r,
	}

	// Theme template override (layouts/shortcodes/<name>.html) takes
	// precedence. Rendered with a STANDALONE html/template instance — never
	// parsed into the global template set (Execute then Parse would panic,
	// same trap RenderPage avoids via Clone). Default context escaping is
	// safe; theme authors use {{.ContentHTML | safeHTML}} for raw HTML.
	if r.provider != nil {
		if tmplSrc, ok := r.provider.LookupShortcodeTemplate(name); ok {
			tmpl, err := template.New("shortcode-" + name).Funcs(template.FuncMap{
				"safeHTML": func(s string) template.HTML { return template.HTML(s) },
			}).Parse(string(tmplSrc))
			if err == nil {
				var buf bytes.Buffer
				err := tmpl.Execute(&buf, map[string]interface{}{
					"Params":      ctx.Params,
					"Positional":  ctx.Positional,
					"Content":     ctx.Content,
					"ContentHTML": ctx.ContentHTML,
					"Page":        ctx.Page,
				})
				if err == nil {
					return buf.String(), true
				}
			}
			// 模板解析/执行失败：回退内置渲染器（不阻断构建）
		}
	}

	out, err := fn(ctx)
	if err != nil {
		return "", true // fn error: swallow, fall back to pass-through below
	}
	return out, true
}

var paramTokenRe = regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9-]*)\s*=\s*"([^"]*)"|([a-zA-Z][a-zA-Z0-9-]*)\s*=\s*([^\s"]+)|("[^"]*"|[^\s"]+)`)

// parseShortcodeParams splits a raw param string into key=value pairs and
// positional tokens. Quoted values may contain spaces.
func parseShortcodeParams(raw string) (map[string]string, []string) {
	params := make(map[string]string)
	var positional []string
	for _, m := range paramTokenRe.FindAllStringSubmatch(raw, -1) {
		switch {
		case m[1] != "": // key="value"
			params[m[1]] = m[2]
		case m[3] != "": // key=value
			params[m[3]] = m[4]
		case m[5] != "": // positional (possibly quoted)
			positional = append(positional, strings.Trim(m[5], `"`))
		}
	}
	return params, positional
}
