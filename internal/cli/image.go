package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KurongTohsaka/chenhai-hugo/internal/imageproc"
	"github.com/spf13/cobra"
)

func newImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "图片处理（压缩 / 缩放 / 贴图）",
	}
	cmd.AddCommand(newImageCompressCmd())
	cmd.AddCommand(newImageResizeCmd())
	cmd.AddCommand(newImageAddCmd())
	return cmd
}

// collectImages walks path (file or dir) and returns transcodable image files.
func collectImages(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	var files []string
	err = filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() && imageproc.IsTranscodable(filepath.Ext(p)) {
			files = append(files, p)
		}
		return nil
	})
	return files, err
}

func newImageCompressCmd() *cobra.Command {
	var quality int
	cmd := &cobra.Command{
		Use:   "compress <file|dir>",
		Short: "将 jpg/png 压缩为 WebP（保留原文件）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			files, err := collectImages(args[0])
			if err != nil {
				return fmt.Errorf("collect images: %w", err)
			}
			converted, skipped := 0, 0
			for _, f := range files {
				ext := strings.ToLower(filepath.Ext(f))
				if ext == ".webp" {
					skipped++
					continue
				}
				img, _, err := imageproc.Decode(f)
				if err != nil {
					fmt.Printf("  ✗ %s: %v\n", f, err)
					skipped++
					continue
				}
				data, err := imageproc.EncodeWebP(img, quality)
				if err != nil {
					fmt.Printf("  ✗ %s: %v\n", f, err)
					skipped++
					continue
				}
				out := strings.TrimSuffix(f, filepath.Ext(f)) + ".webp"
				if err := imageproc.WriteFile(out, data); err != nil {
					return err
				}
				fmt.Printf("  ✓ %s → %s\n", f, out)
				converted++
			}
			fmt.Printf("完成：转换 %d 张，跳过 %d 张\n", converted, skipped)
			return nil
		},
	}
	cmd.Flags().IntVar(&quality, "quality", 80, "WebP 质量 0-100")
	return cmd
}

func newImageResizeCmd() *cobra.Command {
	var width, height int
	cmd := &cobra.Command{
		Use:   "resize <file>",
		Short: "等比缩放图片（输出 <name>_w<W>.jpg，不覆盖原文件）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if width <= 0 && height <= 0 {
				return fmt.Errorf("请指定 --width 或 --height")
			}
			img, _, err := imageproc.Decode(args[0])
			if err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			out := imageproc.Resize(img, width, height)
			ext := filepath.Ext(args[0])
			suffix := fmt.Sprintf("_w%d", width)
			if width <= 0 {
				suffix = fmt.Sprintf("_h%d", height)
			}
			outPath := strings.TrimSuffix(args[0], ext) + suffix + ext
			// Encode back to the original container format (webp stays webp)
			switch strings.ToLower(ext) {
			case ".webp":
				data, err := imageproc.EncodeWebP(out, 80)
				if err != nil {
					return err
				}
				err = imageproc.WriteFile(outPath, data)
				if err != nil {
					return err
				}
			default:
				if err := imageproc.WriteImage(outPath, out, ext); err != nil {
					return err
				}
			}
			fmt.Printf("  ✓ %s → %s\n", args[0], outPath)
			return nil
		},
	}
	cmd.Flags().IntVar(&width, "width", 0, "目标宽度（等比缩放）")
	cmd.Flags().IntVar(&height, "height", 0, "目标高度（等比缩放）")
	return cmd
}

func newImageAddCmd() *cobra.Command {
	var post, dir, name string
	var quality int
	var force bool
	cmd := &cobra.Command{
		Use:   "add <file>",
		Short: "贴图：压缩为 WebP 拷入 static/ 并输出 md 引用",
		Long: "将截图压缩为 WebP 并按文章目录自动命名，输出可直接粘贴的\n" +
			"Markdown 引用路径。示例：\n" +
			"  chenhai image add ~/Desktop/shot.png --post posts/CS224N/lesson_5.md\n" +
			"  → static/img/CS224N/lesson_5/img3.webp\n" +
			"  → ![](/img/CS224N/lesson_5/img3.webp)",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImageAdd(args[0], post, dir, name, quality, force)
		},
	}
	cmd.Flags().StringVar(&post, "post", "", "文章路径（相对 content/，如 posts/CS224N/lesson_5.md）")
	cmd.Flags().StringVar(&dir, "dir", "", "目标目录（相对 static/，如 img/xxx）")
	cmd.Flags().StringVar(&name, "name", "", "输出文件名（默认 imgN.webp 自动递增）")
	cmd.Flags().IntVar(&quality, "quality", 80, "WebP 质量 0-100")
	cmd.Flags().BoolVar(&force, "force", false, "覆盖已存在的同名文件")
	return cmd
}

// postToImgDir maps a content-relative post path to a static/img-relative
// directory, stripping the leading section segment to match the site's
// /img/<分类>/<文章>/ convention:
//
//	content/posts/CS224N/lesson_5.md → img/CS224N/lesson_5
//	about/index.md                   → img/about
func postToImgDir(post string) string {
	rel := strings.TrimSuffix(post, ".md")
	if idx := strings.Index(rel, "/"); idx >= 0 {
		rel = rel[idx+1:]
	}
	rel = strings.TrimSuffix(rel, "/index")
	return filepath.Join("img", rel)
}

func runImageAdd(src, post, dir, name string, quality int, force bool) error {
	if (post == "") == (dir == "") {
		return fmt.Errorf("必须且只能指定 --post 或 --dir 之一")
	}
	// Y12: 站点根检测（循 chenhai build 约定），防在任意目录静默创建 static/
	if _, err := os.Stat("config.yaml"); err != nil {
		return fmt.Errorf("未找到 config.yaml——请在站点根目录运行 image add")
	}

	// 1. Resolve source
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("源文件: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("源文件不能是目录")
	}

	// 2. Resolve target dir (relative to static/)
	var relDir string
	if post != "" {
		relDir = postToImgDir(post)
	} else {
		relDir = dir
	}
	staticDir := filepath.Join("static", relDir)

	// 3. Determine output name
	ext := strings.ToLower(filepath.Ext(src))
	// GIF passes through unchanged; others transcode to webp.
	outExt := "webp"
	if ext == ".gif" {
		outExt = "gif"
	}
	outName := name
	if outName == "" && !force {
		outName, err = imageproc.NextImageName(staticDir, outExt)
		if err != nil {
			return err
		}
	}
	if outName == "" {
		outName = "img1." + outExt // force mode with no name: explicit default
	}
	if ext := filepath.Ext(outName); ext != "" && ext != "."+outExt {
		// --name 带了不匹配的扩展名：强制替换为实际输出格式（防 WebP 数据写进 .jpg）
		outName = strings.TrimSuffix(outName, ext) + "." + outExt
	} else if ext == "" {
		outName += "." + outExt
	}
	outPath := filepath.Join(staticDir, outName)

	// 4. Process: gif copy / transcode
	if ext == ".gif" {
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		if err := imageproc.WriteFile(outPath, data); err != nil {
			return err
		}
	} else {
		img, _, err := imageproc.Decode(src)
		if err != nil {
			return fmt.Errorf("decode %s: %w", src, err)
		}
		data, err := imageproc.EncodeWebP(img, quality)
		if err != nil {
			return fmt.Errorf("encode webp: %w", err)
		}
		if err := imageproc.WriteFile(outPath, data); err != nil {
			return err
		}
	}

	// 5. Output reference line
	ref := fmt.Sprintf("![](/%s/%s)", relDir, outName)
	fmt.Printf("  ✓ %s\n\n%s\n", outPath, ref)
	return nil
}
