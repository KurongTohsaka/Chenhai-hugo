# Chenhai-hugo

> 水墨为底，金线点睛，大巧不工。

**Chenhai** 是一个用 Go 编写的轻量级静态博客生成器，灵感来自 Hugo 和碧蓝航线角色**镇海**（我的妻子）。

面向中文用户的简洁、高度自定义的博客工具。

## 特性

- **Typora 级 Markdown 支持** — GFM 表格/任务列表、数学公式（KaTeX）、Mermaid 图表、代码高亮（Chroma）、Admonition 提示框
- **镇海主题** — "水墨古风"美学，宣纸白/墨色/靛青/鎏金/朱砂配色，明暗双模，浮动 TOC
- **YAML 配置驱动** — 单一 `config.yaml` + Front Matter 覆盖，定义万物
- **分类 & 标签** — 层级分类 + 标签云（五级字号）
- **时间线归档** — 年月分组的折叠式归档页
- **前端搜索** — 构建时生成 JSON 索引，Ctrl+K 唤起搜索框
- **开发服务器** — 文件监听 + 自动重建 + LiveReload 浏览器推送

## 快速开始

### 安装

```bash
git clone https://github.com/KurongTohsaka/chenhai-hugo.git
cd chenhai-hugo
go build -o chenhai ./cmd/chenhai/
```

### 创建站点

```bash
mkdir my-blog && cd my-blog
cat > config.yaml << 'EOF'
title: "我的博客"
theme: "zhenhai"
themeConfig:
  colorMode: "auto"
  postsPerPage: 10
EOF

mkdir -p content/posts
```

### 写文章

```bash
chenhai new posts/hello-world.md
```

### 预览

```bash
chenhai serve
# → http://localhost:1313
```

### 构建

```bash
chenhai build
# → public/
```

## 项目结构

```
my-blog/
├── config.yaml        ← 站点配置
├── content/           ← Markdown 源文件
│   ├── posts/
│   └── about/
├── static/            ← 静态资源（不经处理）
├── themes/            ← 自定义主题覆盖（可选）
└── public/            ← 构建输出（自动生成）
```

## 命令

| 命令 | 说明 |
|---|---|
| `chenhai build` | 全量构建 → `public/` |
| `chenhai serve` | 开发服务器（`:1313`），LiveReload |
| `chenhai new <path>` | 新建文章，自动填充 Front Matter |
| `chenhai clean` | 清空 `public/` |
| `chenhai version` | 版本信息 |

## 技术栈

| 层 | 选型 |
|---|---|
| Markdown 解析 | Goldmark + highlighting + mathjax |
| 代码高亮 | Chroma |
| 模板引擎 | Go `html/template` |
| YAML | `gopkg.in/yaml.v3` |
| CLI | `spf13/cobra` |
| 文件监听 | `fsnotify` |
| WebSocket | `gorilla/websocket` |

## 开发

```bash
# 运行测试
go test ./internal/...

# 构建二进制
go build -o chenhai ./cmd/chenhai/

# 使用示例博客测试
cd testdata/example-blog
../../chenhai serve
```

## 许可证

MIT
