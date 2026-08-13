package imageproc

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// testImage builds a 100x50 RGBA image with a red block for encoding tests.
func testImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	for y := 10; y < 20; y++ {
		for x := 10; x < 40; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	return img
}

// writePNG writes img to a temp .png file and returns its path.
func writePNG(t *testing.T, img image.Image) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "src.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestEncodeWebP_RoundTrip(t *testing.T) {
	data, err := EncodeWebP(testImage(), 80)
	if err != nil {
		t.Fatalf("EncodeWebP failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("encoded data is empty")
	}
	// Decode back and verify dimensions
	decoded, err := DecodeBytes(data, "webp")
	if err != nil {
		t.Fatalf("DecodeBytes failed: %v", err)
	}
	b := decoded.Bounds()
	if b.Dx() != 100 || b.Dy() != 50 {
		t.Errorf("dimensions = %dx%d, want 100x50", b.Dx(), b.Dy())
	}
}

func TestDecode_File(t *testing.T) {
	path := writePNG(t, testImage())
	img, format, err := Decode(path)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if format != "png" {
		t.Errorf("format = %q, want png", format)
	}
	if img.Bounds().Dx() != 100 {
		t.Errorf("width = %d, want 100", img.Bounds().Dx())
	}
}

func TestResize_Downscale(t *testing.T) {
	img := Resize(testImage(), 50, 0)
	b := img.Bounds()
	if b.Dx() != 50 || b.Dy() != 25 {
		t.Errorf("resized = %dx%d, want 50x25", b.Dx(), b.Dy())
	}
}

func TestResize_NoOp(t *testing.T) {
	img := Resize(testImage(), 0, 0)
	if img.Bounds().Dx() != 100 {
		t.Errorf("no-op resize changed width: %d", img.Bounds().Dx())
	}
}

func TestNextImageName(t *testing.T) {
	dir := t.TempDir()
	// empty dir -> img1.webp
	name, err := NextImageName(dir, "webp")
	if err != nil {
		t.Fatal(err)
	}
	if name != "img1.webp" {
		t.Errorf("name = %q, want img1.webp", name)
	}
	// existing img1..img3 -> img4
	for i := 1; i <= 3; i++ {
		if err := os.WriteFile(filepath.Join(dir, "img"+string(rune('0'+i))+".webp"), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	name, err = NextImageName(dir, "webp")
	if err != nil {
		t.Fatal(err)
	}
	if name != "img4.webp" {
		t.Errorf("name = %q, want img4.webp", name)
	}
	// unrelated files ignored
	if err := os.WriteFile(filepath.Join(dir, "cover.png"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	name, _ = NextImageName(dir, "webp")
	if name != "img4.webp" {
		t.Errorf("name = %q, want img4.webp (unrelated files ignored)", name)
	}
}

// TestWriteFileExclusive verifies O_EXCL semantics: first create succeeds,
// a second create of the same path fails with fs.ErrExist and leaves the
// original content untouched.
func TestWriteFileExclusive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "img1.webp")
	if err := WriteFileExclusive(path, []byte("a")); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	if err := WriteFileExclusive(path, []byte("b")); err == nil || !errors.Is(err, fs.ErrExist) {
		t.Fatalf("second create: err = %v, want fs.ErrExist", err)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "a" {
		t.Fatalf("existing content clobbered: %q, err = %v", got, err)
	}
}
