# Chenhai 规划文档

当前版本：**v0.2.0** → 下一版本：**v0.3.0**

---

## 版本历史

### v0.1.0 — 核心骨架

- 项目架构（6 个 internal 包）
- Markdown 渲染管道（Goldmark + Chroma + KaTeX 分隔符）
- YAML 配置系统
- 镇海主题基础框架
- 分类 / 标签 / 时间线 / 搜索索引
- CLI 命令（build / serve / new / clean / version）
- 开发服务器（fsnotify + LiveReload）

### v0.2.0 — 功能完善

- Admonition Goldmark 扩展（`> [!note]` / `[!warning]` / `[!tip]` / `[!danger]`）
- 分页功能（首页 + 分类/标签页，`/page/N/` 路径）
- Favicon（镇海 SVG favicon）
- 标签云 5 级字号（xs→xl）
- 42 个测试，全部通过

---

## 当前问题

| 编号 | 问题 | 优先级 | 状态 |
|------|------|--------|------|
| B-01 | **侧边栏标签云** — `.Site.BuildTagCloud` 通过 `interface{}` 调用受限 | 🟡 中 | 待修复 |
| B-02 | **builder.go 过大** — 400+ 行，需按职责拆分 | 🟡 中 | 待重构 |
| B-03 | 分类/标签 URL 中文编码 | 🟢 低 | 标准 Web 行为 |
| C-01 | CLI 包无测试 | 🟢 低 | 接线代码 |
| C-02 | Server 包无测试 | 🟢 低 | 需 HTTP 测试框架 |

---

## v0.3.0 范围

### 新功能

| 编号 | 功能 | 说明 |
|------|------|------|
| F1 | **数学公式渲染** | 内置 KaTeX，自动检测 `$` 分隔符后注入 JS/CSS |
| F2 | **Mermaid 图表渲染** | 内置 Mermaid.js，自动检测 mermaid 代码块后注入 JS |
| F3 | **图片增强** | `<img>` → `<figure>` + `<figcaption>`，对齐参数 |
| F4 | **镇海角色图片** | 装饰图加入主题 assets/images/ |
| F5 | **GitHub 图床** | 自动上传 + URL 映射双模式 |

### Bug 修复 / 重构

| 编号 | 说明 |
|------|------|
| B-01 | 侧边栏标签云数据通过 `.Extra` 传递 |
| B-02 | builder.go 按职责拆分为多文件 |
| Del | 移除阅读时间功能（CalcReadingTime、模板中的显示、测试） |

---

## 后续版本

### v0.4.0 — 主题生态

| 任务 | 说明 |
|------|------|
| 外部主题加载 | 扫描 `themes/` 目录 |
| 主题参数 | `theme.yaml` 自定义参数 |
| 主题脚手架 | `chenhai new theme <name>` |
| 分类/标签 Slug 化 | 中文拼音路径 |

### v1.0.0 — 稳定发布

| 任务 | 说明 |
|------|------|
| 增量构建 | 只重建变更文件 |
| i18n | 中/英多语言 |
| 性能优化 | 模板缓存、并发渲染 |
| 完整测试 | >80% 覆盖率 |
| 官方文档站 | chenhai.dev |

---

## 远期愿景

- 插件系统（Lua / Starlark 脚本）
- 在线编辑器（Web IDE）
- 多站点工作区

---

## 贡献指南

```bash
git clone https://github.com/KurongTohsaka/chenhai-hugo.git
cd chenhai-hugo
go build -o chenhai ./cmd/chenhai/

# 运行测试
go test ./internal/...

# 示例博客
cd testdata/example-blog && ../../chenhai serve
```
