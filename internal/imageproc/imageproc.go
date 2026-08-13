// Package imageproc provides image decode/encode/resize helpers used by
// the chenhai image command family.
package imageproc

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
	"golang.org/x/image/draw"
)

// Supported formats: jpg/jpeg/png/webp are transcodable; gif is copied as-is.
func IsTranscodable(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	}
	return false
}

// Decode reads an image file and returns the decoded image plus its format
// name ("jpg", "png", "webp").
func Decode(path string) (image.Image, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		img, err := jpeg.Decode(f)
		return img, "jpg", err
	case ".png":
		img, err := png.Decode(f)
		return img, "png", err
	case ".webp":
		img, err := webp.Decode(f)
		return img, "webp", err
	default:
		return nil, "", fmt.Errorf("unsupported image format: %s", ext)
	}
}

// DecodeBytes decodes an image from memory, format is "jpg"/"png"/"webp".
func DecodeBytes(data []byte, format string) (image.Image, error) {
	r := bytes.NewReader(data)
	switch format {
	case "jpg":
		return jpeg.Decode(r)
	case "png":
		return png.Decode(r)
	case "webp":
		return webp.Decode(r)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// EncodeWebP encodes an image to WebP bytes with the given quality (0-100).
func EncodeWebP(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Quality: float32(quality)}); err != nil {
		return nil, fmt.Errorf("encode webp: %w", err)
	}
	return buf.Bytes(), nil
}

// Resize scales img to fit within maxW x maxH, preserving aspect ratio.
// A zero dimension means "unconstrained". Returns the original image when
// no resize is needed.
func Resize(img image.Image, maxW, maxH int) image.Image {
	b := img.Bounds()
	srcW, srcH := b.Dx(), b.Dy()
	if (maxW <= 0 || srcW <= maxW) && (maxH <= 0 || srcH <= maxH) {
		return img
	}

	ratio := 1.0
	switch {
	case maxW > 0 && maxH > 0:
		rw := float64(maxW) / float64(srcW)
		rh := float64(maxH) / float64(srcH)
		if rw < rh {
			ratio = rw
		} else {
			ratio = rh
		}
	case maxW > 0:
		ratio = float64(maxW) / float64(srcW)
	default:
		ratio = float64(maxH) / float64(srcH)
	}

	dstW, dstH := int(float64(srcW)*ratio), int(float64(srcH)*ratio)
	if dstW < 1 {
		dstW = 1
	}
	if dstH < 1 {
		dstH = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)
	return dst
}

// NextImageName returns the next auto-incrementing name (img1.webp, img2.webp, …)
// in dir for the given extension (no dot). Unrelated files are ignored.
func NextImageName(dir, ext string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	maxN := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != "."+ext {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(strings.TrimSuffix(name, "."+ext), "img%d", &n); err == nil && n > maxN {
			maxN = n
		}
	}
	return fmt.Sprintf("img%d.%s", maxN+1, ext), nil
}

// WriteFile writes data to path, creating parent directories as needed.
func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
