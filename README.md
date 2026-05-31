# Chenhai-hugo

> 水墨为底，金线点睛，大巧不工。

面向中文用户的轻量级静态博客生成器，Go 编写。灵感来自 Hugo 和碧蓝航线角色**镇海**。

测试站点：**https://hekurong.github.io**

## 特性

- **Typora 级 Markdown** — GFM 表格/任务列表、KaTeX 数学公式、Mermaid 图表、Chroma 代码高亮、Admonition 提示框
- **镇海主题** — 水墨古风，宣纸白/墨色/靛青/鎏金/朱砂配色，明暗双模，浮动 TOC，JetBrains Mono 代码字体
- **YAML 配置** — 单文件 `config.yaml` + Front Matter 覆盖
- **分类 & 标签** — 层级分类 + 标签云（五级字号）
- **时间线归档** — 年月分组的折叠式归档页
- **前端搜索** — JSON 索引，Ctrl+K 唤起
- **外部主题** — `themes/` 目录加载，缺模板自动回退镇海
- **增量构建** — SHA256 缓存，只重建变更页面
- **GitHub 图床** — auto 上传 + map 映射双模式
- **图片增强** — `<figure>` + `<figcaption>` + 对齐 + 懒加载
- **CI 部署** — `chenhai deploy` 推送，GitHub Actions 自动构建

## 安装

```bash
git clone https://github.com/KurongTohsaka/chenhai-hugo.git
cd chenhai-hugo && go build -o chenhai ./cmd/chenhai/
```

## 快速开始

```bash
chenhai init my-blog && cd my-blog
chenhai new posts/hello.md -c "生活" -t "博客"
chenhai serve    # → http://localhost:1313
```

## 命令

| 命令 | 说明 |
|---|---|
| `chenhai init [path]` | 创建站点骨架 |
| `chenhai new <path>` | 新建文章（`-c` 分类 `-t` 标签） |
| `chenhai serve` | 开发服务器 + LiveReload |
| `chenhai build` | 构建站点 → `public/` |
| `chenhai deploy -m "msg"` | 推送源码（CI 自动构建部署） |
| `chenhai new theme <name>` | 创建外部主题骨架 |
| `chenhai clean` | 清空 `public/` |
| `chenhai version` | 版本信息 |

## 工作流

```
Typora 写 .md  →  chenhai deploy -m "msg"  →  GitHub Actions 构建  →  Pages 部署
```

## 项目结构

```
my-blog/
├── config.yaml
├── content/posts/     ← Markdown 源文件
├── static/            ← 不经处理的静态资源
├── themes/            ← 外部主题（可选）
├── .github/workflows/ ← CI 配置
└── public/            ← 构建输出（gitignored）
```

## 技术栈

| 层 | 选型 |
|---|---|
| Markdown | Goldmark + highlighting + mathjax |
| 代码高亮 | Chroma（CSS class 模式，亮暗双适配） |
| 模板 | Go `html/template` |
| CLI | `spf13/cobra` |
| 文件监听 | `fsnotify` |
| LiveReload | `gorilla/websocket` |

## 版本

当前 **v0.6.1**。详见 [规划文档](docs/planning.md)。

## 文档

| 文档 | 说明 |
|------|------|
| [tutorial.md](docs/tutorial.md) | 使用教程 |
| [workflow.md](docs/workflow.md) | 工作流（写作 → 部署） |
| [planning.md](docs/planning.md) | 版本规划 |

## 开发

```bash
go test ./internal/...
go build -o chenhai ./cmd/chenhai/
cd testdata/example-blog && ../../chenhai serve
```

## 许可证

MIT
