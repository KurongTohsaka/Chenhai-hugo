# Chenhai 规划文档

当前版本：**v0.2.0**

---

## 版本历史

### v0.1.0 — 核心骨架

- 项目架构（6 个 internal 包）
- Markdown 渲染管道（Goldmark + Chroma + KaTeX）
- YAML 配置系统
- 镇海主题基础框架
- 分类 / 标签 / 时间线 / 搜索索引
- CLI 命令（build / serve / new / clean / version）
- 开发服务器（fsnotify + LiveReload）

### v0.2.0 — 功能完善

- Admonition Goldmark 自定义扩展（`> [!note]` / `[!warning]` / `[!tip]` / `[!danger]`）
- 分页功能（首页 + 分类/标签页，`/page/N/` 路径）
- Favicon（镇海 SVG favicon）
- 标签云 5 级字号（xs→xl）
- 42 个测试，全部通过

---

## 当前问题

| 编号 | 问题 | 优先级 | 说明 |
|------|------|--------|------|
| B-01 | **标签云字号不生效** | 🟡 中 | `.Site.BuildTagCloud()` 通过 `interface{}` 无法在模板中调用，需改为 `.Extra.tagCloud` 传递 |
| B-02 | **分类/标签 URL 编码** | 🟢 低 | 中文字符被 percent-encode，属于标准 Web 行为 |
| Q-01 | CLI 包无测试 | 🟡 中 | 纯接线代码 |
| Q-02 | Server 包无测试 | 🟡 中 | 需 HTTP 测试框架 |
| Q-03 | 搜索/暗色模式 UI 未实际验证 | 🟢 低 | JS 交互逻辑待测 |
| Q-04 | reading time 估算偏大 | 🟢 低 | 中文按 `/3` 计算 |

---

## 版本规划

### v0.3.0 — 内容增强

| 任务 | 说明 |
|------|------|
| RSS/Atom Feed | 生成订阅文件 |
| fix B-01 | 标签云通过 `.Extra` 传递 |
| 图片处理 | 构建时缩放、WebP 转换 |
| 阅读时间优化 | 中英文混合统计算法 |
| 相关文章推荐 | 基于标签/分类 |
| SEO 增强 | Open Graph + JSON-LD 结构化数据 |

### v0.4.0 — 主题生态

| 任务 | 说明 |
|------|------|
| 外部主题加载 | 扫描 `themes/` 目录 |
| 主题参数 | `theme.yaml` 自定义参数 |
| 主题脚手架 | `chenhai new theme <name>` |

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
- 图片自动上传 CDN
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
