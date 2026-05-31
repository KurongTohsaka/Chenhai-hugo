package cli

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "输出版本号",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("chenhai v0.6.1")
	},
}
