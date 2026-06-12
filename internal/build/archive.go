package build

import (
	"encoding/json"
	"path/filepath"

	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
	"github.com/KurongTohsaka/chenhai-hugo/internal/theme"
)

// renderArchives renders /archives/index.html with all published pages.
func (b *Builder) renderArchives(site *index.Site, public string) error {
	archiveDir := filepath.Join(public, "archives")
	published := site.PublishedPages()
	archiveData := &theme.TemplateData{
		Site:   site,
		Page:   &content.Page{Title: "Archives"},
		Config: b.cfg,
		Extra:  map[string]interface{}{"title": "Archives", "pages": published},
	}
	if heatmap := site.BuildHeatmap(); len(heatmap) > 0 {
		b, _ := json.Marshal(heatmap)
		archiveData.Extra["heatmapJSON"] = string(b)
	}
	return b.renderToFile(archiveData, filepath.Join(archiveDir, "index.html"), "list.html")
}
