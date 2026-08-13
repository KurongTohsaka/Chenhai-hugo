package cli

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPostToImgDir(t *testing.T) {
	cases := []struct{ post, want string }{
		{"posts/CS224N/lesson_5.md", "img/CS224N/lesson_5"},
		{"posts/hello.md", "img/hello"},
		{"about/index.md", "img/about"},
		{"posts/DeepDive/README.md", "img/DeepDive/README"},
	}
	for _, c := range cases {
		if got := postToImgDir(c.post); got != c.want {
			t.Errorf("postToImgDir(%q) = %q, want %q", c.post, got, c.want)
		}
	}
}

// testImage builds a small RGBA image for transcoding tests.
func testImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 20, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
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

// TestRunImageAdd_OverwriteGuard verifies --force semantics: an existing
// output file is refused without --force and overwritten with --force.
func TestRunImageAdd_OverwriteGuard(t *testing.T) {
	// Site root with config.yaml (runImageAdd requires it)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), []byte("baseURL: https://example.com\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Source PNG and target dir
	src := writePNG(t, testImage())
	staticDir := filepath.Join(root, "static", "img", "test")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Pre-existing output file with sentinel content
	outPath := filepath.Join(staticDir, "shot.webp")
	if err := os.WriteFile(outPath, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	// Non-force with explicit --name must refuse to overwrite
	err = runImageAdd(src, "", "img/test", "shot.webp", 80, false)
	if err == nil || !strings.Contains(err.Error(), "输出文件已存在") {
		t.Fatalf("non-force overwrite: err = %v, want 输出文件已存在 error", err)
	}
	if got, _ := os.ReadFile(outPath); string(got) != "old" {
		t.Fatalf("existing file was clobbered without --force: %q", got)
	}

	// Force overwrites successfully
	if err := runImageAdd(src, "", "img/test", "shot.webp", 80, true); err != nil {
		t.Fatalf("force overwrite failed: %v", err)
	}
	if got, _ := os.ReadFile(outPath); string(got) == "old" {
		t.Fatal("force overwrite did not replace the file")
	}
}
