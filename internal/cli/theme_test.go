package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestThemeCmd(t *testing.T) {
	dir := t.TempDir()

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(dir)

	if err := themeCmd.RunE(themeCmd, []string{"my-theme"}); err != nil {
		t.Fatal(err)
	}

	themeDir := filepath.Join(dir, "themes", "my-theme")
	for _, p := range []string{
		"theme.yaml",
		"layouts/base.html",
		"layouts/index.html",
		"assets/css/style.css",
	} {
		if _, err := os.Stat(filepath.Join(themeDir, p)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", p)
		}
	}
}

func TestThemeCmd_InvalidName(t *testing.T) {
	dir := t.TempDir()

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(dir)

	err := themeCmd.RunE(themeCmd, []string{"Bad-Name"})
	if err == nil {
		t.Fatal("expected error for invalid theme name with uppercase")
	}
}

func TestThemeCmd_Duplicate(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "themes", "existing"), 0755)

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(dir)

	err := themeCmd.RunE(themeCmd, []string{"existing"})
	if err == nil {
		t.Fatal("expected error for duplicate theme")
	}
}
