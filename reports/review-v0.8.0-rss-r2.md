# 复审报告 — v0.8.0 线 C RSS 修复（fabfa61）R2

- **审查对象**: `feat/v0.8.0-rss` @ fabfa61（`fix: rss limit validation and alternate link condition`，5 文件 +100/-6），修复首轮报告 🟡1/🟡2
- **复审基线**: 首轮报告 `reports/review-v0.8.0-rss.md`（247baf9 单提交，0🔴 2🟡 7⚪）
- **审查者**: code-reviewer（独立复跑原探针闭环，不采信 developer 自述）
- **日期**: 2026-08-13
- **方法**: 全量测试独立复跑 + CLI 黑盒探针（/tmp/chenhai-probe，每场景独立干净副本防 config 污染）+ 精确行号 re-grep + main 分支对照（`git show main:`）

## 结论：🟡 有条件通过（修复闭环全部验证通过；新增 1 🟡 建议顺手修复）

首轮 2 条 🟡 **均已修复且经独立探针验证**：limit 负值/零值报错非 panic（错误信息明确）、禁用态 alternate link 全页面归零。**新增 1 条 🟡**：CLI 层 error 传播断裂（`main()` 丢弃 `cli.Execute()` 返回值，所有命令错误路径退出码恒 0，pre-existing 非本分支引入）——修复将 panic 转为 error，但 error 的进程级传播在 main 被截断，脚本化消费方无法检测构建失败。一行修复（`os.Exit(1)`），建议本批顺手修或显式承接。模板层修复缺自动化负向单测（⚪ 记录）。

---

## 1. 首轮 🟡 修复核对表（独立复跑原探针，非采信 developer 自述）

| 原发现 | 修复位置 | 结果 | 复审证据（独立实测） |
|---|---|---|---|
| 🟡1 limit 无边界校验（-1 裸 panic / 0 静默空 feed） | rss.go:57-58 `if cfg.RSS.Limit < 1 { return nil, fmt.Errorf("rss.limit 必须 ≥ 1（当前 %d）", ...) }` | ✅ 已修 | limit:-1 → `Error: build failed: generate rss: rss.limit 必须 ≥ 1（当前 -1）`，无 panic；limit:0 → 同报错；limit:20 → 正常构建，atom.xml 3 entry |
| 🟡2 禁用态 alternate link 悬垂 404 | base.html:55 `{{if and .Config.BaseURL .Config.RSS.Enabled}}` | ✅ 已修 | enabled:false → `grep -c 'rel="alternate"' public/index.html` = 0；index/archives/tags/about 四页全 0；`ls public/atom.xml` = No such file |
| 🟡1 配套：负向单测 | rss_test.go: TestGenerateRSS_LimitNegative / LimitZero / LimitValid | ✅ 已修 | `go test -run` 3 用例 PASS（错误信息含 "rss.limit" 断言、零值报错断言、正常值 5 entry 断言） |
| 🟡1 配套：doctor 检查 | doctor.go:58-59 `if cfg.RSS.Enabled && cfg.RSS.Limit < 1 { errors++ }` | ✅ 已修 | CLI 实测 limit:0 → `✗ rss.limit 必须 ≥ 1（当前 0），RSS 构建将失败` + `Error: 发现 1 个错误`；limit:-1 → 同提示；limit:20 → 无提示（grep count 0） |
| 附带：doctor 单测 | doctor_test.go: TestDoctor_InvalidRSSLimit | ✅ 已修 | PASS（ExecuteDoctor 返回 error 断言） |
| ⚪8 summarizeContent 正则提包级 | rss.go:120 `var htmlTagRegex = regexp.MustCompile(...)` | ✅ 已处理 | 全文件 grep MustCompile 仅此一处，函数内无编译残留 |

**partial-fix 检查**：🟡1/🟡2 各自两子句（校验+负向用例 / 条件+负向断言）全部落地，无半修。
**over-fix 检查**：limit 校验按建议"报错"实现（非 clamp），无额外拒绝面；doctor 检查属建议内容；base.html 条件无额外收紧；改动面 5 文件无旁路副作用。默认配置 `RSS: RSSConfig{Enabled: true, Limit: 20}`（types.go:108）——无 rss 段既有站点行为不变，🟡2 修复零回归。

## 2. 符合项证据表（独立实测）

| # | 项 | 结果 | 证据 |
|---|---|---|---|
| 1 | 全量测试 | ✅ | `go test -count=1 ./internal/...` 8 包全 ok（build/cli/config/content/imagehost/index/server/theme） |
| 2 | 新负向单测运行 | ✅ | 5 个新用例独立 `-run` 复跑全 PASS（build 4 + cli 1） |
| 3 | limit -1 报错非 panic | ✅ | 错误信息明确（见上表），无 `slice bounds out of range` 输出 |
| 4 | limit 0 报错 | ✅ | `rss.limit 必须 ≥ 1（当前 0）` |
| 5 | limit 20 正常 | ✅ | 构建完成，atom.xml 2315B，3 entry |
| 6 | enabled=false 负向断言 | ✅ | 4 页面 alternate 计数全 0，atom.xml 不存在 |
| 7 | doctor rss.limit 检查项 | ✅ | 非法值 error 级提示 + `Error: 发现 1 个错误`；合法值无提示 |
| 8 | 正则提包级 | ✅ | 静态确认 |

## 3. 🔴 阻断

无。

## 4. 🟡 风险

### 🟡R2-1（新发现，pre-existing）— cmd/chenhai/main.go:8：`cli.Execute()` 返回值被丢弃，所有命令错误路径退出码恒 0

- **问题**: `main()` 只调 `cli.Execute()` 不检查 error，无 `os.Exit(1)`。cobra `Execute()` 返回 error 后由 main 负责转退出码，丢弃即静默成功外观——build 失败、doctor 失败等全部以退出码 0 结束。
- **证据**（CLI 实测，探针二进制）:
  ```
  limit:-1 → /tmp/chenhai-probe build → "Error: build failed: generate rss: rss.limit 必须 ≥ 1（当前 -1）"；echo $? = 0
  limit:0  → 同上，$? = 0
  doctor 非法 limit → "Error: 发现 1 个错误"；$? = 0
  ```
- **依据**: 本修复将 panic 转为 error（正确方向），但 error 的消费终点在 `main()` 被截断——peer 验收"退出非 0 或明确错误信息"的 OR 语义靠明确错误信息满足（字面通过），但脚本化消费方（CI 步骤、deploy 钩子、Makefile）无法感知失败；与 skill 中 Go CLI 审查先例同族（"失败被吞 + exit=0 = 静默成功外观"）。**pre-existing**（`git show main:cmd/chenhai/main.go` 同为丢弃形态，非本分支引入），故不阻断本批合入，但与本修复验收语义直接相关。
- **改法**: `if err := cli.Execute(); err != nil { os.Exit(1) }`（一行）；补一个断言退出码的 CLI 级测试（可用 `go run . build` 子进程探针或 TestMain 包装）。

## 5. ⚪ 记录

- **⚪R2-1** base.html 修复无自动化负向单测：修复提交的单测覆盖 rss.go/doctor.go，模板层（theme 包 engine_test.go）无 alternate link 断言——本复审以 CLI 负向探针验证（4 页面全 0），但自动化缺口仍在，建议补模板渲染负向用例或承接。
- 首轮 ⚪1-⚪7 承接项（增量漂移/UX/残留/零日期/author/双斜杠/注释改述）不在本修复范围，维持承接记录。

## 6. 合并前检查单

- [x] 🟡1 修复验证：limit -1 / 0 / 20 三态 CLI 探针（报错非 panic / 报错 / 正常）
- [x] 🟡2 修复验证：enabled=false 时 index.html 及全站页面无 alternate link（负向断言）
- [x] doctor rss.limit 检查项存在且非法值提示（error 级）
- [x] summarizeContent 正则提包级 + 新增负向单测（-1/0/正常）运行通过
- [ ] 🟡R2-1 处置：`main()` 加 `os.Exit(1)`（建议本批顺手修，一行；或显式承接 v0.9）——orchestrator 定夺
- [ ] ⚪R2-1 记录承接（模板负向单测）
- [ ] 首轮 ⚪1-7 承接登记维持
- [ ] 凭据自检：本报告无任何凭证字面量

---

**Final Assessment**: 🟡 有条件通过。首轮 2 🟡 修复闭环经独立探针全部验证通过，无半修无过度修复；新发现 1 🟡（CLI 退出码恒 0，pre-existing）建议本批一行修复后合入，或显式承接。修复本身达标。
