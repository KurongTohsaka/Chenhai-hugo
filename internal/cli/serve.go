package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动开发服务器",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("serve command not yet implemented — coming in Task 9")
	},
}
