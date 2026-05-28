package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/server"
)

var port int

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 1313, "服务器端口")
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动开发服务器",
	Long:  "启动开发服务器，监听文件变化并自动重新构建，通过 LiveReload 推送更新到浏览器。",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		cfg, err := config.Load(filepath.Join(root, "config.yaml"))
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		srv := server.New(cfg, root)
		return srv.Start(port)
	},
}
