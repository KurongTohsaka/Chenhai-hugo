package theme

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/themes/zhenhai"
)

// TemplateData holds the data passed to every rendered template.
type TemplateData struct {
	Site   interface{}            // will be *index.Site after Task 6
	Page   *content.Page
	Config *config.Config
	Extra  map[string]interface{}
}

// Engine manages Go html/template loading and rendering.
type Engine struct {
	templates *template.Template
	siteRoot  string
}

// New creates a new template Engine. Loads embedded Zhenhai theme first,
// then overlays any site-specific layouts/ directory.
func New(cfg *config.Config, siteRoot string) (*Engine, error) {
	funcMap := template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"formatDate": func(d interface{}, f string) string {
			return fmt.Sprintf("%v", d)
		},
	}

	t := template.New("").Funcs(funcMap)

	// Load embedded Zhenhai theme layouts
	if err := fs.WalkDir(zhenhai.FS, "layouts", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := zhenhai.FS.ReadFile(path)
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(path, "layouts/")
		if _, err := t.New(name).Parse(string(b)); err != nil {
			return fmt.Errorf("parse embedded template %s: %w", name, err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("load embedded theme: %w", err)
	}

	// Override with site's layouts/ directory if present
	layoutsDir := filepath.Join(siteRoot, "layouts")
	if info, err := os.Stat(layoutsDir); err == nil && info.IsDir() {
		if err := filepath.Walk(layoutsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			name, _ := filepath.Rel(layoutsDir, path)
			name = filepath.ToSlash(name)
			if _, err := t.New(name).Parse(string(b)); err != nil {
				return fmt.Errorf("parse site template %s: %w", name, err)
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("load site layouts: %w", err)
		}
	}

	return &Engine{templates: t, siteRoot: siteRoot}, nil
}

// Render executes the named template with the given data.
func (e *Engine) Render(w interface{ Write([]byte) (int, error) }, name string, data *TemplateData) error {
	return e.templates.ExecuteTemplate(w, name, data)
}

// HasTemplate returns true if the named template exists.
func (e *Engine) HasTemplate(name string) bool {
	return e.templates.Lookup(name) != nil
}

// RenderPage renders the base layout (base.html) wrapping the content block
// from the named template (e.g., "single.html", "list.html", "index.html").
// It clones the template set to avoid "Parse after Execute" panics,
// re-parses the named template to ensure the correct "content"
// sub-template definition is used, then executes base.html.
func (e *Engine) RenderPage(w interface{ Write([]byte) (int, error) }, name string, data *TemplateData) error {
	// Clone to avoid "cannot Parse after Execute"
	clone, err := e.templates.Clone()
	if err != nil {
		return fmt.Errorf("clone templates: %w", err)
	}

	// Read the template source
	src, err := e.readLayoutSource(name)
	if err != nil {
		return fmt.Errorf("read layout %s: %w", name, err)
	}

	// Re-parse to update the "content" sub-template definition
	if _, err := clone.Parse(string(src)); err != nil {
		return fmt.Errorf("parse layout %s: %w", name, err)
	}

	// Execute base.html which uses {{template "content" .}}
	return clone.ExecuteTemplate(w, "base.html", data)
}

// readLayoutSource reads the source of a layout template, checking site-specific
// layouts first, then falling back to the embedded Zhenhai theme.
func (e *Engine) readLayoutSource(name string) ([]byte, error) {
	// Check site-specific layouts directory
	if e.siteRoot != "" {
		siteLayoutPath := filepath.Join(e.siteRoot, "layouts", name)
		if info, err := os.Stat(siteLayoutPath); err == nil && !info.IsDir() {
			return os.ReadFile(siteLayoutPath)
		}
	}

	// Fall back to embedded theme
	return zhenhai.FS.ReadFile("layouts/" + name)
}
