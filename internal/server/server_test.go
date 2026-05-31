package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
)

func TestNew(t *testing.T) {
	cfg := config.DefaultConfig()
	dir := t.TempDir()
	s := New(cfg, dir, false)
	if s == nil {
		t.Fatal("New returned nil")
	}
	if s.cfg != cfg {
		t.Error("config not stored")
	}
	if s.root != dir {
		t.Error("root not stored")
	}
}

func TestIsHTMLFile(t *testing.T) {
	dir := t.TempDir()

	htmlPath := filepath.Join(dir, "test.html")
	os.WriteFile(htmlPath, []byte("<html>"), 0644)
	if !isHTMLFile(htmlPath) {
		t.Error("expected .html file to be detected")
	}

	txtPath := filepath.Join(dir, "test.txt")
	os.WriteFile(txtPath, []byte("text"), 0644)
	if isHTMLFile(txtPath) {
		t.Error("expected .txt file to NOT be detected")
	}

	if isHTMLFile(filepath.Join(dir, "nonexistent.html")) {
		t.Error("expected nonexistent file to NOT be detected")
	}
}

func TestLiveReloadInjection(t *testing.T) {
	dir := t.TempDir()
	testHTML := "<!DOCTYPE html><html><head></head><body><p>hello</p></body></html>"
	htmlPath := filepath.Join(dir, "test.html")
	os.WriteFile(htmlPath, []byte(testHTML), 0644)

	cfg := config.DefaultConfig()
	s := New(cfg, dir, false)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test.html", nil)
	s.serveHTMLWithLiveReload(w, r, htmlPath)

	body := w.Body.String()
	if !strings.Contains(body, "livereload") {
		t.Error("expected live reload script in output")
	}
	if !strings.Contains(body, "<p>hello</p>") {
		t.Error("expected original content preserved")
	}
}

func TestFileServerHTMLInjection(t *testing.T) {
	dir := t.TempDir()
	htmlPath := filepath.Join(dir, "test.html")
	os.WriteFile(htmlPath, []byte("<html><body>hi</body></html>"), 0644)
	cssPath := filepath.Join(dir, "style.css")
	os.WriteFile(cssPath, []byte("body{}"), 0644)

	cfg := config.DefaultConfig()
	s := New(cfg, dir, false)
	handler := s.fileServerWithLiveReload(dir)

	// HTML file should get live reload injected
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test.html", nil)
	handler.ServeHTTP(w, r)
	if !strings.Contains(w.Body.String(), "livereload") {
		t.Error("HTML response missing live reload script")
	}

	// CSS file should NOT get live reload injected
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/style.css", nil)
	handler.ServeHTTP(w2, r2)
	if strings.Contains(w2.Body.String(), "livereload") {
		t.Error("CSS response should not have live reload script")
	}
}

func TestHandleWebSocket(t *testing.T) {
	cfg := config.DefaultConfig()
	s := New(cfg, t.TempDir(), false)

	// Non-WebSocket request should be handled gracefully (no panic)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/livereload", nil)
	s.handleWebSocket(w, r)
	// Upgrader fails silently — just verify no panic
	if w.Code != 0 && w.Code != http.StatusSwitchingProtocols {
		// Either no response written, or switching protocols — both OK
	}
}
