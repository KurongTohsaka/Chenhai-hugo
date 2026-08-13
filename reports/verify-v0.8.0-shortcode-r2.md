# 验证报告 · Chenhai-hugo v0.8.0 线 B shortcode（复审 R2）

| 项 | 值 |
|---|---|
| 验证对象 | `feat/v0.8.0-shortcode`（worktree: `/Users/kurong/Project/GoProject/Chenhai-hugo-worktrees/shortcode`） |
| HEAD | `1de9c74`（修复提交 `f8c8a2e` fix: shortcode inline close and unclosed-termination semantics + `1de9c74` docs: demo page draft flag） |
| 复审者 | tester（独立第三方，证据导向） |
| 日期 | 2026-08-13 |
| 复审依据 | 上一轮 `reports/verify-v0.8.0-shortcode.md`（🔴1 🟡2 🟡3 🟡4 ⚪5 ⚪6）逐条复验 |
| 测试命令 | `go test ./internal/... -count=1`；`go build -o chenhai ./cmd/chenhai/`；module 内独立探针 `tmp_verify/`（即建即删）；/tmp/sc_r2 副本站点端到端构建 |

## 结论摘要

| 级别 | 数量 | 摘要 |
|---|---|---|
| 🔴 阻塞 | 0 | 原 🔴（单行 inline 吞并）已修复，对抗探针 + 端到端全过 |
| 🟡 建议 | 1 | **新发现（f8c8a2e 引入）**：多行组件 Content 内含 `{{< 名 >}}` 形态行时被「新 open 终止」逻辑截断——details 内展示 shortcode 语法示例的用例回归（可见有提示，非静默） |
| ⚪ 可选 | 2 | 任务消息「8 个 TestRenderShortcode_*」口径（实测 9 个 + TestParseShortcodeParams）；Close 的 TrimSpace 对未知透传首行缩进边缘影响 |

**门禁判定：放行。** 原 🔴（内容完整性红线）与 🟡2（无提示吞并变体）均已修复并有回归测试；demo 页 draft 已按计划修正。新 🟡 为边缘用例行为回归，触发条件明确（Content 内出现 open 形态行）、影响可见有注释提示、符合 v0.8「无嵌套」设计语义，不阻塞合并，建议留作后续改进或写入文档限制说明。

## 原问题逐条复验

| 原问题 | 修复证据（f8c8a2e / 1de9c74） | 复审结果 |
|---|---|---|
| 🔴 1 单行 inline 吞并后续内容 | shortcode.go:87-89 `if sc.Closed { return parser.Close }`（Continue 开头）；strip 层 :210 `if !shortcodeCloseInlineRe.MatchString(trimmed)` 单行 inline 不进入 inSC | ✅ 探针 A/B 全过：未知单行 inline 后 `## 后续标题`+正文在块外正常渲染、透传块仅含 `x`；details 单行后「结尾正文」在 `</details>` 之后（`close@N < 结尾@M`） |
| 🟡 2 未闭合吞到下一同名 close | shortcode.go:94-96 `if shortcodeOpenRe.Match(line) { return parser.Close }`（不消费行，新 open 自行开块） | ✅ 探针 C 全过：`<!-- unclosed shortcode: details -->` 标记出现、第二个 details 正常渲染、内容甲不泄漏进乙块、内容甲透传可见 |
| 🟡 3 demo 页 draft | 1de9c74：`draft: false → true`（已确认 diff 单行） | ✅ shortcode-demo.md:6 现为 `draft: true` |
| 🟡 4 单测断言弱 | Unknown 改断言 `<!-- unknown shortcode: nosuch -->` 注释标记（shortcode_test.go:17）；Details 增「结尾在 </details> 后」位置断言（:113-119）；新增 `TestRenderShortcode_InlineFollowedByContent`、`TestRenderShortcode_UnclosedBeforeNewOpen` | ✅ 两回归测试形态与原缺陷同构且断言位置，非弱断言 |
| ⚪ 5 .gitignore 声明 | 未动（无影响项） | ✅ 不适用，维持原判 |
| ⚪ 6 Content 首尾空行 | shortcode.go:120 `sc.Content = strings.TrimSpace(buf.String())` | ✅ 主探针无回归；首行缩进边缘影响见新 ⚪ 2 |

## 测试输出

`go test ./internal/... -count=1`：8 包全 ok（build/cli/config/content/imagehost/index/server/theme）。

`TestRenderShortcode_*` 明细（9/9 PASS，含 2 个新增回归）：

```
=== RUN   TestRenderShortcode_Unknown            --- PASS
=== RUN   TestRenderShortcode_Unclosed           --- PASS
=== RUN   TestRenderShortcode_InsideCodeBlock    --- PASS
=== RUN   TestRenderShortcode_Gallery            --- PASS
=== RUN   TestRenderShortcode_Tabs               --- PASS
=== RUN   TestRenderShortcode_Details            --- PASS
=== RUN   TestRenderShortcode_InlineFollowedByContent   --- PASS（新增，🔴 回归）
=== RUN   TestRenderShortcode_UnclosedBeforeNewOpen     --- PASS（新增，🟡2 回归）
=== RUN   TestRenderShortcode_TabsAndHLLines     --- PASS
```

（另 `TestParseShortcodeParams` PASS。注：任务消息「8 个 TestRenderShortcode_*」沿用了原报告含 ParseShortcodeParams 的口径，纯 TestRenderShortcode 前缀实测 9 个。）

`go build -o chenhai ./cmd/chenhai/`：EXIT=0（20,165,394 字节）。

## 对抗实测（独立探针 + 端到端）

module 内独立程序（`tmp_verify/` 即建即删）+ /tmp/sc_r2 副本站点（probe 页含 A/B/C 三形态 + 顶层 python 代码块）端到端构建（EXIT=0）后断言产物 HTML：

| 实测项 | 独立探针 | 端到端产物 |
|---|---|---|
| A 单行 inline 未知 + 后续标题/正文，内容在块外 | 5/5 PASS（h2 在注释标记之后、透传块不含后续内容） | 5/5 PASS |
| B 单行 details + 结尾正文，结尾在 `</details>` 之后 | 3/3 PASS（`iJ > iD`） | 3/3 PASS |
| C 未闭合中间形态，终止于新 open | 6/6 PASS（unclosed 标记 + 乙块正常 + 无泄漏 + 甲内容可见） | 4/4 PASS |
| E strip 层：单行 inline 后顶层 python 代码块完整（lang/hl 提取不受损） | 2/2 PASS（chroma `data-lang="python"` + 代码体完整；首版「return 1」断言被 chroma span 分割属脚本构造错，修正后过） | 2/2 PASS |
| D 副作用观察：details 内容含 `{{< nosuch >}}演示语法` 行 | 见新 🟡 | — |

## 问题清单

### 🟡 1.（新发现）「新 open 终止」截断组件内 shortcode 语法示例

- **触发**：多行组件（details/tabs 等）的 Content 内出现一行完整 `{{< 名 >}}` open 形态（作者展示语法示例），shortcode.go:94 的 `shortcodeOpenRe.Match(line)` 立即终止当前块。
- **实证**（独立探针 D）：`{{< details "示例" >}}\n第一行\n{{< nosuch >}}演示语法\n第二行\n{{< /details >}}\n` →
  `<!-- unclosed shortcode: details -->{{< details "示例" >}}\n第一行<!-- unknown shortcode: nosuch -->…` —— details 组件渲染失败（unclosed 透传），演示语法行被误解析为新的短码块，`{{< /details >}}` 沦为孤立文本。修复前该用例 Content 完整收集、details 正常渲染（原代码无 openRe 检查），故为 f8c8a2e 引入的行为回归。
- **影响评估**：非静默（有 unclosed/unknown 注释提示）、无内容丢失（全部原文可见）、触发条件为「Content 内 open 形态行」这一边缘用例；与 strip 层语义不一致（strip 层 inSC 内不因 open 行退出，块边界判定两层不同，但未知/未闭合走透传不经占位符管线，无隐藏后果）。
- **建议**：①v0.8 文档注明「组件内不可出现 `{{<` 行（无嵌套）」；②或细化终止条件——新 open 行同行含 inline close（自闭合演示形态）时不终止；③或接受现状，列为 v0.9 嵌套/转义支持的前置约束。不阻塞合并。

### ⚪ 1. 任务消息测试数口径

任务消息「8 个 TestRenderShortcode_*」与原报告统计一致（原 8 项含 TestParseShortcodeParams），实测 `-run 'TestRenderShortcode'` 9 个全 PASS + TestParseShortcodeParams PASS。无功能影响，记录口径。

### ⚪ 2. TrimSpace 对透传 Content 首行缩进的边缘影响

shortcode.go:120 的 `TrimSpace` 在修复 ⚪6 的同时，会去掉多行块 Content 首/末行的空白——若 Content 首行为有意义缩进（如缩进代码块 `    code`，存在于未闭合/未知透传输出中），前导空格丢失。已知组件（innerMD 本就 TrimSpace）无变化；仅未知/未闭合透传的视觉级差异。超边缘场景，观察点。

## Final Assessment

- **门禁状态：放行（原 🔴/🟡2/🟡3/🟡4 全部修复并有回归测试，主探针闭环 A/B/C/E 端到端 14/14 + 独立探针 16/16 通过）。**
- 新 🟡（组件内语法示例截断）为设计取舍的已知副作用，建议随 v0.8.0 发布说明注明限制或后续细化，不构成合并阻塞。
- 本复审未修改任何被测代码（探针即建即删、/tmp 副本、二进制与副本已清理），`git status --short` 仅剩约定报告目录 `reports/`。
