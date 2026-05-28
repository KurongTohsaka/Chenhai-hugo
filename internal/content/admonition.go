package content

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// admonRe matches the [!TYPE] syntax at the beginning of a line.
// The paragraph lines inside a blockquote already have the "> " prefix stripped
// by goldmark's blockquote parser, so we only match the bracket syntax.
// We use [ \t] instead of \s to avoid matching newlines.
var admonRe = regexp.MustCompile(`^\[!(note|warning|tip|danger)\][ \t]*`)

var admonTitles = map[string]string{
	"note":    "笔记",
	"warning": "注意",
	"tip":     "提示",
	"danger":  "危险",
}

// KindAdmonition is a NodeKind of the Admonition node.
var KindAdmonition = ast.NewNodeKind("Admonition")

// AdmonitionNode represents an admonition block (note, warning, tip, danger).
type AdmonitionNode struct {
	ast.BaseBlock
	AdType    string
	Title     string
	InnerHTML string
}

// Kind implements ast.Node.Kind.
func (n *AdmonitionNode) Kind() ast.NodeKind {
	return KindAdmonition
}

// Dump implements ast.Node.Dump.
func (n *AdmonitionNode) Dump(source []byte, level int) {
	m := map[string]string{"Type": n.AdType, "Title": n.Title}
	ast.DumpHelper(n, source, level, m, nil)
}

type admonitionExt struct{}

// Admonition is a Goldmark extension that converts blockquote-style admonition
// syntax like > [!note] into styled HTML divs.
var Admonition = &admonitionExt{}

func (e *admonitionExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&admonitionTransformer{}, 200),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&admonRenderer{}, 200),
	))
}

// ---------------------------------------------------------------------------
// AST Transformer: converts blockquotes that match [!TYPE] into AdmonitionNodes
// ---------------------------------------------------------------------------

type admonitionTransformer struct{}

func (t *admonitionTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()

	// Collect matching blockquotes before modifying anything.
	type replacement struct {
		bq     *ast.Blockquote
		adType string
		title  string
	}
	var replacements []replacement

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		bq, ok := n.(*ast.Blockquote)
		if !ok {
			return ast.WalkContinue, nil
		}
		adType, title, found := checkAdmonition(bq, source)
		if !found {
			return ast.WalkContinue, nil
		}
		replacements = append(replacements, replacement{bq, adType, title})
		return ast.WalkContinue, nil
	})

	for _, r := range replacements {
		replaceBlockquote(r.bq, r.adType, r.title, source)
	}
}

// checkAdmonition determines whether a blockquote is an admonition by checking
// its first paragraph's first line for the [!TYPE] syntax.
func checkAdmonition(bq *ast.Blockquote, source []byte) (string, string, bool) {
	firstChild := bq.FirstChild()
	if firstChild == nil {
		return "", "", false
	}
	p, ok := firstChild.(*ast.Paragraph)
	if !ok {
		return "", "", false
	}
	lines := p.Lines()
	if lines.Len() == 0 {
		return "", "", false
	}
	firstSeg := lines.At(0)
	firstText := string(firstSeg.Value(source))
	m := admonRe.FindStringSubmatch(firstText)
	if m == nil {
		return "", "", false
	}
	adType := m[1]
	return adType, admonTitles[adType], true
}

// replaceBlockquote replaces a matching blockquote with an AdmonitionNode that
// carries pre-rendered inner HTML.
func replaceBlockquote(bq *ast.Blockquote, adType, title string, source []byte) {
	parent := bq.Parent()
	if parent == nil {
		return
	}

	// Collect raw content from all child paragraphs, stripping [!TYPE] from
	// the first line of the first paragraph.
	var contentBuf bytes.Buffer
	firstLine := true
	for child := bq.FirstChild(); child != nil; child = child.NextSibling() {
		cp, ok := child.(*ast.Paragraph)
		if !ok {
			continue
		}
		for i := 0; i < cp.Lines().Len(); i++ {
			seg := cp.Lines().At(i)
			line := string(seg.Value(source))
			if firstLine && i == 0 {
				line = admonRe.ReplaceAllString(line, "")
				firstLine = false
			}
			contentBuf.WriteString(line)
			if !strings.HasSuffix(line, "\n") {
				contentBuf.WriteByte('\n')
			}
		}
	}

	// Render the inner content as Markdown using a minimal goldmark instance
	// (no extensions, so no admonition — no infinite recursion risk).
	innerMD := goldmark.New()
	var innerHTML bytes.Buffer
	if err := innerMD.Convert(contentBuf.Bytes(), &innerHTML); err != nil {
		innerHTML.Write(contentBuf.Bytes())
	}

	// Create the admonition node and swap it in.
	adNode := &AdmonitionNode{
		BaseBlock: ast.BaseBlock{},
		AdType:    adType,
		Title:     title,
		InnerHTML: innerHTML.String(),
	}
	parent.ReplaceChild(parent, bq, adNode)
}

// ---------------------------------------------------------------------------
// Node renderer: renders AdmonitionNodes (not Blockquote nodes)
// ---------------------------------------------------------------------------

type admonRenderer struct{}

func (r *admonRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindAdmonition, r.renderAdmonition)
}

func (r *admonRenderer) renderAdmonition(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	adNode := n.(*AdmonitionNode)
	_, _ = fmt.Fprintf(w,
		`<div class="admonition admonition-%s"><div class="admonition-title">%s</div><div class="admonition-content">%s</div></div>`,
		adNode.AdType, adNode.Title, adNode.InnerHTML,
	)
	return ast.WalkSkipChildren, nil
}
