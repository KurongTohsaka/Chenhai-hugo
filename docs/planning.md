# Chenhai 规划文档

当前版本：**v0.1.0**

---

## 已完成

### 核心管道

- [x] **配置系统** — YAML 加载 + 默认值合并 + Front Matter 覆盖
- [x] **Markdown 渲染** — Goldmark + Chroma 高亮 + KaTeX 数学公式 + GFM 全支持
- [x] **模板引擎** — Go html/template + 嵌入式镇海主题 + 站点模板覆盖
- [x] **构建编排** — 内容采集 → 索引构建 → 页面渲染 → 静态资源复制 → SEO 文件生成
- [x] **开发服务器** — 文件监听（fsnotify）+ 自动重建 + WebSocket LiveReload
- [x] **CLI 工具** — build / serve / new / clean / version（基于 Cobra）

### 内容组织

- [x] **分类** — 层级分类，自动 `/categories/` 路径生成
- [x] **标签** — 扁平标签，标签云页面
- [x] **时间线归档** — 年/月分组 `/archives/`
- [x] **搜索** — 构建时生成 `search-index.json`，前端模糊搜索 + Ctrl+K 唤起

### 主题

- [x] **镇海主题** — Chinese ink-wash 风格，宣纸白/墨色/靛青/鎏金/朱砂配色
- [x] **暗色模式** — auto/light/dark 三态切换，JavaScript 交互
- [x] **浮动 TOC** — 文章页右侧固定目录，IntersectionObserver 跟踪当前标题
- [x] **代码块增强** — Copy 按钮、行号、文件名标注
- [x] **响应式布局** — 移动端单栏、桌面端侧边栏

---

## 当前问题

### 功能缺陷

| 编号 | 问题 | 影响 | 优先级 |
|------|------|------|--------|
| B-01 | **Admonition 扩展未实现** — `> [!note]` 以 blockquote 原样渲染 | `> [!note]` 语法失效 | 🔴 高 |
| B-02 | **分页未生效** — `postsPerPage` 配置项存在但构建器未使用 | 首页列出全部文章 | 🔴 高 |
| B-03 | **标签云字号未传递** — `BuildTagCloud()` 结果未注入模板 Extra | 标签页字号无差异 | 🟡 中 |
| B-04 | **分类/标签 URL 中文编码** — `/categories/%e6%8a%80%e6%9c%af/` | URL 不美观 | 🟡 中 |

### 质量缺口

| 编号 | 问题 | 优先级 |
|------|------|--------|
| Q-01 | `internal/cli/` 无测试 | 🟡 中 |
| Q-02 | `internal/server/` 无测试 | 🟡 中 |
| Q-03 | 搜索 UI 交互未实际验证 | 🟡 中 |
| Q-04 | 暗色模式 JS 未实际验证 | 🟡 中 |
| Q-05 | reading time 估算偏大（中文按 `/3` 计算） | 🟢 低 |
| Q-06 | 无 favicon | 🟢 低 |

---

## 版本规划

### v0.2.0 — 功能完善

目标：修复已知缺陷，让当前承诺的功能真正可用。

| 任务 | 说明 |
|------|------|
| Fix B-01 | 实现 Admonition Goldmark 自定义扩展（note/warning/tip/danger） |
| Fix B-02 | 实现分页渲染：首页 + 列表页按 `postsPerPage` 分页 |
| Fix B-03 | 标签云数据传递到模板，渲染 5 级字号 |
| Fix B-04 | 分类/标签路径 slug 化（中文 → 拼音或自定义 slug） |
| 测试补齐 | CLI 包集成测试、Server 包单元测试 |
| favicon | 添加镇海主题专属 favicon |

### v0.3.0 — 内容增强

目标：完善内容创作和阅读体验。

| 任务 | 说明 |
|------|------|
| RSS/Atom | 生成 Feed 文件 |
| 图片处理 | 构建时自动缩放、WebP 转换、响应式 srcset |
| 阅读时间 | 优化中英文混合字数统计算法 |
| 相关文章 | 基于标签/分类推荐相关文章 |
| 上一篇/下一篇 | 文章间导航（目前已有基础实现，需完善同类目导航） |
| SEO 增强 | Open Graph meta 标签、JSON-LD 结构化数据 |

### v0.4.0 — 主题生态

目标：让主题可插拔、可共享。

| 任务 | 说明 |
|------|------|
| 外部主题加载 | 扫描站点 `themes/` 目录，动态加载用户主题 |
| 主题参数 | `theme.yaml` 自定义参数传递到模板 |
| 多主题预览 | `chenhai serve --theme my-theme` |
| 主题脚手架 | `chenhai new theme <name>` 创建主题骨架 |

### v1.0.0 — 稳定发布

目标：生产级质量，面向中文社区正式发布。

| 任务 | 说明 |
|------|------|
| 增量构建 | 只重建有变更的文件，提升大站点构建速度 |
| i18n | 多语言支持（至少中/英） |
| 性能优化 | 模板缓存、并发渲染、内存使用优化 |
| 完整测试 | >80% 测试覆盖率 |
| 官方文档站 | 使用 Chenhai 自身搭建 `chenhai.dev` |

---

## 远期愿景

| 方向 | 说明 |
|------|------|
| 插件系统 | Lua 或 Starlark 脚本插件，扩展构建生命周期钩子 |
| 在线 IDE | Web 端在线编辑 + 实时预览 |
| 评论集成 | 可选的 Giscus / Waline / Twikoo 评论系统支持 |
| 图片 CDN | 自动上传图片到图床/CDN 并替换链接 |
| 多站点 | 单一工作区管理多个博客站点 |

---

## 贡献指南

欢迎提交 Issue 和 PR。开发前请阅读：

- [设计规格文档](superpowers/specs/2026-05-28-chenhai-hugo-design.md)
- [实现计划](superpowers/plans/2026-05-28-chenhai-hugo-implementation.md)

### 开发环境

```bash
git clone https://github.com/KurongTohsaka/chenhai-hugo.git
cd chenhai-hugo
go build -o chenhai ./cmd/chenhai/

# 运行测试
go test ./internal/...

# 使用示例博客验证
cd testdata/example-blog
../../chenhai serve
```
