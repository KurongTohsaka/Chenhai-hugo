package content

import (
	"fmt"
	"regexp"
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
	preRe := regexp.MustCompile(`<pre class="chroma">`)
	idx := 0
	return preRe.ReplaceAllStringFunc(html, func(match string) string {
		if idx < len(hlInfo) && hlInfo[idx] != "" {
			result := fmt.Sprintf(`<pre class="chroma" data-hl-lines="%s">`, hlInfo[idx])
			idx++
			return result
		}
		idx++
		return match
	})
}
