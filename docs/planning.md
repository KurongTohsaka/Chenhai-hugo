# Chenhai 规划文档

当前版本：**v0.5.0**（长期维护，持续打磨）

测试站点：**https://hekurong.github.io**

---

## 版本历史

| 版本 | 时间 | 主要交付 |
|------|------|---------|
| v0.1.0 | 2026-05 | 项目架构、Markdown 管道、YAML 配置、镇海主题、分类/标签/时间线/搜索、CLI、开发服务器 |
| v0.2.0 | 2026-05 | Admonition 扩展、分页、Favicon、标签云字号 |
| v0.3.0 | 2026-05 | KaTeX & Mermaid CDN 渲染、图片增强（figure + figcaption）、GitHub 图床、镇海装饰图、builder.go 拆分 |
| v0.4.0 | 2026-05 | 外部主题加载 & 回退、主题参数、主题脚手架 |
| v0.5.0 | 2026-05 | 增量构建、`chenhai init`、`chenhai new` 增强（--category/--tags）、构建进度日志、Site 类型修正、io.Writer 规范化 |

### v0.5.x 持续打磨

| 编号 | 内容 |
|------|------|
| fix | Chroma CSS class 模式切换（亮/暗双模代码颜色） |
| fix | 代码字体升级 JetBrains Mono（Google Fonts CDN） |
| fix | 搜索/暗色 UI 验证修复 |
| fix | Mermaid 渲染修复（`<pre class="mermaid">` 格式） |
| fix | 跨版本模板回退与 tagCloud 接口修正 |
| docs | 工作流文档（Typora → Chenhai → GitHub Pages） |

---

## 当前问题

| 编号 | 问题 | 优先级 |
|------|------|--------|
| C-01 | CLI 包无测试 | 🟢 低 |
| C-02 | Server 包无测试 | 🟢 低 |
| B-01 | Mermaid 无 pan-zoom 交互 | 🟢 低 |

---

## 版本规划

### v0.6.0 — 待定

| 任务 | 说明 |
|------|------|
| 相关文章推荐 | 基于标签/分类推荐 |
| RSS/Atom Feed | 生成订阅文件 |
| 搜索增强 | Fuse.js 集成 |

### v1.0.0 — 稳定发布

| 任务 | 说明 |
|------|------|
| 性能优化 | 模板预编译缓存、并发页面渲染 |
| 测试覆盖 | >80%，补齐 CLI / Server 包测试 |
| 官方文档站 | chenhai.dev |

---

## 远期愿景

- 插件系统（Lua / Starlark）
- 在线编辑器（Web IDE）
- 多站点工作区

---

## 文档索引

| 文档 | 说明 |
|------|------|
| [tutorial.md](tutorial.md) | 使用教程（安装 → 配置 → 写作 → 部署） |
| [workflow.md](workflow.md) | 工作流文档（Typora 写作 → GitHub Pages） |
| [superpowers/specs/](superpowers/specs/) | 各版本设计规格文档 |
| [superpowers/plans/](superpowers/plans/) | 各版本实现计划 |

---

## 贡献指南

```bash
git clone https://github.com/KurongTohsaka/chenhai-hugo.git
cd chenhai-hugo
go build -o chenhai ./cmd/chenhai/
go test ./internal/...
cd testdata/example-blog && ../../chenhai serve
```
