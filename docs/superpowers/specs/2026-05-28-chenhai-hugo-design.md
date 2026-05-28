# Chenhai-hugo 设计规格文档

> 一个高度自定义的静态博客生成器，Go 语言实现，中文优先

## 1. 项目定位

**为什么不用 Hugo？** Hugo 过于庞杂，对中文社区支持较差。需要一个更简洁好用、中文优先的博客编译器。

**目标用户**：个人使用为主，兼顾小范围中文社区分享（开源方向）。

**交付形态**：单一二进制 `chenhai`，开箱即用。

## 2. 系统架构

整体式管道架构，六个核心模块：

```
chenhai binary
├── config     → YAML 站点配置解析
├── content    → Markdown 解析管道（Goldmark + 扩展）
├── theme      → Go html/template 主题引擎
├── build      → 静态文件生成（HTML/CSS/JS/资源）
├── serve      → 内置开发服务器（热重载、LiveReload）
└── cli        → build / serve / new / clean / version
```

**数据流向**：

```
.md 文件 → Goldmark 解析 → HTML 片段
Front Matter → YAML 解析 → PageMeta
PageMeta + HTML → html/template → 最终页面
所有页面 Meta → Index Builder → 搜索/分类/Tag/时间线
```

## 3. 内容渲染管道（Content Pipeline）

基于 Goldmark 生态，优先级从高到低：

| 优先级 | 特性 | 实现方式 |
|--------|------|---------|
| D | 代码高亮 | Goldmark-highlighting + Chroma（200+ 语言、行号、文件名标注、复制按钮） |
| B | 数学公式 | Goldmark-mathjax 扩展，输出 KaTeX 兼容 HTML，前端 KaTeX 渲染 |
| E | 多媒体增强 | 自定义 AST 后处理，图片包裹 figure、支持对齐、alt 作为图注；视频/音频嵌入 |
| C | Mermaid 图表 | 自定义 Goldmark 扩展，` ```mermaid` → `<pre class="mermaid">`，前端 Mermaid.js |
| A | GFM 基础 | 表格、任务列表、删除线（Goldmark 内置）+ 脚注（扩展） |
| F | Admonition | 自定义扩展，`> [!note]` → 样式化提示框 div |

**Front Matter**（参考 Hugo）：

```yaml
---
title: "文章标题"
date: 2026-05-28
lastmod: 2026-06-01
draft: false
categories: ["技术", "Go"]
tags: ["静态博客", "Goldmark"]
slug: "custom-url"
url: "/any/path/"
weight: 5
description: "SEO 描述"
summary: "列表摘要，不写则自动截取"
toc: true
math: false
---
```

## 4. 主题系统

### 4.1 镇海主题（内置主题、核心亮点）

**设计理念**：水墨为底、金线点睛、大巧不工。

镇海角色特质取自碧蓝航线：文静的秘书舰，常年处理文书，内敛从容但锋芒暗藏。远看是素雅古风博文，细看处处精致细节。

**配色体系**：

| 颜色 | 色值 | 用途 |
|------|------|------|
| 宣纸白 | #f7f4ef | 主背景 |
| 墨色 | #2c2c2c | 正文文字 |
| 靛青 | #1a3650 | 标题/导航 |
| 鎏金 | #b8962e | 链接/点缀 |
| 朱砂 | #8b1a2b | 强调/标记 |

支持明暗双模（auto/light/dark），暗色模式对应配色为墨底/古纸灰/霁蓝/暗金/胭脂。

主题中包含镇海角色图片资源。

### 4.2 主题架构

```
themes/
├── zhenhai/              ← 内置主题
│   ├── theme.yaml        ← 主题参数
│   ├── layouts/          ← Go html/template
│   │   ├── base.html
│   │   ├── index.html
│   │   ├── single.html
│   │   ├── list.html
│   │   ├── taxonomy.html
│   │   └── partials/
│   │       ├── header.html
│   │       ├── footer.html
│   │       ├── sidebar.html
│   │       ├── toc.html      ← 浮动目录组件
│   │       └── pagination.html
│   ├── assets/           ← CSS/JS/Images
│   ├── static/           ← 不经处理的静态文件
│   └── archetypes/       ← 内容原型
└── custom-theme/         ← 用户自定义（可选）
```

**优先级**：站点根目录模板 > 自定义主题 > 内置镇海主题

用户只需覆盖想定制的部分，其余自动回退到镇海主题。模板数据上下文：`.Page`、`.Site`、`.Config`。

## 5. YAML 配置系统

单一 `config.yaml` + Front Matter 覆盖，优先级：Front Matter > config.yaml > 内建默认值。

### 5.1 config.yaml 结构

```yaml
# 站点元信息
title: "镇海阁"
subtitle: "读书、记事、静观沧海"
description: "一个安静写字的地方"
baseURL: "https://example.com"
language: "zh-CN"
copyright: "CC BY-NC-SA 4.0"

# 作者
author:
  name: "指挥官"
  avatar: "/images/avatar.webp"

# 主题配置
theme: "zhenhai"
themeConfig:
  colorMode: "auto"       # light | dark | auto
  showToc: true
  tocFloat: true          # 浮动式目录
  codeTheme: "github-dark"
  dateFormat: "2006-01-02"
  postsPerPage: 10

# 导航
menu:
  - name: "首页";  url: "/";         icon: "home"
  - name: "归档";  url: "/archives/"; icon: "archive"
  - name: "标签";  url: "/tags/";     icon: "tag"
  - name: "关于";  url: "/about/";    icon: "info"

# Markdown 渲染
markup:
  highlight:
    style: "github-dark"
    lineNumbers: true
    showFilename: true
  math:
    engine: "katex"
  mermaid: true
  toc:
    minDepth: 2
    maxDepth: 4

# 社交链接
social:
  github: "https://github.com/..."
  email: "example@email.com"

# SEO
seo:
  googleAnalytics: "G-XXXXXXXX"
  enableRobotsTXT: true
  enableSitemap: true
```

## 6. 内容组织功能

### 6.1 分类（Categories）

- 一篇文章一个分类，支持层级（`categories: ["技术", "Go"]` → `/categories/技术/go/`）
- 自动生成分类聚合页 `/categories/`，展示所有分类及文章数

### 6.2 标签（Tags）

- 一篇文章多个标签，扁平结构
- `/tags/` 标签云，字号按文章数 5 级（xs→sm→md→lg→xl）
- `/tags/golang/` 展示标签下所有文章

### 6.3 时间线（Archives）

- `/archives/` 单页归档，按年→月分组
- 前端 JS 折叠展开交互

### 6.4 搜索

- 构建时生成 `search-index.json`（标题/URL/正文前500字/摘要/标签/分类/日期）
- 前端 Fuse.js 或自实现 Trie 模糊搜索
- `Ctrl+K` 唤起搜索框

## 7. CLI 命令

| 命令 | 作用 |
|------|------|
| `chenhai build` | 全量构建，输出到 `public/`（`--drafts` 包含草稿） |
| `chenhai serve` | 开发服务器（默认 localhost:1313），文件监听 + LiveReload（`--port`） |
| `chenhai new <path>` | 创建新文章，自动填充 front matter（`--kind page` 创建独立页面） |
| `chenhai clean` | 清空 `public/` |
| `chenhai version` | 版本号 |

## 8. 目录结构

### 8.1 用户项目

```
my-blog/
├── config.yaml
├── content/           ← Markdown 源文件
│   ├── posts/
│   └── about/
├── static/            ← 不经处理的静态文件
├── archetypes/        ← 内容原型（可选）
├── themes/            ← 自定义主题覆盖（可选）
└── public/            ← 构建输出（.gitignore）
```

### 8.2 chenhai 自身

```
Chenhai-hugo/
├── cmd/chenhai/main.go
├── internal/
│   ├── config/        ← YAML 配置解析
│   ├── content/       ← Markdown 管道
│   ├── theme/         ← 模板引擎
│   ├── build/         ← 构建编排
│   ├── index/         ← 分类/标签/搜索索引
│   └── server/        ← 开发服务器
├── themes/zhenhai/    ← 内置主题（嵌入二进制）
├── go.mod
└── go.sum
```

## 9. 技术选型

| 层面 | 选型 |
|------|------|
| Markdown 解析 | Goldmark + 扩展（highlighting, mathjax, 自定义） |
| 代码高亮 | Chroma |
| 模板引擎 | Go html/template |
| 搜索（后端） | 构建时 JSON 索引生成 |
| 搜索（前端） | Fuse.js 或自实现 Trie |
| 热重载 | fsnotify 文件监听 + WebSocket LiveReload |
| 命令行 | spf13/cobra（标准 Go CLI 框架） |
| YAML 解析 | gopkg.in/yaml.v3 |

## 10. 明确不支持

- 评论系统
- 多语言（i18n）
- RSS/Atom Feed（暂不考虑，可后续扩展）
