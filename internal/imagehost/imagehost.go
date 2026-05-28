package imagehost

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
)

// imagePattern matches Markdown image syntax: ![alt](url) or ![alt](url "title").
// Group 1: alt text
// Group 2: URL/path
// Group 3: optional title with leading space, e.g. ` "My Title"`
var imagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^\s)]+)(\s+"[^"]*")?\)`)

// Host handles image upload and URL mapping for GitHub image hosting.
type Host struct {
	cfg         *config.ImageHostConfig
	uploadCache map[string]string // filename -> sha256 hex (for auto mode)
}

// New creates a new Host.
func New(cfg *config.ImageHostConfig) *Host {
	return &Host{
		cfg:         cfg,
		uploadCache: make(map[string]string),
	}
}

// Process takes raw markdown content and returns processed content
// with image URLs replaced. In "map" mode it just rewrites paths.
// In "auto" mode it uploads to GitHub and rewrites paths.
func (h *Host) Process(markdown []byte, contentDir string) ([]byte, error) {
	if !h.cfg.Enabled {
		return markdown, nil
	}

	token := h.cfg.Token
	if token == "" {
		token = os.Getenv("CHENHAI_IMG_TOKEN")
	}

	switch h.cfg.Mode {
	case "auto":
		return h.processAuto(markdown, contentDir, token)
	case "map":
		return h.processMap(markdown)
	default:
		return markdown, nil
	}
}

// isLocalPath returns true if the URL is a local file path (not a remote URL or data URI).
func isLocalPath(url string) bool {
	return !strings.HasPrefix(url, "http://") &&
		!strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "data:")
}

// processAuto scans markdown for local image references, uploads each
// image to GitHub via API, and replaces local paths with CDN URLs.
func (h *Host) processAuto(markdown []byte, contentDir, token string) ([]byte, error) {
	result := imagePattern.ReplaceAllStringFunc(string(markdown), func(match string) string {
		parts := imagePattern.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		alt := parts[1]
		url := parts[2]
		title := ""
		if len(parts) >= 4 {
			title = parts[3]
		}

		// Skip remote URLs and data URIs
		if !isLocalPath(url) {
			return match
		}

		// Resolve file path relative to content directory
		imgPath := url
		if !filepath.IsAbs(imgPath) {
			imgPath = filepath.Join(contentDir, imgPath)
		}

		filename := filepath.Base(url)

		// Only attempt upload if we have both token and repo
		if token != "" && h.cfg.Repo != "" {
			if err := h.uploadToGitHub(imgPath, filename, token); err != nil {
				log.Printf("image host: failed to upload %s: %v — keeping local path", filename, err)
				return match
			}
		} else {
			// No token or repo configured, keep original path
			log.Printf("image host: skipping upload for %s (no token or repo configured)", filename)
			return match
		}

		// Build CDN URL: https://raw.githubusercontent.com/{repo}/{branch}/{basePath}{filename}
		cdnBase := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s",
			strings.Trim(h.cfg.Repo, "/"),
			strings.TrimLeft(h.cfg.Branch, "/"),
			strings.Trim(h.cfg.BasePath, "/"))
		cdnURL := cdnBase + "/" + filename

		if title != "" {
			return "![" + alt + "](" + cdnURL + title + ")"
		}
		return "![" + alt + "](" + cdnURL + ")"
	})

	return []byte(result), nil
}

// uploadToGitHub uploads a file to the GitHub repository using the GitHub
// Contents API. If the upload fails (e.g., missing token, API error), it
// returns an error and the caller keeps the original local path.
func (h *Host) uploadToGitHub(filePath, filename, token string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Compute hash for caching
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	if cachedHash, ok := h.uploadCache[filename]; ok && cachedHash == hash {
		// Same filename + hash already uploaded this build — skip
		return nil
	}

	// Prepare GitHub API request body
	ghPath := strings.Trim(h.cfg.BasePath, "/") + "/" + filename
	content := base64.StdEncoding.EncodeToString(data)

	apiBody := map[string]interface{}{
		"message": "Upload via Chenhai " + filename,
		"content": content,
		"branch":  h.cfg.Branch,
	}
	bodyBytes, err := json.Marshal(apiBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s",
		strings.Trim(h.cfg.Repo, "/"),
		strings.TrimLeft(ghPath, "/"))

	req, err := http.NewRequest("PUT", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	// 201 Created (new file) or 200 OK (updated file) indicate success
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	h.uploadCache[filename] = hash
	return nil
}

// processMap replaces local image paths with the configured base URL prefix.
// For each local image, the path is replaced with {baseURL}/{filename}.
func (h *Host) processMap(markdown []byte) ([]byte, error) {
	baseURL := strings.TrimRight(h.cfg.BaseURL, "/")

	result := imagePattern.ReplaceAllStringFunc(string(markdown), func(match string) string {
		parts := imagePattern.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		alt := parts[1]
		url := parts[2]
		title := ""
		if len(parts) >= 4 {
			title = parts[3]
		}

		// Skip remote URLs and data URIs
		if !isLocalPath(url) {
			return match
		}

		// Replace with baseURL + filename
		filename := filepath.Base(url)
		newURL := baseURL + "/" + filename

		if title != "" {
			return "![" + alt + "](" + newURL + title + ")"
		}
		return "![" + alt + "](" + newURL + ")"
	})

	return []byte(result), nil
}
