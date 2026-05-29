package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	buf := new(bytes.Buffer)
	versionCmd.SetOut(buf)
	versionCmd.SetErr(buf)
	versionCmd.Run(versionCmd, nil)

	out := buf.String()
	if !strings.Contains(out, "chenhai v") {
		t.Fatalf("expected version output to contain 'chenhai v', got: %s", out)
	}
}
