package build

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/KurongTohsaka/chenhai-hugo/themes/zhenhai"
)

// copyStatic copies all files from static/ to public/ recursively.
func (b *Builder) copyStatic(public string) error {
	staticDir := filepath.Join(b.root, "static")
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		return nil
	}
	return filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(staticDir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		outPath := filepath.Join(public, relPath)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(outPath, data, 0644)
	})
}

// copyThemeAssets copies embedded theme CSS and JS to public/assets/,
// then overlays external theme assets if configured.
func (b *Builder) copyThemeAssets(public string) error {
	// Copy embedded zhenhai assets first
	if err := fs.WalkDir(zhenhai.FS, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := zhenhai.FS.ReadFile(path)
		if err != nil {
			return err
		}
		outPath := filepath.Join(public, path)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(outPath, data, 0644)
	}); err != nil {
		return err
	}

	// Overlay with external theme assets if configured
	if b.cfg.Theme != "zhenhai" {
		extAssets := filepath.Join(b.root, "themes", b.cfg.Theme, "assets")
		if info, err := os.Stat(extAssets); err == nil && info.IsDir() {
			if err := filepath.Walk(extAssets, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return err
				}
				relPath, err := filepath.Rel(extAssets, path)
				if err != nil {
					return err
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				outPath := filepath.Join(public, "assets", relPath)
				if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
					return err
				}
				return os.WriteFile(outPath, data, 0644)
			}); err != nil {
				return fmt.Errorf("copy external theme assets: %w", err)
			}
		}
	}
	return nil
}

// copyThemeStatic copies embedded theme static files (e.g. favicon) to public/,
// then overlays external theme static if configured.
func (b *Builder) copyThemeStatic(public string) error {
	// Copy embedded zhenhai static first
	if err := fs.WalkDir(zhenhai.FS, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := zhenhai.FS.ReadFile(path)
		if err != nil {
			return err
		}
		relPath := path[len("static/"):]
		outPath := filepath.Join(public, relPath)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(outPath, data, 0644)
	}); err != nil {
		return err
	}

	// Overlay with external theme static if configured
	if b.cfg.Theme != "zhenhai" {
		extStatic := filepath.Join(b.root, "themes", b.cfg.Theme, "static")
		if info, err := os.Stat(extStatic); err == nil && info.IsDir() {
			if err := filepath.Walk(extStatic, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return err
				}
				relPath, err := filepath.Rel(extStatic, path)
				if err != nil {
					return err
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				outPath := filepath.Join(public, relPath)
				if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
					return err
				}
				return os.WriteFile(outPath, data, 0644)
			}); err != nil {
				return fmt.Errorf("copy external theme static: %w", err)
			}
		}
	}
	return nil
}
