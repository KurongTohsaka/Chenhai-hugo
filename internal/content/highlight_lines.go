package content

import (
	"fmt"
	"regexp"
	"strings"
)

// hlLinesRe matches {hl_lines=[...]} in fenced code info strings.
var hlLinesRe = regexp.MustCompile(`\{hl_lines=\[([0-9,\-\s]+)\]\}`)

// extractHLLines extracts and returns hl_lines annotations from markdown source,
// and returns the cleaned source without these annotations.
func extractHLLines(source []byte) (cleaned []byte, hlInfo []string) {
	hlInfo = []string{}
	cleaned = hlLinesRe.ReplaceAll(source, []byte(""))
	matches := hlLinesRe.FindAllSubmatch(source, -1)
	for _, m := range matches {
		hlInfo = append(hlInfo, string(m[1]))
	}
	return
}

// injectHLLines injects data-hl-lines attributes into <pre class="chroma"> elements.
func injectHLLines(html string, hlInfo []string) string {
	if len(hlInfo) == 0 {
		return html
	}
	preRe := regexp.MustCompile(`<pre[^>]*class="chroma"[^>]*>`)
	idx := 0
	return preRe.ReplaceAllStringFunc(html, func(match string) string {
		if idx < len(hlInfo) && hlInfo[idx] != "" {
			// Insert data-hl-lines before the closing >
			result := match[:len(match)-1] + fmt.Sprintf(` data-hl-lines="%s">`, hlInfo[idx])
			idx++
			return result
		}
		idx++
		return match
	})
}

// chromaBlockRe matches <pre ... class="chroma" ...><code>content</code></pre>.
var chromaBlockRe = regexp.MustCompile(`<pre[^>]*class="chroma"[^>]*><code>([\s\S]*?)</code></pre>`)

// lnSpanRe matches a <span class="ln">text</span> element.
var lnSpanRe = regexp.MustCompile(`<span class="ln">[^<]*</span>`)

// splitLineNumbers restructures <pre class="chroma"> blocks into a two-column
// flex layout: line numbers in a fixed left column, code in a scrollable right
// column. Preserves any extra attributes (data-hl-lines, etc.) on the pre tag.
func splitLineNumbers(html string) string {
	return chromaBlockRe.ReplaceAllStringFunc(html, func(block string) string {
		subs := chromaBlockRe.FindStringSubmatch(block)
		if len(subs) < 2 {
			return block
		}
		inner := subs[1]

		// Extract line number text
		var lnLines []string
		lnMatches := lnSpanRe.FindAllString(inner, -1)
		for _, ln := range lnMatches {
			text := ln[len(`<span class="ln">`) : len(ln)-len(`</span>`)]
			lnLines = append(lnLines, text)
		}

		if len(lnLines) == 0 {
			return block // no line numbers, keep original
		}

		// Remove line number spans from code
		cleaned := lnSpanRe.ReplaceAllString(inner, "")

		lnHTML := "<pre><code>" + strings.Join(lnLines, "\n") + "</code></pre>"

		// Extract extra attributes (data-hl-lines etc.) from original pre tag
		attrs := ""
		openTagEnd := strings.Index(block, ">")
		if openTagEnd > 0 {
			openTag := block[:openTagEnd]
			// Remove <pre and class="chroma" to get extra attrs
			rest := strings.Replace(openTag, `<pre`, "", 1)
			rest = strings.Replace(rest, `class="chroma"`, "", 1)
			rest = strings.TrimSpace(rest)
			if rest != "" {
				attrs = " " + rest
			}
		}

		return fmt.Sprintf(
			`<div class="code-wrapper"><div class="code-ln">%s</div><div class="code-body"><pre class="chroma"%s><code>%s</code></pre></div></div>`,
			lnHTML, attrs, cleaned,
		)
	})
}
