package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "输出版本号",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("chenhai v0.5.1")
	},
}
