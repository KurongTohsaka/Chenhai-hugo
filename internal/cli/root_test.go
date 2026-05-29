package cli

import (
	"testing"
)

func TestRootCmd_Subcommands(t *testing.T) {
	subs := rootCmd.Commands()
	names := make(map[string]bool)
	for _, c := range subs {
		names[c.Name()] = true
	}

	for _, want := range []string{"build", "serve", "new", "clean", "version", "init", "deploy"} {
		if !names[want] {
			t.Errorf("expected subcommand %q to be registered", want)
		}
	}
}

func TestRootCmd_Help(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}
