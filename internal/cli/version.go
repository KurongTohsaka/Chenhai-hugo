package cli

import (
	"github.com/spf13/cobra"
)

// Version is the current chenhai release version.
// Keep in sync with docs/planning.md and README.md.
const Version = "v0.8.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "输出版本号",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("chenhai " + Version)
	},
}
