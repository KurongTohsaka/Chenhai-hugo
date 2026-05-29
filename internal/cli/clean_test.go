package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanCmd(t *testing.T) {
	dir := t.TempDir()
	public := filepath.Join(dir, "public")
	if err := os.MkdirAll(public, 0755); err != nil {
		t.Fatal(err)
	}
	dummy := filepath.Join(public, "index.html")
	if err := os.WriteFile(dummy, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(dir)

	cleanCmd.RunE(cleanCmd, nil)

	if _, err := os.Stat(public); !os.IsNotExist(err) {
		t.Fatal("expected public/ to be removed")
	}
}
