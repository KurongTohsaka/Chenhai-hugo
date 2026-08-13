package content

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// kindShortcode is the NodeKind of ShortcodeNode.
var kindShortcode = ast.NewNodeKind("Shortcode")

// ShortcodeNode holds a block-level shortcode: {{< name params >}}…{{< /name >}}.
type ShortcodeNode struct {
	ast.BaseBlock
	Name    string
	Params  string // raw param string between name and ">}}"
	Content string // raw inner lines joined with \n
	Closed  bool   // true when a matching {{< /name >}} closing line was seen
}

// Kind implements ast.Node.Kind.
func (n *ShortcodeNode) Kind() ast.NodeKind { return kindShortcode }

// Dump implements ast.Node.Dump.
func (n *ShortcodeNode) Dump(source []byte, level int) {
	m := map[string]string{"Name": n.Name, "Params": n.Params}
	ast.DumpHelper(n, source, level, m, nil)
}

var shortcodeOpenRe = regexp.MustCompile(`^\{\{<\s*([a-zA-Z][a-zA-Z0-9-]*)\s*([^>]*?)\s*>\}\}`)
var shortcodeCloseRe = regexp.MustCompile(`^\{\{<\s*/\s*([a-zA-Z][a-zA-Z0-9-]*)\s*>\}\}\s*$`)
var shortcodeCloseInlineRe = regexp.MustCompile(`\{\{<\s*/\s*([a-zA-Z][a-zA-Z0-9-]*)\s*>\}\}`)

// shortcodeBlockParser collects lines between {{< name >}} and {{< /name >}}.
// Adapted to goldmark v1.8.2's parser.State API: Open must not consume the
// line (the parser framework advances past the opening line itself), and
// Continue consumes one line per call via AdvanceToEOL (which stops before
// the newline; the framework's AdvanceLine skips it).
type shortcodeBlockParser struct{}

func (p *shortcodeBlockParser) Trigger() []byte {
	return []byte("{{<")
}

func (p *shortcodeBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, seg := reader.PeekLine()
	m := shortcodeOpenRe.FindSubmatchIndex(line)
	if m == nil {
		return nil, parser.NoChildren
	}
	node := &ShortcodeNode{
		Name:   string(line[m[2]:m[3]]),
		Params: string(line[m[4]:m[5]]),
	}
	// Inline form: {{< name params >}}rest… (possibly with a same-line
	// {{< /name >}}). The rest after the opening marker becomes the first
	// content line; a same-line closing marker closes the block immediately.
	rest := line[m[1]:]
	if len(rest) > 0 {
		if cm := shortcodeCloseInlineRe.FindSubmatchIndex(rest); cm != nil && string(rest[cm[2]:cm[3]]) == node.Name {
			node.Closed = true
			rest = rest[:cm[0]]
		}
		if len(rest) > 0 {
			// seg.Start is the absolute source offset of line[0].
			absStart := seg.Start + m[1]
			node.Lines().Append(text.NewSegment(absStart, absStart+len(rest)))
		}
	}
	// The opening line is consumed by the framework (AdvanceLine after a
	// successful Open); it is not part of Content beyond the inline rest.
	return node, parser.NoChildren
}

func (p *shortcodeBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	sc := node.(*ShortcodeNode)
	// Inline form: the closing marker was consumed on the opening line, so
	// the block is already closed — terminate here. Without this, every
	// line to EOF gets swallowed into Content (verify 🔴).
	if sc.Closed {
		return parser.Close
	}
	line, seg := reader.PeekLine()
	// v0.8 has no nested shortcodes: a new {{< name >}} open line before a
	// matching close means this block is unclosed — terminate it WITHOUT
	// consuming the line, so the new open starts its own block (verify 🟡).
	if shortcodeOpenRe.Match(line) {
		return parser.Close
	}
	m := shortcodeCloseRe.FindSubmatch(line)
	if m != nil && string(m[1]) == sc.Name {
		sc.Closed = true
		reader.AdvanceToEOL()
		return parser.Close
	}
	sc.Lines().Append(seg)
	reader.AdvanceToEOL()
	return parser.Continue | parser.NoChildren
}

func (p *shortcodeBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// Collect raw content lines into node.Content.
	sc := node.(*ShortcodeNode)
	var buf bytes.Buffer
	lines := sc.Lines()
	for i := 0; i < lines.Len(); i++ {
		buf.Write(reader.Value(lines.At(i)))
		buf.WriteByte('\n')
	}
	// Trim the leading/trailing blank lines goldmark's line advancement
	// leaves around multi-line content (verify ⚪6); known components already
	// tolerate them, pass-through output becomes cleaner.
	sc.Content = strings.TrimSpace(buf.String())
	sc.Lines().Clear()
}

func (p *shortcodeBlockParser) CanInterruptParagraph() bool { return true }
func (p *shortcodeBlockParser) CanAcceptIndentedLine() bool { return false }

// shortcodeRenderer renders ShortcodeNodes through the registry.
type shortcodeRenderer struct {
	reg *ShortcodeRegistry
}

func (r *shortcodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindShortcode, r.renderShortcode)
}

func (r *shortcodeRenderer) renderShortcode(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	sc := node.(*ShortcodeNode)
	if !sc.Closed {
		// Unclosed shortcode (EOF before {{< /name >}}): pass through
		// verbatim (NOT escaped — the literal {{< must remain visible so
		// the author can spot the missing closing marker; TestRenderShortcode_Unclosed).
		_, _ = w.WriteString("<!-- unclosed shortcode: " + sc.Name + " -->")
		_, _ = w.WriteString("{{< " + sc.Name + " " + sc.Params + " >}}\n" + sc.Content)
		return ast.WalkSkipChildren, nil
	}
	if r.reg == nil {
		_, _ = w.WriteString(escapeHTML("{{< " + sc.Name + " " + sc.Params + " >}}"))
		return ast.WalkSkipChildren, nil
	}
	html, ok := r.reg.Render(sc.Name, sc.Params, sc.Content, r.reg.innerMD)
	if !ok {
		// Unknown: pass through verbatim (escaped) so the author can see it.
		_, _ = w.WriteString("<!-- unknown shortcode: " + sc.Name + " -->")
		_, _ = w.WriteString(escapeHTML("{{< " + sc.Name + " " + sc.Params + " >}}\n" + sc.Content + "{{< /" + sc.Name + " >}}"))
		return ast.WalkSkipChildren, nil
	}
	// Y1: write a placeholder token instead of raw HTML so downstream
	// post-processing (injectLangLabels / injectHLLines / splitLineNumbers)
	// only sees top-level code blocks. Replaced in RenderHTML's tail via
	// reg.ReplacePlaceholders.
	r.reg.WritePlaceholder(w, html)
	return ast.WalkSkipChildren, nil
}

// shortcodeExt wires the parser and renderer into a goldmark instance.
type shortcodeExt struct {
	reg *ShortcodeRegistry
}

func (e *shortcodeExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(&shortcodeBlockParser{}, 250),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&shortcodeRenderer{reg: e.reg}, 250),
	))
}

// stripShortcodeBlocks blanks shortcode block lines ({{< name >}} … {{< /name >}})
// while preserving line count, so top-level lang/hl extraction never sees
// shortcode-internal code. Fenced code regions are skipped entirely.
func stripShortcodeBlocks(src []byte) []byte {
	lines := bytes.Split(src, []byte("\n"))
	inCode, inSC := false, false
	for i, line := range lines {
		trimmed := strings.TrimSpace(string(line))
		switch {
		case inSC:
			// shortcode-internal lines are blanked; a matching close ends the block.
			// NOTE: checked FIRST — a ``` line inside a shortcode must not be
			// mistaken for a top-level fence (probe-verified).
			if shortcodeCloseRe.MatchString(trimmed) {
				inSC = false
			}
			lines[i] = nil
		case !inCode && strings.HasPrefix(trimmed, "```"):
			inCode = true
		case inCode && strings.HasPrefix(trimmed, "```"):
			inCode = false
		case inCode:
			// inside fenced code: untouched (shortcodes inside fences are literal)
		case shortcodeOpenRe.MatchString(trimmed):
			// Single-line inline form ({{< name >}}rest{{< /name >}} on one
			// line) closes itself: blank the line but do NOT enter inSC,
			// otherwise every following line gets blanked and top-level
			// lang/hl extraction is lost (verify 🔴, same root cause).
			if !shortcodeCloseInlineRe.MatchString(trimmed) {
				inSC = true
			}
			lines[i] = nil
		}
	}
	return bytes.Join(lines, []byte("\n"))
}
