package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/KurongTohsaka/chenhai-hugo/internal/build"
	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/internal/theme"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "构建站点",
	Long:  "解析 content/ 目录中的所有 Markdown 文件并生成静态站点到 public/ 目录。",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		cfg, err := config.Load(filepath.Join(root, "config.yaml"))
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		renderer := content.NewRenderer()
		engine, err := theme.New(cfg, root)
		if err != nil {
			return fmt.Errorf("init theme: %w", err)
		}

		builder := build.New(cfg, root, renderer, engine)
		if err := builder.Build(); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}

		fmt.Println("站点构建完成 → public/")
		return nil
	},
}
