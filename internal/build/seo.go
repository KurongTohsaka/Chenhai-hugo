package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
)

// writeSitemap generates a minimal XML sitemap.
func (b *Builder) writeSitemap(site *index.Site, public string) error {
	var buf strings.Builder
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	baseURL := strings.TrimRight(b.cfg.BaseURL, "/")

	// Homepage entry
	buf.WriteString(fmt.Sprintf("  <url><loc>%s/</loc></url>\n", baseURL))

	// Page entries
	for _, page := range site.PublishedPages() {
		buf.WriteString(fmt.Sprintf("  <url><loc>%s%s</loc></url>\n", baseURL, page.Permalink()))
	}

	buf.WriteString("</urlset>\n")
	return os.WriteFile(filepath.Join(public, "sitemap.xml"), []byte(buf.String()), 0644)
}

// writeRobotsTXT generates a simple robots.txt allowing all crawlers.
func (b *Builder) writeRobotsTXT(public string) error {
	var buf strings.Builder
	buf.WriteString("User-agent: *\n")
	buf.WriteString("Allow: /\n\n")
	buf.WriteString(fmt.Sprintf("Sitemap: %s/sitemap.xml\n", strings.TrimRight(b.cfg.BaseURL, "/")))
	return os.WriteFile(filepath.Join(public, "robots.txt"), []byte(buf.String()), 0644)
}
