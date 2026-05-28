package imagehost

import (
	"bytes"
	"testing"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
)

func TestMapMode_ReplacesLocalPaths(t *testing.T) {
	h := New(&config.ImageHostConfig{
		Enabled: true,
		Mode:    "map",
		BaseURL: "https://example.com/images",
	})

	input := []byte(`![alt](local.png)
![](another.jpg)
![with title](photo.jpg "My Photo")
`)
	result, err := h.Process(input, "/tmp")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(result, []byte("https://example.com/images/local.png")) {
		t.Errorf("expected CDN URL in result, got:\n%s", string(result))
	}
	if !bytes.Contains(result, []byte("https://example.com/images/another.jpg")) {
		t.Errorf("expected CDN URL for second image, got:\n%s", string(result))
	}
	if !bytes.Contains(result, []byte(`![with title](https://example.com/images/photo.jpg "My Photo")`)) {
		t.Errorf("expected CDN URL with title preserved, got:\n%s", string(result))
	}
}

func TestMapMode_SubdirectoryPaths(t *testing.T) {
	h := New(&config.ImageHostConfig{
		Enabled: true,
		Mode:    "map",
		BaseURL: "https://example.com/images",
	})

	input := []byte(`![alt](subdir/image.png)
![alt](../other/photo.jpg)
![alt](./relative/img.png)
`)
	result, err := h.Process(input, "/tmp")
	if err != nil {
		t.Fatal(err)
	}

	// Only the base filename should be kept, directory path stripped
	if !bytes.Contains(result, []byte("https://example.com/images/image.png")) {
		t.Errorf("expected image.png replaced, got:\n%s", string(result))
	}
	if !bytes.Contains(result, []byte("https://example.com/images/photo.jpg")) {
		t.Errorf("expected photo.jpg replaced, got:\n%s", string(result))
	}
	if !bytes.Contains(result, []byte("https://example.com/images/img.png")) {
		t.Errorf("expected img.png replaced, got:\n%s", string(result))
	}
}

func TestDisabled_NoChanges(t *testing.T) {
	h := New(&config.ImageHostConfig{
		Enabled: false,
	})

	input := []byte(`![alt](local.png)`)
	result, err := h.Process(input, "/tmp")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(result, input) {
		t.Errorf("expected unchanged output, got:\n%s", string(result))
	}
}

func TestURLDetection_NotReplaced(t *testing.T) {
	h := New(&config.ImageHostConfig{
		Enabled: true,
		Mode:    "map",
		BaseURL: "https://example.com/images",
	})

	input := []byte(`![alt](http://example.com/img.png)
![alt](https://example.com/img.png)
`)
	result, err := h.Process(input, "/tmp")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(result, input) {
		t.Errorf("expected unchanged for remote URLs, got:\n%s", string(result))
	}
}

func TestDataURIs_NotReplaced(t *testing.T) {
	h := New(&config.ImageHostConfig{
		Enabled: true,
		Mode:    "map",
		BaseURL: "https://example.com/images",
	})

	input := []byte(`![alt](data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==)`)
	result, err := h.Process(input, "/tmp")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(result, input) {
		t.Errorf("expected unchanged for data URIs, got:\n%s", string(result))
	}
}

func TestNoImages_PassesThrough(t *testing.T) {
	h := New(&config.ImageHostConfig{
		Enabled: true,
		Mode:    "map",
		BaseURL: "https://example.com/images",
	})

	input := []byte(`# Hello
This is text without images.

Just a [link](local.png) but not an image.
`)
	result, err := h.Process(input, "/tmp")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(result, input) {
		t.Errorf("expected unchanged, got:\n%s", string(result))
	}
}

func TestAutoMode_NoToken_KeepsPaths(t *testing.T) {
	h := New(&config.ImageHostConfig{
		Enabled:  true,
		Mode:     "auto",
		Provider: "github",
		Repo:     "",
		Branch:   "main",
		BasePath: "images/",
	})

	input := []byte(`![alt](local.png)`)
	result, err := h.Process(input, "/tmp")
	if err != nil {
		t.Fatal(err)
	}

	// Without a repo configured, auto mode should keep original paths
	if !bytes.Equal(result, input) {
		t.Errorf("expected unchanged (no repo), got:\n%s", string(result))
	}
}

func TestDefaultMode_NoChanges(t *testing.T) {
	h := New(&config.ImageHostConfig{
		Enabled: true,
		Mode:    "",
	})

	input := []byte(`![alt](local.png)`)
	result, err := h.Process(input, "/tmp")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(result, input) {
		t.Errorf("expected unchanged for unknown mode, got:\n%s", string(result))
	}
}
