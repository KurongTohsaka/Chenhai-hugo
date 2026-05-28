package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var newCategory string
var newTags string

func init() {
	newCmd.Flags().StringVarP(&newCategory, "category", "c", "", "分类")
	newCmd.Flags().StringVarP(&newTags, "tags", "t", "", "标签（逗号分隔）")
}

var newCmd = &cobra.Command{
	Use:   "new <path>",
	Short: "创建新文章",
	Long:  "基于 archetypes/default.md 原型创建新 Markdown 文件。\n示例: chenhai new posts/my-post.md",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := os.Getwd()
		relPath := args[0]

		// Ensure path has .md extension
		if !strings.HasSuffix(relPath, ".md") {
			relPath += ".md"
		}

		fullPath := filepath.Join(root, "content", relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}

		if _, err := os.Stat(fullPath); err == nil {
			return fmt.Errorf("文件已存在: %s", relPath)
		}

		// Generate front matter from archetype or default
		now := time.Now().Format("2006-01-02")
		title := strings.TrimSuffix(filepath.Base(relPath), ".md")

		// Build categories list
		catList := "[]"
		if newCategory != "" {
			cats := strings.Split(newCategory, ",")
			for i, c := range cats {
				cats[i] = `"` + strings.TrimSpace(c) + `"`
			}
			catList = "[" + strings.Join(cats, ", ") + "]"
		}

		// Build tags list
		tagList := "[]"
		if newTags != "" {
			tags := strings.Split(newTags, ",")
			for i, t := range tags {
				tags[i] = `"` + strings.TrimSpace(t) + `"`
			}
			tagList = "[" + strings.Join(tags, ", ") + "]"
		}

		content := fmt.Sprintf(`---
title: "%s"
date: %s
draft: true
categories: %s
tags: %s
---

`, title, now, catList, tagList)

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("write file: %w", err)
		}

		fmt.Printf("已创建: content/%s\n", relPath)
		return nil
	},
}
