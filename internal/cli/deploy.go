package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var deployMsg string

func init() {
	deployCmd.Flags().StringVarP(&deployMsg, "message", "m", "", "提交信息")
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "提交并推送到远程仓库（CI 会负责构建）",
	Long:  "将 content/、static/、config.yaml 等源文件 add → commit → push。\n构建和部署由 GitHub Actions 在 CI 中完成。\n\n示例:\n  chenhai deploy -m \"新文章\"",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}

		if deployMsg == "" {
			return fmt.Errorf("请通过 -m 指定提交信息，如: chenhai deploy -m \"add: 新文章\"")
		}

		// Ensure we're in a git repo
		gitDir := filepath.Join(root, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			return fmt.Errorf("当前目录不是 Git 仓库")
		}

		run := func(name string, args ...string) error {
			cmd := exec.Command(name, args...)
			cmd.Dir = root
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}

		// git add source files only (not public/, not .chenhai-cache.json)
		fmt.Println("→ git add")
		addPaths := []string{"content", "static", "config.yaml", "archetypes", "themes", "README.md"}
		for _, p := range addPaths {
			fullPath := filepath.Join(root, p)
			if _, err := os.Stat(fullPath); err == nil {
				if err := run("git", "add", p); err != nil {
					return fmt.Errorf("git add %s: %w", p, err)
				}
			}
		}

		// git commit
		fmt.Printf("→ git commit -m \"%s\"\n", deployMsg)
		if err := run("git", "commit", "-m", deployMsg); err != nil {
			return fmt.Errorf("git commit: %w\n(可能没有需要提交的变更)", err)
		}

		// git push
		fmt.Println("→ git push")
		if err := run("git", "push"); err != nil {
			return fmt.Errorf("git push: %w", err)
		}

		fmt.Println("✓ 已推送，CI 将自动构建并部署")
		return nil
	},
}
