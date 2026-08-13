# 验证报告 · Chenhai-hugo v0.8.0 线 B shortcode

| 项 | 值 |
|---|---|
| 验证对象 | `feat/v0.8.0-shortcode`（worktree: `/Users/kurong/Project/GoProject/Chenhai-hugo-worktrees/shortcode`） |
| HEAD | `8c2cf10`（4 提交：13c2d0e Task4 → 19d762d Task5 → 35f76fb strip 修复 → 8c2cf10 Task6） |
| 验证者 | tester（独立第三方，证据导向） |
| 日期 | 2026-08-13 |
| 验收依据 | `docs/superpowers/plans/2026-08-13-v0.8.0-writing-experience.md` Task 4-6；`CONTRIBUTING.md` |
| 测试命令 | `go test ./internal/... -count=1`；`go build -o chenhai ./cmd/chenhai/`；worktree 内站点构建 + /tmp 副本对抗实测 |

## 结论摘要

| 级别 | 数量 | 摘要 |
|---|---|---|
| 🔴 阻塞 | 1 | 单行 inline shortcode（`{{< name >}}…{{< /name >}}` 同行）**吞并其后全部文档内容**，静默无提示（含未闭合吞并变体） |
| 🟡 建议 | 2 | demo 页 `draft: false` 违背计划；两处单测断言弱/输入复现缺陷却假 PASS |
| ⚪ 可选 | 2 | 任务消息 .gitignore 声明与分支事实不符（无实际影响）；多行形态 Content 首尾多余空行 |

**门禁判定：🔴 存在，不可合入。** 核心渲染链（Y1 占位符管线、fenced 优先、未知透传、details 内层渲染、主题覆盖接口、CSS/JS、?v=18）全部通过，但短码块终止语义缺陷会在受支持语法下静默吞内容，属数据/内容完整性红线。

## 逐项核对（Task 4-6 规格 → 实现证据）

| # | 规格条款 | 实现 | 证据 | 判定 |
|---|---|---|---|---|
| 1 | Task4 文件：shortcode.go / shortcode_registry.go / renderer.go / shortcode_test.go | 全部新增/修改 | `git diff main...HEAD --name-status`：A shortcode.go, A shortcode_registry.go, A shortcode_test.go, M renderer.go | ✅ |
| 2 | Task5：shortcodes.go + RegisterBuiltins(details/gallery/tabs) | 存在 | `shortcodes.go` RegisterBuiltins 注册三组件；renderer.go:51 `RegisterBuiltins(reg)` | ✅ |
| 3 | Task6：theme.Engine.LookupShortcodeTemplate（三层查找复用 readLayoutSource） | 存在 | `internal/theme/engine.go:250-252`；readLayoutSource 182-241 行 = 缓存→site→extTheme→embedded 四层（含缓存+锁） | ✅ |
| 4 | Task6：build.New 注入 provider（一处覆盖 build/serve） | 存在 | `internal/build/builder.go:42` `r.Registry().SetTemplateProvider(e)` | ✅ |
| 5 | Task6：Registry getter | 存在 | `renderer.go:91` `Registry() *ShortcodeRegistry` | ✅ |
| 6 | Y1 占位符管线：stripShortcodeBlocks → extract → Convert → inject → ReplacePlaceholders | 存在 | `renderer.go:74-87`（RenderHTML）、101-114（RenderHTMLWithTOC）；shortcode.go:167-196 stripShortcodeBlocks | ✅ |
| 7 | Y1 行为：tabs 内代码块不抢正文 lang/hl | **端到端验证** | /tmp 副本构建：正文 python 块 `<pre class="chroma" tabindex="0" data-hl-lines="1">`；tabs 内 go 块 `<pre tabindex="0" class="chroma">` 无 hl 属性；`data-hl-lines="1"` 全局仅 1 处。单测 `TestRenderShortcode_TabsAndHLLines` PASS（断言 data-hl-lines，与端到端一致） | ✅ |
| 8 | fenced code 内 `{{<` 不解析 | 端到端验证 | `/tmp/sc_variant2` 构建：fenced 内 `{{&lt; details` 转义输出，`shortcode-details` 计数 0；单测 `InsideCodeBlock` PASS | ✅ |
| 9 | 未知 shortcode 透传 HTML 注释 | 端到端验证 | `<!-- unknown shortcode: nosuch -->` 出现；单测 PASS | ✅ |
| 10 | details 内层 markdown 渲染 | 端到端验证 | 仓库 demo 页：`<div class="details-body"><p>这里是被折叠的<strong>重点内容</strong>…</p>`；单测 `Details` PASS | ✅ |
| 11 | tabs 结构（data-tab/data-panel/active/language-go） | 端到端验证 | demo 页：`data-tab="0" role="tab">Go<`、`data-tab="1" role="tab">Python<`、2 tab-panel、language-go/language-python 各 1 | ✅ |
| 12 | gallery 结构 | 端到端验证 | demo 页：`<img src="/img/demo-a.png" alt="示例一" loading="lazy">`、demo-b 同 | ✅ |
| 13 | 主题覆盖接口（layouts/shortcodes/ 模板命中） | 端到端验证 | `/tmp/sc_variant2` site 层放 `layouts/shortcodes/details.html` → 输出 `<div class="override-details">OVERRIDE:[没闭合]:&lt;p&gt;…&lt;/p&gt;…`（Positional 索引 + ContentHTML 均正确）；无模板时（demo 页）回退内置 `<details class="shortcode-details">`——优先级正确 | ✅ |
| 14 | innerMD 防递归 | 代码 + 端到端 | renderer.go:45-55 innerMD 用 exts（无 shortcodeExt）；覆盖模板输出中嵌套 `{{< details … >}}` 被转义为 `{{&amp;lt; details…}}` 文本（未二次解析） | ✅ |
| 15 | CSS 用 `--color-*` 体系 | 代码核对 | style.css:2889-2906 组件样式全用 `var(--color-border/bg-card/text/text-secondary/link, fallback)`；`:root` 21 行与暗色模式 111/168 行均有定义（Y6 满足） | ✅ |
| 16 | tabs 交互 JS | 代码核对 | main.js:669-673（closest .tab-btn → [data-tabs] → toggle active） | ✅ |
| 17 | base.html `?v=18` | 代码核对 | base.html:22 `style.css?v=18` | ✅ |
| 18 | 产物清理 / .gitignore | 验证 | 构建产物 public/、.chenhai-cache.json、chenhai 二进制均被忽略（.gitignore:38/46/48），`git status --porcelain` 空 | ✅ |
| 19 | shortcode-demo.md（Task6 Step6 要求 draft: true） | **偏差** | `testdata/example-blog/content/posts/shortcode-demo.md:6` 为 `draft: false` | 🟡 |
| 20 | 单行 inline 形态（Open 注释宣称支持） | **🔴** | 见下问题 1 | 🔴 |

## 测试输出

`go test ./internal/... -count=1`：8 包全 ok（build/cli/config/content/imagehost/index/server/theme）。

`TestRenderShortcode_*` 明细（8/8 PASS）：

```
=== RUN   TestRenderShortcode_Unknown      --- PASS
=== RUN   TestRenderShortcode_Unclosed     --- PASS
=== RUN   TestParseShortcodeParams         --- PASS
=== RUN   TestRenderShortcode_InsideCodeBlock --- PASS
=== RUN   TestRenderShortcode_Gallery      --- PASS
=== RUN   TestRenderShortcode_Tabs         --- PASS
=== RUN   TestRenderShortcode_Details      --- PASS
=== RUN   TestRenderShortcode_TabsAndHLLines --- PASS
```

功能实测：`go build -o chenhai ./cmd/chenhai/` → `testdata/example-blog` 全量构建（rm -rf public .chenhai-cache.json，EXIT=0）→ grep：`class="tabs"`=1、`class="gallery"`=1、`shortcode-details`=1、language-go=1、language-python=1、`<summary>答案</summary>` 命中。

对抗实测（/tmp/sc_variant、/tmp/sc_variant2 副本，独立验证程序 `tmp_verify/` 项目 module 内运行后清理）：

| 实测项 | 结果 |
|---|---|
| tabs+正文 hl_lines 共存（Y1） | PASS（data-hl-lines="1" 正确归属正文块） |
| fenced code 内 {{< 不解析 | PASS |
| 未知 shortcode 透传注释 | PASS |
| 未闭合（EOF）透传 | PASS（单测覆盖） |
| 主题覆盖模板命中 | PASS |
| **单行 inline 吞并** | **FAIL（🔴）** |
| **未闭合吞到下一同名 close** | **FAIL（🟡 变体）** |

## 问题清单

### 🔴 1. 单行 inline shortcode 吞并后续全部文档内容（内容完整性红线）

- **位置**：`internal/content/shortcode.go` — Open 62-76 行（inline form 支持，同行闭合设 `node.Closed=true` 后返回 `parser.NoChildren`）；Continue 82-94 行（无 Closed 检查）。
- **机制**：goldmark v1.8.2 中块一旦 Open 入栈，后续每行先调该块 `Continue`（`parser/parser.go:1046`）；inline 形态的闭合标签已在 Open 同行被 `shortcodeCloseInlineRe` 消费，之后**没有可匹配的 close 行**——`Continue` 持续 append 到 EOF，`Content` 吞入后续全部文档。
- **端到端实证**（/tmp/sc_variant 构建）：`{{< nosuch >}}内容{{< /nosuch >}}` 后 `## details 内层 markdown`、details 组件等全部内容被吞入透传输出，details 组件渲染丢失。
- **AST 实证**（goldmark 解析树）：`{{< nosuch >}}x{{< /nosuch >}}\n\n## 后续标题\n\n正文内容\n` → `Content="x\n\n\n## 后续标题\n\n\n\n正文内容\n\n"`，Heading/Text 节点全被吞入短码块。
- **单测盲区（假 PASS）**：`TestRenderShortcode_Details`（shortcode_test.go:97）输入**本身即单行形态且带后续内容**（`…{{< details "答案" >}}折叠的**内容**{{< /details >}}\n\n结尾\n`）——"结尾"被吞入 `<details>` 内部，断言仅检查 wrapper/summary/strong 存在，未检查后续内容位置 → 测试 PASS 而缺陷在场。`TestRenderShortcode_Unknown` 断言仅 `contains "nosuch"`（原文透传也含），无法区分识别/未识别。
- **影响**：作者按 Open 注释宣称的 inline 语法（`{{< name params >}}rest…{{< /name >}}`）书写即触发，**静默**吞并（Closed=true 走正常渲染路径，无 `unclosed` 注释提示），大段内容渲染错乱或丢失。
- **修复建议（已在副本验证有效）**：`Continue` 开头加 `if sc.Closed { return parser.Close }` —— 模拟修复后 AST：`Content="x\n"` 且后续 Heading/Text 正常解析、块终止。需 developer 修复 + 补回归测试（单行形态 + 后续内容，断言后续内容在块外）。

### 🟡 2. 未闭合 shortcode 吞到下一个同名 close（无提示吞并变体）

- **实证**：/tmp/sc_variant2 页第一个 `{{< details "没闭合" >}}`（漏写闭合）把 `## 主题覆盖测试` 和第二个 details 全部吞入 Content，由第二个 `{{< /details >}}` 关闭并渲染（输出显示 `OVERRIDE:[没闭合]:…&lt;h2 id=…&gt;主题覆盖测试&lt;/h2&gt;…` 内容错位）。
- EOF 形态有 `<!-- unclosed shortcode -->` 提示（单测覆盖）；**中间形态（遇下一同名 close）无任何提示**，Content 含后续内容，已知组件渲染错乱。
- 同根因（块终止语义），修复成本高于 #1（需 Content 内嵌套 open 检测或闭合匹配策略），与 #1 一并处理。

### 🟡 3. demo 页 draft 与计划不符

- `testdata/example-blog/content/posts/shortcode-demo.md:6` 为 `draft: false`，计划 Task 6 Step 6 明确要求 `draft: true`（避免污染线上示例）。文档级一行改，orchestrator 可直办。

### 🟡 4. 单测断言强度不足（与 🔴 直接相关）

- `TestRenderShortcode_Unknown`（:16）仅断言 `contains "nosuch"`，无法区分"识别+透传"与"未识别原文输出"；建议断言 `<!-- unknown shortcode: nosuch -->` 注释标记。
- `TestRenderShortcode_Details` 输入复现 🔴 缺陷（单行+后续），建议补"后续内容在块外"断言（如 `strings.Index(html, "结尾") > strings.Index(html, "</details>")`）。

### ⚪ 5. 任务消息 .gitignore 声明与分支事实不符（无实际影响）

- 任务消息称".gitignore 已加 testdata/example-blog/static/"——本分支 .gitignore 实际仅含 `testdata/example-blog/public/`（:38），无 static/；`git show main:.gitignore` 显示 main 已有 `static/`（:39，线 A 合并时加）。本线构建不产生 static 产物（目录不存在），实际无风险；合并后自动获得该行。

### ⚪ 6. 多行形态 Content 首尾多余空行（观察点）

- AST 实证：`{{< nosuch >}}\nx\n{{< /nosuch >}}\n` → `Content="\n\nx\n\n"`（首尾各多空行，源自 goldmark 行推进 + AdvanceToEOL 停在 `\n` 的段收集细节）。已知组件不受影响（tabs 段值 TrimSpace、innerMD 渲染忽略首尾空行、gallery 正则提取）；仅未知透传输出多空行（视觉差异）。无功能影响，可与 #1 修复一并 trim。

## Final Assessment

- **门禁状态**：🔴 驳回 → 回 developer 修复（问题 1 + 补回归测试；建议连带 2/4）→ 复审重跑对抗探针（单行 inline 形态 + 后续内容端到端、未闭合中间形态）。
- 修复验证路径：`Continue` 开头 `if sc.Closed { return parser.Close }`（副本模拟已验证），回归测试建议形态：`{{< details "标题" >}}折叠**内容**{{< /details >}}\n\n结尾\n` 断言"结尾"在 `</details>` 之后。
- 核心渲染链其余项（Y1 占位符管线、fenced 优先、未知透传、主题覆盖三层查找、innerMD 防递归、CSS/JS、?v=18、build.New 注入）全部通过，未发现其他问题。
- 报告随分支入库；本验证未修改任何被测代码（只读 + /tmp 副本 + tmp_verify 即建即删，`git status --porcelain` 干净收尾）。
