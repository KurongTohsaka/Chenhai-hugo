package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/themes/zhenhai"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "站点健康检查",
	Long:  "检测 config.yaml 语法、content/ 结构完整性、模板文件、缓存一致性等。",
	RunE: runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	root, _ := os.Getwd()
	return ExecuteDoctor(root)
}

// ExecuteDoctor runs the site health check from the given root directory.
func ExecuteDoctor(root string) error {
	fmt.Println("🔍 Chenhai Doctor — 站点健康检查")
	fmt.Println()

	errors := 0
	warnings := 0

	// 1. Check config.yaml exists
	configPath := filepath.Join(root, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("✗ config.yaml 不存在")
		fmt.Println("  → 提示：在站点根目录创建 config.yaml 配置文件")
		fmt.Println()
		errors++
	} else {
		// 1b. Check config.yaml syntax
		_, err := config.Load(configPath)
		if err != nil {
			fmt.Printf("✗ config.yaml 解析失败：%v\n", err)
			fmt.Println()
			errors++
		} else {
			fmt.Println("✓ config.yaml 语法正确")
		}
	}

	// 2. Check content/ directory
	contentDir := filepath.Join(root, "content")
	if info, err := os.Stat(contentDir); os.IsNotExist(err) || !info.IsDir() {
		fmt.Println("✗ content/ 目录不存在")
		fmt.Println("  → 提示：创建 content/ 目录并放入 .md 文件")
		fmt.Println()
		errors++
	} else {
		// 2b. Walk all .md files and validate front matter
		var mdFiles []string
		filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(path, ".md") {
				mdFiles = append(mdFiles, path)
			}
			return nil
		})

		if len(mdFiles) == 0 {
			fmt.Println("⚠ content/ 目录为空（没有 .md 文件）")
			warnings++
		} else {
			fmt.Printf("✓ content/ 目录存在（%d 个 .md 文件）\n", len(mdFiles))

			fmErrors := 0
			for _, path := range mdFiles {
				raw, err := os.ReadFile(path)
				if err != nil {
					fmt.Printf("✗ 无法读取 %s：%v\n", path, err)
					errors++
					continue
				}
				relPath, _ := filepath.Rel(root, path)
				_, _, err = content.ParseFrontMatter(raw)
				if err != nil {
					fmt.Printf("✗ %s\n  → Front Matter 解析失败：%v\n", relPath, err)
					fmErrors++
				}
			}
			if fmErrors > 0 {
				fmt.Printf("  共 %d 个文件 Front Matter 解析失败\n", fmErrors)
				errors += fmErrors
			}
		}
	}
	fmt.Println()

	// 3. Check required templates
	requiredTemplates := []string{"base.html", "single.html", "list.html", "index.html", "404.html"}
	missingTemplates := 0
	for _, name := range requiredTemplates {
		found := false
		for _, dir := range []string{
			filepath.Join(root, "layouts", name),
			filepath.Join(root, "themes", "zhenhai", "layouts", name),
		} {
			if _, err := os.Stat(dir); err == nil {
				found = true
				break
			}
		}
		// Check embedded zhenhai theme as fallback
		if !found {
			if _, err := zhenhai.FS.ReadFile("layouts/" + name); err == nil {
				found = true
			}
		}
		if !found {
			fmt.Printf("⚠ 模板 %s 未找到\n", name)
			warnings++
			missingTemplates++
		}
	}
	if missingTemplates == 0 {
		fmt.Println("✓ 所有必需模板存在")
	}
	fmt.Println()

	// 4. Check static/ directory
	staticDir := filepath.Join(root, "static")
	if info, err := os.Stat(staticDir); os.IsNotExist(err) || !info.IsDir() {
		fmt.Println("⚠ static/ 目录不存在（静态资源将跳过）")
		warnings++
	} else {
		fmt.Println("✓ static/ 目录存在")
	}
	fmt.Println()

	// 5. Check cache consistency
	cachePath := filepath.Join(root, ".chenhai-cache.json")
	if _, err := os.ReadFile(cachePath); err == nil {
		fmt.Println("✓ .chenhai-cache.json 存在")
	} else {
		fmt.Println("⚠ .chenhai-cache.json 不存在（首次构建后自动生成）")
		warnings++
	}
	fmt.Println()

	// Summary
	if errors > 0 || warnings > 0 {
		fmt.Printf("✗ 发现 %d 个错误，%d 个警告\n", errors, warnings)
	} else {
		fmt.Println("✓ 站点健康检查通过，一切正常")
	}

	if errors > 0 {
		return fmt.Errorf("发现 %d 个错误", errors)
	}
	return nil
}
