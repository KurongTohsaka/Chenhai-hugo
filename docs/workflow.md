# Chenhai 博客工作流

> 从 Typora 写作到 GitHub Pages 部署的完整流程

当前测试站点：**https://hekurong.github.io**

## 简化的三步工作流

```
Typora 写作  →  chenhai deploy -m "msg"  →  CI 自动构建部署
   (.md)          (add + commit + push)        (GitHub Actions)
```

不再需要本地 `chenhai build`——构建交给 CI。

## 1. 环境要求

- Go 1.21+（编译 Chenhai）
- Git
- Typora 或任意 Markdown 编辑器
- GitHub 仓库（用于托管源码和 Pages 部署）

## 2. 项目结构

```
chenhai-site/                  ← GitHub 仓库根目录
├── config.yaml                ← 站点配置
├── content/                   ← Markdown 源文件（Typora 编辑）
│   └── posts/
│       ├── CS224N/
│       ├── DailyDev/
│       └── ...
├── static/                    ← 静态资源（不经 Chenhai 处理）
├── archetypes/                ← 内容原型
├── .github/workflows/         ← GitHub Actions CI 配置
└── public/                    ← CI 构建输出（已 gitignore）
```

## 3. 写作（Typora）

在 Typora 中打开 `content/posts/` 目录，直接编辑 `.md` 文件。

### 新建文章

```bash
chenhai new posts/分类/文件名.md --category "分类" --tags "标签1,标签2"
```

### Front Matter 模板

```yaml
---
title: "文章标题"
date: 2026-05-28
categories: ["分类"]
tags: ["标签1", "标签2"]
description: "文章描述"
toc: true
---
```

### 图片引用

将图片放入 `static/img/` 对应目录，Markdown 中使用绝对路径：

```markdown
![图片描述](/img/分类/文件名.png "可选标题")
```

### 提示框（Admonition）

```markdown
> [!note]
> 这是笔记/提示信息

> [!warning]
> 这是警告信息，需要注意

> [!tip]
> 实用小技巧

> [!danger]
> 危险/重要警告
```

## 4. 本地预览

```bash
cd chenhai-site

# 启动预览服务器（含自动构建 + LiveReload）
chenhai serve
# → http://localhost:1313
```

修改内容后自动重建，浏览器自动刷新。

## 5. 部署

写完文章后，一条命令搞定：

```bash
chenhai deploy -m "add: 新文章标题"
```

等价于手动执行：

```bash
git add content/ static/ config.yaml archetypes/ themes/
git commit -m "add: 新文章标题"
git push
```

推送后 CI 自动构建并部署到 GitHub Pages，1-2 分钟生效。

## 6. CI 配置

在仓库中创建 `.github/workflows/deploy.yml`：

```yaml
name: Deploy

on:
  push:
    branches: [main]

permissions:
  contents: write

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.21"

      - name: Build Chenhai
        run: |
          git clone https://github.com/KurongTohsaka/Chenhai-hugo.git /tmp/chenhai
          cd /tmp/chenhai && go build -o /usr/local/bin/chenhai ./cmd/chenhai/

      - name: Build site
        run: chenhai build

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./public
```

### GitHub Pages 设置

仓库 Settings → Pages：
- **Source**: `Deploy from a branch`
- **Branch**: `gh-pages` / `/ (root)`

### CI 运行机制

配置完成后，每次 push 会有 **两个 Action** 串联执行：

```
main push
    │
    ├─→ "Deploy"（我们的）      clone chenhai → build → 推 gh-pages     ~1min
    │
    └─→ "pages build"（内置）   从 gh-pages 发布静态文件到 CDN           ~30s
```

1. **Deploy**（`.github/workflows/deploy.yml`）——自定义 Action，编译 chenhai 后执行 `chenhai build`，将 `public/` 推送到 `gh-pages` 分支
2. **pages build and deployment**（GitHub 内置）——检测到 `gh-pages` 分支更新，将静态文件分发到全球 CDN

两个 Action 缺一不可。内置的 pages build 无法删除，但它只做纯文件分发（`gh-pages` 里只有 HTML/CSS/JS，不跑 Jekyll）。

### 常见问题

**两个 Action 都成功了但首页没更新**

检查文章 Front Matter 中是否有 `draft: true`。草稿不会被构建。

**pages build 报 Jekyll 错误**

说明 Pages Source 的 Branch 还指向 `main`。`main` 分支里有 `.md` 源文件，GitHub 会尝试用 Jekyll 处理。改为 `gh-pages` 即可。

**只想跑我们自己的 Action**

将 Pages Source 设为 `Deploy from a branch` + Branch `gh-pages`。内置 Action 只做文件分发，不会再尝试 Jekyll 构建。

## 7. 日常写作流程

```bash
# 1. Typora 写文章
open content/posts/DailyDev/新文章.md

# 2. 带图文章：截图一键转 WebP 并生成引用（v0.8.0+）
chenhai image add ~/Desktop/shot.png --post posts/DailyDev/新文章.md
# → 拷入 static/img/DailyDev/新文章/img1.webp
# → 输出 ![](/img/DailyDev/新文章/img1.webp) 粘贴进文章

# 3. 本地预览（可选）
chenhai serve

# 4. 确认无误，部署
chenhai deploy -m "add: 新文章"
```

推送后等待 1-2 分钟，https://hekurong.github.io 自动更新。

## 7.1 写作可用扩展（v0.8.0+）

- **贴图工作流**：`chenhai image add` 自动压缩（WebP）+ 归档 + 引用输出；`compress` 批量转换、`resize` 等比缩放。详见 [tutorial.md](tutorial.md) 第 11 章
- **Shortcode 组件**：`{{< details "标题" >}}` 折叠块、`{{< gallery >}}` 图片画廊、`{{< tabs "A" "B" >}}` 标签页（`=== 名 ===` 分块）。详见 tutorial.md 第 12 章
- **RSS 订阅**：站点已启用（baseURL + author 已配置），构建自动生成 atom.xml，订阅地址 https://hekurong.github.io/atom.xml

## 8. 常用操作

### 只修改配置

```bash
vim config.yaml
chenhai deploy -m "update: 修改站点配置"
```

### 添加图片

```bash
cp screenshot.png static/img/
chenhai deploy -m "add: 文章配图"
```

### 本地调试

```bash
chenhai serve                 # 本地预览
chenhai build                 # 本地构建（检查输出）
cat .chenhai-cache.json       # 查看增量缓存
```

### 强制全量重建（CI 或本地）

```bash
rm -rf public/ .chenhai-cache.json
chenhai build
```
