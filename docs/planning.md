# Chenhai 规划文档

当前版本：**v0.4.0**

---

## 版本历史

### v0.1.0 — 核心骨架（2026-05）

- 项目架构（6 个 internal 包）
- Markdown 渲染管道（Goldmark + Chroma + KaTeX 分隔符）
- YAML 配置系统
- 镇海主题基础框架
- 分类 / 标签 / 时间线 / 搜索索引
- CLI 命令（build / serve / new / clean / version）
- 开发服务器（fsnotify + LiveReload）

### v0.2.0 — 功能完善（2026-05）

- Admonition Goldmark 扩展（`> [!note]` / `[!warning]` / `[!tip]` / `[!danger]`）
- 分页功能（首页 + 分类/标签页，`/page/N/` 路径）
- Favicon（镇海 SVG favicon）
- 标签云 5 级字号（xs→xl）
- 移除阅读时间功能

### v0.3.0 — Markdown 渲染完整性（2026-05）

- **KaTeX 渲染** — CDN 引入 KaTeX 0.16.11，自动检测数学分隔符并注入 JS/CSS
- **Mermaid 渲染** — CDN 引入 Mermaid.js 11，自动检测 `<pre class="mermaid">` 并渲染 SVG 图表
- **图片增强** — `<img>` → `<figure>` + `<figcaption>` + 居中/右对齐 + lazy loading
- **镇海角色图片** — 首页 hero 区背景图 + 关于页展示图（SVG 占位，可替换真图）
- **GitHub 图床** — auto 上传（GitHub API）+ map 映射双模式
- **builder.go 拆分** — 单文件 400+ 行拆为 8 个职责文件

### v0.4.0 — 主题生态（2026-05）

- **外部主题加载** — 扫描 `themes/` 目录，`config.yaml` 切换 theme
- **模板回退** — 缺模板自动回退镇海，只需写想定制的那几页
- **主题参数** — `theme.yaml` 定义 params 默认值，`config.yaml` 的 `themeConfig` 可覆盖，模板中访问
- **主题脚手架** — `chenhai new theme <name>` 生成完整主题骨架
- **中文 URL** — 确认为标准 Web 行为，非 bug

---

## 当前问题

| 编号 | 问题 | 优先级 |
|------|------|--------|
| C-01 | CLI 包无测试 | 🟢 低 |
| C-02 | Server 包无测试 | 🟢 低 |

---

## 版本规划

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
