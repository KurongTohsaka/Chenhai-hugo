package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "清空构建输出",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := os.Getwd()
		public := filepath.Join(root, "public")
		if err := os.RemoveAll(public); err != nil {
			return fmt.Errorf("clean public/: %w", err)
		}
		fmt.Println("已清空 public/ 目录")
		return nil
	},
}
