package theme

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
	"github.com/KurongTohsaka/chenhai-hugo/themes/zhenhai"
	"gopkg.in/yaml.v3"
)
// ThemeMeta holds theme metadata read from theme.yaml.
type ThemeMeta struct {
	Name        string                 `yaml:"name"`
	Version     string                 `yaml:"version"`
	Description string                 `yaml:"description"`
	Author      string                 `yaml:"author"`
	Params      map[string]interface{} `yaml:"params"`
}


// TemplateData holds the data passed to every rendered template.
type TemplateData struct {
	Site   *index.Site
	Page   *content.Page
	Config *config.Config
	Extra  map[string]interface{}
}

// Engine manages Go html/template loading and rendering.
type Engine struct {
	templates          *template.Template
	siteRoot           string
	extThemeLayoutsDir string // path to external theme layouts/, empty if n/a
	templateSrcs       map[string][]byte // template source cache, keyed by name
	srcMu              sync.RWMutex
}

// SetTemplateSrcs sets the template source cache for reuse across renders.
func (e *Engine) SetTemplateSrcs(srcs map[string][]byte) {
	e.templateSrcs = srcs
}

// TemplateSrcs returns the template source cache.
func (e *Engine) TemplateSrcs() map[string][]byte {
	return e.templateSrcs
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

	// Override with external theme layouts/ directory if applicable
	extThemeLayoutsDir := ""
	if cfg.Theme != "zhenhai" {
		extDir := filepath.Join(siteRoot, "themes", cfg.Theme, "layouts")
		if info, err := os.Stat(extDir); err == nil && info.IsDir() {
			if err := filepath.Walk(extDir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return err
				}
				b, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				name, _ := filepath.Rel(extDir, path)
				name = filepath.ToSlash(name)
				if _, err := t.New(name).Parse(string(b)); err != nil {
					return fmt.Errorf("parse theme template %s: %w", name, err)
				}
				return nil
			}); err != nil {
				return nil, fmt.Errorf("load external theme: %w", err)
			}
		}
		extThemeLayoutsDir = extDir
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

	// Merge theme.yaml params into config (custom params from user's config.yaml
	// take precedence over theme defaults)
	if err := mergeThemeParams(cfg, siteRoot); err != nil {
		return nil, fmt.Errorf("merge theme params: %w", err)
	}

	return &Engine{templates: t, siteRoot: siteRoot, extThemeLayoutsDir: extThemeLayoutsDir}, nil
}

// Render executes the named template with the given data.
func (e *Engine) Render(w io.Writer, name string, data *TemplateData) error {
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
func (e *Engine) RenderPage(w io.Writer, name string, data *TemplateData) error {
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

// readLayoutSource reads the source of a layout template, checking the in-memory
// cache first, then site-specific layouts, then the external theme layouts, then
// falling back to the embedded Zhenhai theme.
func (e *Engine) readLayoutSource(name string) ([]byte, error) {
	// Check in-memory cache first (read lock)
	e.srcMu.RLock()
	if e.templateSrcs != nil {
		if src, ok := e.templateSrcs[name]; ok {
			e.srcMu.RUnlock()
			return src, nil
		}
	}
	e.srcMu.RUnlock()

	var src []byte
	var err error

	// Check site-specific layouts directory
	if e.siteRoot != "" {
		siteLayoutPath := filepath.Join(e.siteRoot, "layouts", name)
		if info, statErr := os.Stat(siteLayoutPath); statErr == nil && !info.IsDir() {
			src, err = os.ReadFile(siteLayoutPath)
			if err == nil {
				e.srcMu.Lock()
				if e.templateSrcs == nil {
					e.templateSrcs = make(map[string][]byte)
				}
				e.templateSrcs[name] = src
				e.srcMu.Unlock()
				return src, nil
			}
		}
	}

	// Check external theme layouts directory
	if e.extThemeLayoutsDir != "" {
		extPath := filepath.Join(e.extThemeLayoutsDir, name)
		if info, statErr := os.Stat(extPath); statErr == nil && !info.IsDir() {
			src, err = os.ReadFile(extPath)
			if err == nil {
				e.srcMu.Lock()
				if e.templateSrcs == nil {
					e.templateSrcs = make(map[string][]byte)
				}
				e.templateSrcs[name] = src
				e.srcMu.Unlock()
				return src, nil
			}
		}
	}

	// Fall back to embedded theme
	src, err = zhenhai.FS.ReadFile("layouts/" + name)
	if err == nil {
		e.srcMu.Lock()
		if e.templateSrcs == nil {
			e.templateSrcs = make(map[string][]byte)
		}
		e.templateSrcs[name] = src
		e.srcMu.Unlock()
	}
	return src, err
}

// LookupShortcodeTemplate returns a theme-provided shortcode template
// (layouts/shortcodes/<name>.html), reusing readLayoutSource's three-layer
// lookup (site → external theme → embedded Zhenhai) plus its templateSrcs
// cache and srcMu lock. Zero new traversal code.
func (e *Engine) LookupShortcodeTemplate(name string) ([]byte, bool) {
	src, err := e.readLayoutSource("shortcodes/" + name + ".html")
	if err != nil {
		return nil, false
	}
	return src, true
}

// mergeThemeParams loads theme.yaml from the active theme and merges
// any default params into cfg.ThemeConfig.Params.
// User-defined values in config.yaml take precedence and are not overwritten.
func mergeThemeParams(cfg *config.Config, siteRoot string) error {
	// Try external theme first
	themeYAMLPath := filepath.Join(siteRoot, "themes", cfg.Theme, "theme.yaml")
	data, err := os.ReadFile(themeYAMLPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// Fall back to embedded zhenhai theme
		if cfg.Theme == "zhenhai" {
			data, err = zhenhai.FS.ReadFile("theme.yaml")
			if err != nil {
				return nil // shouldn't happen, but be safe
			}
		} else {
			return nil // no theme.yaml, nothing to merge
		}
	}

	var meta ThemeMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("parse theme.yaml: %w", err)
	}

	if len(meta.Params) == 0 {
		return nil // nothing to merge
	}

	// Initialize Params map if needed
	if cfg.ThemeConfig.Params == nil {
		cfg.ThemeConfig.Params = make(map[string]interface{})
	}

	// Merge defaults: only add keys not already set by user's config.yaml
	for k, v := range meta.Params {
		if _, exists := cfg.ThemeConfig.Params[k]; !exists {
			cfg.ThemeConfig.Params[k] = v
		}
	}

	return nil
}
