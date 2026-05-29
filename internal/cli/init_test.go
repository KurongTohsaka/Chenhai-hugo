package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCmd(t *testing.T) {
	dir := t.TempDir()

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(dir)

	target := filepath.Join(dir, "blog")
	if err := initCmd.RunE(initCmd, []string{target}); err != nil {
		t.Fatal(err)
	}

	for _, p := range []string{
		"config.yaml",
		"content/posts",
		"archetypes/default.md",
		"static",
	} {
		if _, err := os.Stat(filepath.Join(target, p)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", p)
		}
	}
}
