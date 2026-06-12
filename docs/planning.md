# Chenhai 规划文档

当前版本：**v0.6.5**（功能迭代）

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

### v0.5.5 — 体验细节

| 编号 | 内容 |
|------|------|
| feat | 标题锚点链接（hover 显示 #，点击复制 URL） |
| fix | 复制代码块时排除行号（JS 过滤 .code-ln） |

### v0.6.0 — 功能迭代

| 编号 | 内容 |
|------|------|
| feat | 相关文章推荐（标签 Jaccard 相似度） |
| feat | 代码块语言标签（浮动卡片样式，hover 显隐） |
| feat | Fuse.js 模糊搜索（替换 indexOf 子串匹配，CDN 引入） |
| feat | Series 系列文章（播放列表盒子，weight/date/title 三级排序，自由跳转） |
| feat | 构建耗时统计（各阶段耗时输出 + 总耗时） |
| feat | Draft 预览模式（`chenhai serve --drafts` 本地预览草稿） |
| fix | 双 Copy 按钮去重、语言标签边框样式、Series 滚动限制与 visited 色 |
| revert | 面包屑导航（已移除，不实用） |

### v0.6.1 — 细节修复

| 编号 | 内容 |
|------|------|
| fix | 暗色模式全局 visited 链接色加 0.85 透明度 |
| fix | 无行号代码块也支持语言标签 |
| fix | Series 列表长标题截断省略 |
| perf | Google Fonts 从 CSS @import 改为 HTML link（非阻塞加载） |
| fix | 分页器移动端换行（flex-wrap） |

### v0.6.2 — AI 标记、Series、关于页完善

| 编号 | 内容 |
|------|------|
| feat | AI 生成文章标记（front matter `ai_generated` + 紫色徽章） |
| feat | Series 系列分类渲染（`renderSeries` 生成 `/series/` 聚合页） |
| feat | 关于页完善：12 张镇海素材图按分类展示（立绘/换装/改造/皮肤/誓约） |
| feat | 首页 Hero 图替换为镇海誓约日服预告 |
| fix | 关于页 404 修复（Section 解析去 `.md` 扩展名） |
| fix | Series 链接 404 修复（新增 `/series/` 分类页生成） |

### v0.6.3 — 搜索、移动端、错误提示、暗色过渡

| 编号 | 内容 |
|------|------|
| feat | 搜索评分优化（标题×10、标签×5、时效加分、限 20 条） |
| feat | 移动端适配（768/480 断点，隐藏 TOC、分页器换行、代码块缩小） |
| feat | 构建错误中文提示（Front Matter / config.yaml 解析失败时给修复建议） |
| feat | 暗色模式图片亮度调节 + 代码块背景过渡 |

### v0.6.4 — 搜索体验打磨

| 编号 | 内容 |
|------|------|
| feat | 搜索历史（localStorage，最近 32 条，× 删除按钮） |
| fix | 搜索历史仅按 Enter 确认时保存，输入过程不记录 |
| fix | 中文输入法下 Enter 不再误跳转（IME composition 检测） |
| fix | 标题锚点 # 链接修复（改用 location.hash 导航） |
| fix | 搜索框中文提示、无历史时空白、叉号/快捷键低调化 |

### v0.6.5 — 搜索 UI 修复

| 编号 | 内容 |
|------|------|
| fix | 搜索框输入区全宽可点击（icon 绝对定位 + flex: 1 1 0%） |
| fix | 历史项 SVG 删除图标（currentColor 自适应亮暗，hover 显现） |
| fix | CSS 语法错误修复（多余 `}` 导致后续规则全部失效） |
| fix | 搜索框右侧按钮不遮挡输入区域 |

---

## 当前问题

| 编号 | 问题 | 优先级 |
|------|------|--------|

---

## 版本规划

### v0.6.6 — 待定

> 继续打磨体验。

| 任务 | 说明 |
|------|------|
| 图片懒加载优化 | 首屏图片 eager、折叠内容 lazy |
| 其他已知问题 | 日常使用中收集修复 |

### v0.7.0 — 内容创作增强

| 任务 | 说明 |
|------|------|
| Shortcodes 短代码 | `{{< youtube id >}}` `{{< bilibili aid >}}` 等可复用模板片段 |
| 自定义分类法 | config 中定义额外 taxonomy（如 `series`、`tools`） |
| 图片响应式 | 自动生成 srcset / WebP，图片尺寸属性 |
| 图片处理命令 | `chenhai image` 子命令：压缩、转换、尺寸调整 |
| 相关文章优化 | 分类加权 + 时效衰减 + 兜底策略（标签不够时降级为同分类推荐） |

### v0.8.0 — 构建引擎升级

| 任务 | 说明 |
|------|------|
| 并发页面渲染 | goroutine pool 并行渲染，大幅缩短全量构建时间 |
| 模板预编译缓存 | 首次编译后缓存解析结果，后续构建跳过模板解析 |
| 构建报告面板 | 各阶段耗时表格、文件统计、缓存命中率 |
| 开发服务器增强 | 按需构建（只渲染当前访问页面）+ 快速增量 |

### v0.9.0 — 生态与扩展

| 任务 | 说明 |
|------|------|
| 数据文件 | `data/` 目录 JSON/YAML，模板中 `{{.Data.xxx}}` 访问 |
| i18n 内容支撑 | 多语言文章（`hello.en.md` / `hello.zh.md`） |
| 钩子系统雏形 | 内容/渲染生命周期钩子（Go 接口方式，不用 Lua） |
| 健康检查命令 | `chenhai doctor` 检测配置、目录结构、依赖完整性 |
| 站点配置文件拆分 | 支持 `config/` 目录，按环境拆分（`dev.yaml` / `prod.yaml`） |

### v1.0.0 — 稳定发布

| 任务 | 说明 |
|------|------|
| 性能基线 | 全量构建 100 篇文章 <3s |
| 测试覆盖 | >80% |
| 官方文档站 | chenhai.dev |
| 向后兼容保证 | 配置文件、模板 API、目录结构冻结 |

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
