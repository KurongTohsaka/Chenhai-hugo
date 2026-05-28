# Chenhai 规划文档

当前版本：**v0.4.0** → 下一版本：**v0.5.0**

---

## 版本历史

### v0.1.0 — 核心骨架（2026-05）
项目架构、Markdown 管道、YAML 配置、镇海主题基础、分类/标签/时间线/搜索、CLI、开发服务器

### v0.2.0 — 功能完善（2026-05）
Admonition 扩展、分页、Favicon、标签云字号

### v0.3.0 — Markdown 渲染完整性（2026-05）
KaTeX & Mermaid CDN 渲染、图片增强（figure + figcaption）、GitHub 图床、镇海角色装饰图、builder.go 拆分

### v0.4.0 — 主题生态（2026-05）
外部主题加载 & 回退、主题参数（theme.yaml → config.yaml 覆盖）、主题脚手架（`chenhai new theme`）

---

## v0.5.0 范围

> 体验打磨与工具链完善。后续较长一段时间将停留在 v0.5.x，持续修 bug、打磨细节。

### 新功能

| 编号 | 功能 | 说明 |
|------|------|------|
| F1 | **增量构建** | 只重建有变更的文件（内容、模板、配置），大幅提速 |
| F2 | **`chenhai init`** | 一键生成站点骨架（config.yaml + content/ + static/ + archetypes/） |
| F3 | **`chenhai new` 增强** | 支持 `--category`、`--tags` 参数自动填充 front matter |
| F4 | **构建进度与日志** | 实时输出「扫描 → 渲染 N/M → 生成分类 → 完成」 |

### Bug 修复 & UI 打磨

| 编号 | 说明 |
|------|------|
| B1 | 搜索 UI 验证 — 前端搜索交互测试与修复 |
| B2 | 暗色模式验证 — JS 交互、CSS 变量切换测试 |
| B3 | 错误提示优化 — 构建 / 配置错误时给出可操作的提示 |
| B4 | `.Site.BuildTagCloud` 接口类型问题 — `interface{}` → `*index.Site` |
| B5 | `Render` 签名 — 内联 `io.Writer` → import `"io"` |

---

## v1.0.0 — 稳定发布

| 任务 | 说明 |
|------|------|
| 性能优化 | 模板预编译缓存、并发页面渲染 |
| API 参考文档 | 完整 config.yaml 字段 + 模板变量参考 |
| 场景教程 | Hugo 迁移指南、GitHub Pages / Vercel 部署 |
| 测试覆盖 | >80%，补齐 CLI / Server 包测试 |

---

## 远期愿景

- 插件系统（Lua / Starlark）
- 在线编辑器（Web IDE）
- 多站点工作区

---

## 贡献指南

```bash
git clone https://github.com/KurongTohsaka/chenhai-hugo.git
cd chenhai-hugo
go build -o chenhai ./cmd/chenhai/
go test ./internal/...                            # 运行测试
cd testdata/example-blog && ../../chenhai serve   # 示例博客
```
