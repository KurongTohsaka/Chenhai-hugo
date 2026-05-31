package server

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"

	"github.com/KurongTohsaka/chenhai-hugo/internal/build"
	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/internal/theme"
)

// liveReloadScript is injected into HTML responses for automatic reload.
const liveReloadScript = `<script>
(function(){var ws=new WebSocket("ws://"+location.host+"/livereload");
ws.onmessage=function(e){if(e.data==="reload")location.reload();};
ws.onclose=function(){setTimeout(function(){location.reload()},2000);};})();
</script>`

// Server wraps an HTTP server with file watching capabilities.
type Server struct {
	cfg        *config.Config
	root       string
	showDrafts bool
	http       *http.Server
	wsConns    map[*websocket.Conn]bool
	wsMu       sync.Mutex
	upgrader   websocket.Upgrader
}

// New creates a new dev server.
func New(cfg *config.Config, root string, showDrafts bool) *Server {
	s := &Server{
		cfg:        cfg,
		root:       root,
		showDrafts: showDrafts,
		wsConns: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
	return s
}

// Start builds the site, starts watching files, and serves on the given port.
func (s *Server) Start(port int) error {
	// Initial build
	if err := s.rebuild(); err != nil {
		log.Printf("initial build warning: %v", err)
	}

	// Start file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}

	// Watch directories
	for _, dir := range []string{"content", "themes", "static", "layouts"} {
		path := filepath.Join(s.root, dir)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
				if err == nil && info.IsDir() {
					watcher.Add(p)
				}
				return nil
			})
		}
	}
	// Watch config file
	configPath := filepath.Join(s.root, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		watcher.Add(configPath)
	}

	// Handle file changes with debounce
	go func() {
		defer watcher.Close()
		var debounceTimer *time.Timer
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Ignore writes to public/ directory
				if filepath.HasPrefix(event.Name, filepath.Join(s.root, "public")) {
					continue
				}
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(300*time.Millisecond, func() {
					log.Printf("file changed: %s", event.Name)
					if err := s.rebuild(); err != nil {
						log.Printf("rebuild error: %v", err)
					}
					s.notifyClients()
				})
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("watcher error: %v", err)
			}
		}
	}()

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/livereload", s.handleWebSocket)
	mux.Handle("/", s.fileServerWithLiveReload(filepath.Join(s.root, "public")))

	addr := fmt.Sprintf(":%d", port)
	s.http = &http.Server{Addr: addr, Handler: mux}

	fmt.Printf("Chenhai 开发服务器启动 → http://localhost%s\n", addr)
	return s.http.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func (s *Server) rebuild() error {
	renderer := content.NewRenderer(s.cfg.Markup.Highlight.Style, s.cfg.Markup.Highlight.LineNumbers)
	engine, err := theme.New(s.cfg, s.root)
	if err != nil {
		return fmt.Errorf("init theme: %w", err)
	}
	builder := build.New(s.cfg, s.root, renderer, engine, s.showDrafts)
	return builder.Build()
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.wsMu.Lock()
	s.wsConns[conn] = true
	s.wsMu.Unlock()

	// Keep connection alive — read until error (client disconnect)
	go func() {
		defer func() {
			s.wsMu.Lock()
			delete(s.wsConns, conn)
			s.wsMu.Unlock()
			conn.Close()
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

func (s *Server) notifyClients() {
	s.wsMu.Lock()
	defer s.wsMu.Unlock()
	for conn := range s.wsConns {
		if err := conn.WriteMessage(websocket.TextMessage, []byte("reload")); err != nil {
			conn.Close()
			delete(s.wsConns, conn)
		}
	}
}

// fileServerWithLiveReload returns an http.Handler that serves static files
// from publicDir, injecting the LiveReload script into HTML responses.
func (s *Server) fileServerWithLiveReload(publicDir string) http.Handler {
	fs := http.FileServer(http.Dir(publicDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		filePath := filepath.Join(publicDir, filepath.Clean(path))

		// Check if the request maps to a .html file
		if isHTMLFile(filePath) {
			s.serveHTMLWithLiveReload(w, r, filePath)
			return
		}

		// For directory-like paths, check for index.html
		if path == "" || path[len(path)-1] == '/' || filepath.Ext(path) == "" {
			indexPath := filepath.Join(filePath, "index.html")
			if info, err := os.Stat(indexPath); err == nil && !info.IsDir() {
				s.serveHTMLWithLiveReload(w, r, indexPath)
				return
			}
		}

		fs.ServeHTTP(w, r)
	})
}

// isHTMLFile returns true if the given path points to an existing .html file.
func isHTMLFile(path string) bool {
	if filepath.Ext(path) != ".html" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// serveHTMLWithLiveReload reads an HTML file and writes it with the
// LiveReload script injected before the closing </body> tag.
func (s *Server) serveHTMLWithLiveReload(w http.ResponseWriter, r *http.Request, filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Inject live reload script before the last </body> tag
	bodyClose := []byte("</body>")
	if idx := bytes.LastIndex(content, bodyClose); idx >= 0 {
		w.Write(content[:idx])
		w.Write([]byte(liveReloadScript))
		w.Write(content[idx:])
		return
	}

	// No </body> tag found — write script at the end
	w.Write(content)
	w.Write([]byte(liveReloadScript))
}
