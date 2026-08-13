# 审查报告 — v0.8.0 线 C RSS/Atom 生成（feat/v0.8.0-rss）

- **审查对象**: Chenhai-hugo `feat/v0.8.0-rss` @ 247baf9（单提交 `feat: atom RSS feed generation`，7 文件 +331/-2）
- **工作目录**: /Users/kurong/Project/GoProject/Chenhai-hugo-worktrees/rss
- **审查者**: code-reviewer（对抗性独立审查，不采信 tester 自述，全部结论经实测）
- **日期**: 2026-08-13
- **验收依据**: docs/superpowers/plans/2026-08-13-v0.8.0-writing-experience.md Task 7（Step 1-10）+ CONTRIBUTING.md
- **方法**: 独立复跑全量测试 + CLI 黑盒探针（/tmp/rssprobe 副本站点，`go build -o /tmp/chenhai-probe`）+ 包内临时单元探针（跑完即删，工作区无残留）+ Go 标准库源码核实

## 结论：🟡 有条件通过

**0 🔴；2 🟡；7 ⚪。** 两条 🟡（limit 负值裸 panic、RSS 禁用时 alternate link 悬垂）建议修复后合入或显式承接；⚪ 按检查单记录。XML 注入/转义、控制字符清洗、rune 截断、Draft 过滤等核心安全面经对抗实测**全部正确**。

---

## 1. 符合项证据表（独立实测，非采信 tester）

| # | 项 | 结果 | 证据 |
|---|---|---|---|
| 1 | 全量测试 | ✅ | `go test -count=1 ./internal/...` 8 包全 ok（build/cli/config/content/imagehost/index/server/theme），探针文件已删除后复跑 |
| 2 | 基线构建 + XML 合法 | ✅ | example-blog 副本 `chenhai build` → public/atom.xml 2315B；`xmllint --noout` 通过；3 entry |
| 3 | XML 注入/转义（对抗） | ✅ | 恶意 title `"><feed><entry><title>PWNED…` → 输出 `&#34;&gt;&lt;feed&gt;&lt;entry&gt;&lt;title&gt;PWNED…`（全转义）；`grep -c "<entry><title>PWNED"` = 0；xmllint 通过；description `<script>alert(1)</script> & x < y` → `&lt;script&gt;… &amp; x &lt; y` |
| 4 | 控制字符清洗（entry 级） | ✅ | cleanXMLText("a\x01b\x0b\x0cc\x1fd") = "abcd"；正文 `&#x01;lead` → 输出无 0x01、lead 保留（实体解码→清洗顺序正确） |
| 5 | 截断 rune 边界 | ✅ | 600 rune 正文 → summary 300 rune + "…"（共 301）；恰好 300 rune → 不加 "…"；多字节中文无切半 |
| 6 | Draft 过滤 / 排序 / Limit | ✅ | rss_test.go 5/5 PASS（含 draft 不泄漏、newest-first、Limit=2→2 entry）；代码 `!p.Draft` + `Date.After` 排序 |
| 7 | 模板转义安全 | ✅ | 主题引擎用 `html/template`（engine.go:5）→ base.html alternate link 的 title/href 属性自动转义 |
| 8 | NoBaseURL → 空 feed | ✅ | TestGenerateRSS_NoBaseURL PASS；构建不写文件 |
| 9 | enabled=false → 不生成 atom.xml | ✅ | CLI 实测 `ls public/atom.xml` = No such file |
| 10 | doctor nil 防护 | ✅ | TestDoctor_BadConfigNoPanic / MissingBaseURLWarns 5/5 PASS；`cfg.BaseURL` 检查仅在 Load 成功分支 |
| 11 | 增量构建 Description 优先 | ✅ | 增量构建 summaries=2 = 2 个有 description 的帖子（与 Y2 设计一致，见 ⚪1） |

---

## 2. 🔴 阻断

无。

---

## 3. 🟡 风险（建议修复）

### 🟡1 — rss.go:66-67：`cfg.RSS.Limit` 无边界校验，负值裸 panic，零值静默空 feed

- **问题**: `if len(published) > cfg.RSS.Limit { published = published[:cfg.RSS.Limit] }` 对用户 YAML 配置无任何校验。
  - `limit: -1` → **裸 panic**（非 error 路径，无 stack 信息友好度可言，构建进程直接崩）；
  - `limit: 0` → **静默空 feed**（构建"成功"，订阅器收到 0 条目）。
- **证据**（CLI 实测，/tmp/rssprobe 副本 config 追加 `rss: {enabled: true, limit: -1}`）:
  ```
  生成 RSS 订阅 ... panic: runtime error: slice bounds out of range [:-1]
  ```
  `limit: 0` → 构建退出码 0，`grep -c "<entry>" public/atom.xml` = 0。
- **依据**: plan Step 1 仅定义 `Limit int` 默认 20，无约束；但 CLI 接受任意 int，config 加载器不校验，doctor 也不检查——三个入口全无防线。panic 属响亮失败（不升 🔴），但零值静默空 feed 属静默错误数据路径，需一并处理。
- **改法**: generateRSS 入口校验 `if cfg.RSS.Limit < 1 { return nil, fmt.Errorf("rss.limit 必须 ≥ 1（当前 %d）", cfg.RSS.Limit) }`（或 clamp 回默认 20）；doctor 增加 rss.limit 检查项；补负向单测（limit -1 / 0 / 正常三态）。

### 🟡2 — base.html:54-57：RSS 禁用时仍输出 alternate link（dangling 404）

- **问题**: 条件仅 `{{if .Config.BaseURL}}`，不查 `.Config.RSS.Enabled`。用户显式 `rss: {enabled: false}` 后，**每个页面**的 head 仍宣传 `/atom.xml`，而文件不存在。
- **证据**（CLI 实测）: config 追加 `rss: {enabled: false}` → 构建成功、`public/atom.xml` 不存在，但 `grep -c 'rel="alternate"' public/index.html` = 1（link 照常输出）。
- **依据**: 配置语义——enabled=false 应全链路无 RSS 痕迹；且 TemplateData 暴露完整 `*config.Config`（engine.go:30-34），`.Config.RSS.Enabled` 直接可用，修复零成本。
- **改法**: `{{if and .Config.BaseURL .Config.RSS.Enabled}}`；构建验证 enabled=false 时 index.html 无 alternate link（负向断言）。

---

## 4. ⚪ 观察点

### ⚪1（tester ⚪1 独立复现 + 生产路径分析）增量构建 summary 漂移
- **复现**: 全量构建 md5=341a20ce… summaries=5 → 紧接着增量构建 md5=58639dc9… summaries=2（无 description 帖子的 summary 消失）。与 collect.go:63 缓存路径置空 Content、Y2 注释预期一致。
- **补充**: 漂移后果是 entry 既无 `<summary>` 也无 `<content>`，多数聚合器显示空白正文；且 feed 级 `Updated: time.Now()` 使 atom.xml 每次构建必重写（md5 变化实证）。
- **定级理由（不升 🟡）**: 生产路径不受影响——`chenhai deploy` 只推源码（deploy.go:46-48），构建在 GitHub Actions 全新 checkout 执行（ci.yml 实证，无缓存 → 恒全量）。漂移仅影响本地增量 serve。若未来本地增量产物直接上线，升 🟡。
- **承接**: collect.go 缓存路径保留 Description（正文不可得但摘要可缓存）；或增量无内容变更时复用上次 atom.xml。

### ⚪2（tester ⚪2 复现）enabled=false 控制台 UX
- 实测输出 `生成 RSS 订阅 ... ` 后直接接 `✓ 构建完成`，无"完成/跳过"字样（builder.go:197-199 else-if 只覆盖 `Enabled && BaseURL==""` 分支）。仅 UX。

### ⚪3（tester ⚪3 确认）关闭开关后旧 atom.xml 残留
- 与 sitemap/robots 同构（builder.go:183-191 同样只写不删，项目既有惯例）。`rss.enabled=false` 不删 public/atom.xml。低影响，记录承接。

### ⚪4 零日期文章 → entry `<updated>0001-01-01T00:00:00Z</updated>`
- **证据**（单元+CLI 双复现）: 无 `date:` front matter 的帖子 → `updated = LastMod(零) → Date(零) → Format` 输出 `0001-01-01T00:00:00Z`（rss.go:85-92）。RFC3339 格式合法但时间为公元元年，聚合器显示异常。
- **改法**: feed 跳过零日期 entry，或回退文件 mtime；至少加单测。

### ⚪5 无 author 配置 → feed 零 `<author>`，违反 RFC 4287 §4.1.1
- **证据**（CLI 实测）: config 移除 author 段 → `grep -c "<author>" public/atom.xml` = 0（feed 级与 entry 级均无）。RFC 4287 要求 feed 至少含一个 atom:author，除非所有 entry 各自含 author。默认配置（DefaultConfig 无 Author）即触发。
- **缓解**: Task 8 Step 4 已为真实站点补 author.name；建议 doctor 加提示，或默认配置补 Author。

### ⚪6 尾斜杠 BaseURL → alternate link 双斜杠
- **证据**（CLI 实测）: `baseURL: "https://example.com/"` → 页面 link `href="https://example.com//atom.xml"`，而 feed self link 为 `https://example.com/atom.xml`（rss.go:55 TrimRight 生效，模板未 Trim）——两处不一致。多数服务器容忍 `//`，但模板侧应 `{{strings.TrimRight .Config.BaseURL "/"}}`（或文档约定 baseURL 不带尾斜杠）。

### ⚪7 cleanXMLText 注释与 stdlib 实际行为不符（防"构建失败"的说法不成立）
- **证据**: Go 源码 encoding/xml/xml.go:1944-1948 `escapeText` 对 XML 非法 rune（\x01、U+FFFE/U+FFFF、无效 UTF-8）**静默替换为 U+FFFD**，不报错。实测 cfg.Title="feed\x01title"（未经 cleanXMLText 的 feed 级字段）→ 输出 `<title>feed�title</title>`、err=nil——即非法字符永远不会让整站构建失败，只会产生 � 乱码。
- **推论**: ① cleanXMLText 的真实价值 = 防 � 乱码（仍是好卫生），注释应改述；② U+FFFE/U+FFFF 不在清洗范围（`r >= 0x20` 放行），靠 stdlib 兜底替换为 �（实测 B1: 无错、字符被替换）；③ feed 级字段（cfg.Title/Description/BaseURL/Author/Permalink）不过清洗的残余风险 = 乱码级，非崩溃级，无需升级。
- **注**: TestGenerateRSS_XMLClean 断言前提（"否则单篇文章即可让整站构建失败"）与 stdlib 行为不符，但测试本身有效，仅注释需修正。

### ⚪8 summarizeContent 每次调用编译正则
- rss.go:115 `regexp.MustCompile` 在函数内，每 page 每构建编译一次。静态站点规模无感，建议提包级变量（一行）。

---

## 5. 测试覆盖缺口（记录，非缺陷）

- 无负向用例: `limit ≤ 0`、RSS 禁用时 alternate link 仍输出、零日期 entry、无 author、feed 级字段清洗路径——rss_test.go 5 测试与 plan Step 2 完全一致，共享 plan 的盲区。
- 🟡1/🟡2 修复提交必须自带上述负向用例（按 CONTRIBUTING 门禁第 4 条：🟡 代码级回 developer 同分支追加提交，reviewer 复跑原探针闭环）。

## 6. 过程注记

- **worktree 状态**: 审查开始时发现 ` D testdata/example-blog/content/about/index.md`（未提交删除，非本审查产生，疑似 tester 验证操作残留）——已 `git checkout --` 恢复。当前 `git status --short` 仅 `?? reports/`。
- **reports/ 未入库**: verify 报告与本文均未 commit（untracked）。按 CONTRIBUTING 合并序列（先提交 reports/ 尾巴 → checkout main → merge），orchestrator 合入前需一并提交两份报告。本次审查不自行 commit/push（peer 未授权）。
- **探针卫生**: 临时探针测试文件（rss_probe_tmp_test.go / tmp2）跑完即删；探针站点在 /tmp，不落 worktree。

## 7. 合并前检查单

- [ ] 🟡1 修复（rss.limit 校验 + 负向单测）→ reviewer 复跑 `limit: -1 / 0 / 20` 三态探针
- [ ] 🟡2 修复（alternate link 加 `.Config.RSS.Enabled`）→ 复跑 enabled=false 时 index.html 无 alternate link
- [ ] ⚪ 承接项登记（增量漂移/UX/残留/零日期/author/双斜杠/注释改述/正则提级）：orchestrator 记录，可入 Task 8 或 v0.9 清单
- [ ] reports/（verify + review 两份）随合并入库
- [ ] 凭据自检: 本报告无任何凭证字面量，`grep -c` 扫描归零
- [ ] 工作区干净（已恢复误删文件；探针零残留）
- [ ] 分支基线: 单提交 247baf9，`git log main..HEAD` = 1 commit，符合提交契约

---
**Final Assessment**: 🟡 有条件通过。核心安全面（XML 注入/转义、控制字符、截断、模板转义）对抗实测全绿；两条 🟡 均为低成本一行级修复，建议本批次闭环后合入。
