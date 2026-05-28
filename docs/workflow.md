# Chenhai 博客工作流

> 从 Typora 写作到 GitHub Pages 部署的完整流程

当前测试站点：**https://hekurong.github.io**

## 工作流总览

```
Typora 写作  →  chenhai build  →  git push  →  GitHub Actions  →  GitHub Pages
   (.md)          (public/)        (仓库)       (自动部署)         (线上站点)
```

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
│       ├── CS224N/            ← 分类目录
│       ├── DailyDev/
│       └── ...
├── static/                    ← 静态资源（不经过 Chenhai 处理）
│   ├── img/                   ← 文章配图
│   └── cover/                 ← 封面图
├── public/                    ← 构建输出（GitHub Pages 源）
├── .chenhai-cache.json        ← 增量构建缓存
└── .github/workflows/         ← GitHub Actions 配置
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

## 4. 本地预览

```bash
cd chenhai-site

# 构建
chenhai build

# 启动预览服务器
chenhai serve
# → http://localhost:1313
```

修改内容后自动重建，浏览器自动刷新。

## 5. 增量构建

首次构建全量处理所有文章，后续构建只重建有变更的文件：

```
首次: 扫描 → 发现 91 篇文章 → 渲染 91/91 → 完成（~2s）
二次: 扫描 → 发现 91 篇文章（91 跳过未变更）→ 完成（~0.3s）
```

配置文件或模板变更时自动触发全量重建。

## 6. 部署到 GitHub Pages

### GitHub Actions 配置

创建 `.github/workflows/deploy.yml`：

```yaml
name: Deploy Chenhai Site

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

      - name: Build Chenhai binary
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

### 推送部署

```bash
# 本地构建
chenhai build

# 提交并推送
git add -A
git commit -m "new: 文章标题"
git push origin main
```

推送后 GitHub Actions 自动：
1. 编译 Chenhai
2. 构建站点（`chenhai build`）
3. 将 `public/` 推送到 `gh-pages` 分支
4. GitHub Pages 自动更新

### GitHub Pages 设置

仓库 Settings → Pages：
- **Source**: Deploy from a branch
- **Branch**: `gh-pages` / `/ (root)`
- **URL**: `https://hekurong.github.io`

## 7. 日常写作流程

```bash
# 1. Typora 写文章
open content/posts/DailyDev/新文章.md

# 2. 本地预览
chenhai serve

# 3. 确认无误后提交
git add content/posts/DailyDev/新文章.md static/img/
git commit -m "add: 新文章"
git push
```

推送后等待 1-2 分钟，线上站点自动更新。

## 8. 常见操作

### 修改配置

```bash
vim config.yaml    # 修改站点配置
chenhai build      # 重建（自动检测配置变更，全量重建）
```

### 添加新分类

直接在 `content/posts/` 下创建新目录即可：

```bash
mkdir content/posts/新分类
chenhai new posts/新分类/文章.md -c "新分类" -t "tag1,tag2"
```

### 排查问题

```bash
# 查看构建日志
chenhai build 2>&1 | tee build.log

# 检查缓存状态
cat .chenhai-cache.json

# 强制全量重建
rm -rf public/ .chenhai-cache.json && chenhai build
```
