package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

var themeCmd = &cobra.Command{
	Use:   "theme <name>",
	Short: "创建新主题",
	Long: `生成一个完整的新主题骨架到 themes/<name>/ 目录。

主题名必须为 snake_case 或 kebab-case 格式（仅小写字母、数字、下划线和连字符）。`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Validate name: snake_case or kebab-case
		matched, _ := regexp.MatchString(`^[a-z][a-z0-9_-]*$`, name)
		if !matched {
			return fmt.Errorf("主题名只能包含小写字母、数字、下划线(_)和中划线(-)，且必须以字母开头")
		}

		root, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("获取工作目录失败: %w", err)
		}
		themeDir := filepath.Join(root, "themes", name)

		// Check if theme already exists
		if _, err := os.Stat(themeDir); err == nil {
			return fmt.Errorf("主题已存在: themes/%s/", name)
		}

		// Create directory structure
		dirs := []string{
			"layouts",
			"layouts/partials",
			"assets/css",
			"assets/js",
			"assets/images",
			"static",
		}
		for _, d := range dirs {
			if err := os.MkdirAll(filepath.Join(themeDir, d), 0755); err != nil {
				return fmt.Errorf("创建目录失败 %s: %w", d, err)
			}
		}

		// Create theme.yaml
		themeYAML := fmt.Sprintf(`name: "%s"
version: "1.0.0"
description: "A custom Chenhai theme"
author: ""
params: {}
`, name)
		if err := os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(themeYAML), 0644); err != nil {
			return fmt.Errorf("写入 theme.yaml 失败: %w", err)
		}

		// Create base.html — minimal HTML shell
		baseHTML := `<!DOCTYPE html>
<html lang="{{or .Config.Language "zh-CN"}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Config.Title}}{{if .Page}} - {{.Page.Title}}{{end}}</title>
    <link rel="stylesheet" href="/assets/css/style.css">
</head>
<body>
    <header>
        <h1><a href="/">{{.Config.Title}}</a></h1>
        <nav>
            {{range .Config.Menu}}
            <a href="{{.URL}}">{{.Name}}</a>
            {{end}}
        </nav>
    </header>
    <main>
        {{block "content" .}}{{end}}
    </main>
    <footer>
        <p>&copy; {{.Config.Copyright}}</p>
    </footer>
</body>
</html>`
		if err := os.WriteFile(filepath.Join(themeDir, "layouts", "base.html"), []byte(baseHTML), 0644); err != nil {
			return fmt.Errorf("写入 base.html 失败: %w", err)
		}

		// Create index.html — homepage content block
		indexHTML := `{{define "content"}}
<h2>Recent Posts</h2>
{{$pages := .Extra.pages}}
{{if $pages}}
<ul>
    {{range $pages}}
    <li>
        <a href="{{.Permalink}}">{{.Title}}</a>
        <time datetime="{{.Date.Format "2006-01-02"}}">{{.Date.Format "2006-01-02"}}</time>
    </li>
    {{end}}
</ul>
{{else}}
<p>No posts yet.</p>
{{end}}
{{end}}`
		if err := os.WriteFile(filepath.Join(themeDir, "layouts", "index.html"), []byte(indexHTML), 0644); err != nil {
			return fmt.Errorf("写入 index.html 失败: %w", err)
		}

		// Create style.css — minimal stylesheet
		themeName := name
		css := `/* Theme: ` + themeName + ` */
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: system-ui, sans-serif; line-height: 1.6; color: #333; max-width: 800px; margin: 0 auto; padding: 1rem; }
a { color: #0066cc; text-decoration: none; }
a:hover { text-decoration: underline; }
header { border-bottom: 1px solid #ddd; margin-bottom: 2rem; padding-bottom: 1rem; }
header h1 { display: inline; }
nav { display: inline; margin-left: 1rem; }
nav a { margin-right: 0.5rem; }
footer { border-top: 1px solid #ddd; margin-top: 2rem; padding-top: 1rem; font-size: 0.9rem; color: #666; }
`
		if err := os.WriteFile(filepath.Join(themeDir, "assets", "css", "style.css"), []byte(css), 0644); err != nil {
			return fmt.Errorf("写入 style.css 失败: %w", err)
		}

		fmt.Printf("已创建主题: themes/%s/\n", name)
		return nil
	},
}
