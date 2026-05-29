# Chenhai 规划文档

当前版本：**v0.5.4**（长期维护，持续打磨）

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

### v0.5.1 — 工作流与主题打磨

| 编号 | 内容 |
|------|------|
| feat | `chenhai deploy` 命令（add + commit + push，CI 负责构建） |
| ci | GitHub Actions CI 配置（push → test/build/deploy） |
| fix | Chroma CSS class 模式切换（亮/暗双模代码颜色） |
| feat | 代码字体升级 JetBrains Mono（Google Fonts CDN） |
| fix | 搜索/暗色 UI 验证修复 |
| fix | Mermaid 渲染修复（`<pre class="mermaid">` 格式） |
| fix | 跨版本模板回退与 tagCloud 接口修正 |
| docs | 工作流文档（Typora → Chenhai → GitHub Pages） |
| docs | 简化工作流（去掉本地 build，deploy 直达 CI） |

### v0.5.2 — 体验优化与工程健康

| 编号 | 内容 |
|------|------|
| feat | 首页 Hero 角色图替换 + 背景色占位 + 淡入动画 |
| feat | 代码块行号显示（启用已有 LineNumbers 配置） |
| feat | 代码块行高亮（`{hl_lines=[1,3-5]}` 语法） |
| feat | About 页面布局（自定义 layout 支持 + 图片画廊） |
| fix | 归档页月份倒序排列（12月→1月） |
| fix | Admonition 提示框 CSS 样式补全 |
| fix | `chenhai new` 模板默认 draft=false，增加 toc/description |
| test | CLI 包测试补全（version/clean/init/new/theme/root，10 个用例） |
| docs | 移除 Mermaid pan-zoom 已知问题，清理 C-01 |

### v0.5.3 — 暗色模式防闪烁与性能优化

| 编号 | 内容 |
|------|------|
| fix | 暗色模式页面切换白屏闪烁（head 内联脚本提前注入 dark class + meta theme-color 适配） |
| perf | JS 脚本全量添加 defer（main.js / search.js / Mermaid），消除渲染阻塞 |
| perf | Hero 图片添加 preload 预加载 |
| perf | 代码块行号 sticky 定位（水平滚动时行号固定） |
| fix | Hero 高度缩减 + 暗色模式标题色加亮 + text-shadow |
| fix | 代码块内边距收紧（行号左/代码右） |

### v0.5.4 — 交互体验与工程健康

| 编号 | 内容 |
|------|------|
| feat | 回到顶部按钮（滚动 >400px 显隐，平滑滚动） |
| feat | 404 页面 "沧海无迹"（含返回首页 / 搜索入口） |
| feat | 图片点击放大 Lightbox（Esc / 点击关闭） |
| fix | 暗色模式动态跟随系统（auto 模式实时切换） |
| feat | 分页器页码按钮（1 2 3 … 8） |
| test | Server 包测试补全（5 个用例） |
| feat | Front Matter 未知字段警告 |
| perf | 主题资源增量复制（跳过未变更的嵌入资源） |

---

## 当前问题

| 编号 | 问题 | 优先级 |
|------|------|--------|

---

## 版本规划

### v0.6.0 — 待定

| 任务 | 说明 |
|------|------|
| 相关文章推荐 | 基于标签/分类推荐 |
| 代码块语言标签 | 右上角显示语言名（Go / Python 等） |
| 搜索增强 | Fuse.js 集成（替换当前 indexOf 子串匹配） |

### v1.0.0 — 稳定发布

| 任务 | 说明 |
|------|------|
| 性能优化 | 模板预编译缓存、并发页面渲染 |
| 测试覆盖 | >80% 测试覆盖 |
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
