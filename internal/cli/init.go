package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "创建新站点",
	Long:  "在当前目录（或指定路径）生成站点骨架。",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := os.Getwd()
		if len(args) > 0 {
			root = args[0]
		}
		os.MkdirAll(root, 0755)

		// Create directories
		for _, d := range []string{"content/posts", "static", "archetypes"} {
			os.MkdirAll(filepath.Join(root, d), 0755)
		}

		// Create config.yaml
		cfg := `title: "我的博客"
subtitle: "读书、记事、静观沧海"
theme: "zhenhai"
themeConfig:
  colorMode: "auto"
  postsPerPage: 10
menu:
  - name: "首页"
    url: "/"
  - name: "归档"
    url: "/archives/"
  - name: "标签"
    url: "/tags/"
`
		os.WriteFile(filepath.Join(root, "config.yaml"), []byte(cfg), 0644)

		// Create default archetype
		arch := `---
title: "{{.Title}}"
date: {{.Date}}
draft: false
categories: []
tags: []
---

`
		os.WriteFile(filepath.Join(root, "archetypes", "default.md"), []byte(arch), 0644)

		fmt.Printf("站点已创建: %s\n", root)
		return nil
	},
}
