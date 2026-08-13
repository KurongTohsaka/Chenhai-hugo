package cli

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
		got, err := postToImgDir(c.post)
		if err != nil {
			t.Errorf("postToImgDir(%q) error: %v", c.post, err)
			continue
		}
		if got != c.want {
			t.Errorf("postToImgDir(%q) = %q, want %q", c.post, got, c.want)
		}
	}
	// 穿越输入必须拒绝（.. 段、绝对路径、越界 section）
	bad := []string{
		"a/../../etc/evil.md",
		"../posts/x.md",
		"../../escape.md",
		"/abs/path.md",
	}
	for _, p := range bad {
		if _, err := postToImgDir(p); err == nil {
			t.Errorf("postToImgDir(%q) = nil error, want rejection", p)
		}
	}
}

func TestCheckQuality(t *testing.T) {
	for _, q := range []int{-1, 101, 1000} {
		if err := checkQuality(q); err == nil {
			t.Errorf("checkQuality(%d) = nil, want error", q)
		}
	}
	for _, q := range []int{0, 50, 100} {
		if err := checkQuality(q); err != nil {
			t.Errorf("checkQuality(%d) = %v, want nil", q, err)
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

// siteRoot creates a temp site root with config.yaml and chdirs into it,
// restoring the previous working directory on test end.
func siteRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), []byte("baseURL: https://example.com\n"), 0644); err != nil {
		t.Fatal(err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatal(err)
		}
	})
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestRunImageAdd_PathTraversal verifies --dir/--post/--name traversal
// inputs are rejected before any file is written outside static/.
func TestRunImageAdd_PathTraversal(t *testing.T) {
	root := siteRoot(t)
	src := writePNG(t, testImage())

	cases := []struct {
		name string
		dir  string
		post string
		out  string
	}{
		{"dir traversal", "../../escape", "", ""},
		{"dir absolute", "/tmp/escape", "", ""},
		{"post traversal", "", "a/../../etc/evil.md", ""},
		{"post leading dotdot", "", "../posts/x.md", ""},
		{"name separator", "img/x", "", "a/b.webp"},
		{"name traversal", "img/x", "", "../../evil.webp"},
		{"name backslash", "img/x", "", `..\evil.webp`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := runImageAdd(src, c.post, c.dir, c.out, 80, false); err == nil {
				t.Fatalf("runImageAdd(%s) succeeded, want error", c.name)
			}
		})
	}
	// 站点根下未落任何穿越文件
	for _, p := range []string{"escape", "etc", filepath.Join("static", "img", "x", "evil.webp")} {
		if _, err := os.Stat(filepath.Join(root, p)); !os.IsNotExist(err) {
			t.Fatalf("traversal artifact escaped the site root: %s", p)
		}
	}
}

// TestRunImageAdd_QualityRange verifies the 0-100 quality contract at the
// add entry point.
func TestRunImageAdd_QualityRange(t *testing.T) {
	siteRoot(t)
	src := writePNG(t, testImage())
	for _, q := range []int{-1, 101, 1000} {
		if err := runImageAdd(src, "", "img/test", "", q, false); err == nil || !strings.Contains(err.Error(), "quality 必须在 0-100 之间") {
			t.Errorf("quality %d: err = %v, want quality range error", q, err)
		}
	}
}

// TestRunImageAdd_ConcurrentAutoName verifies that concurrent adds with
// auto-naming each land on a unique file (O_EXCL retry-increment), with no
// silent overwrites: n goroutines must produce exactly img1..imgN.
func TestRunImageAdd_ConcurrentAutoName(t *testing.T) {
	root := siteRoot(t)
	src := writePNG(t, testImage())

	const n = 4
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = runImageAdd(src, "", "img/conc", "", 80, false)
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: %v", i, err)
		}
	}
	// O_EXCL 保证 n 次并发各自拿到唯一名字：恰好 img1..imgN 各一
	dir := filepath.Join(root, "static", "img", "conc")
	for i := 1; i <= n; i++ {
		name := fmt.Sprintf("img%d.webp", i)
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}
