# CONTRIBUTING — Chenhai-hugo 多 Agent 协作契约

> 2026-08-13 v0.8.0 起生效。多 profile 并行开发（orchestrator + developer×3 + tester + code-reviewer + merge-reconciler）。

## 角色职责

| 角色 | Profile | 职责 | 红线 |
|------|---------|------|------|
| Orchestrator | default（镇海） | 编排、派活、门禁、合并、呈批 | 不亲跑业务链路；只做只读验证；不改核心代码 |
| Developer | developer / developer-2 / developer-3 | 按实现计划 Task 实现（TDD） | 只做任务范围内；不 merge 不 push main；不编造测试结果 |
| Tester | tester | 独立验证（结构/依赖/spec 逐条核对/测试输出） | 报告 🔴🟡⚪ 分级，证据导向，凭证脱敏 |
| Code-reviewer | code-reviewer | 对抗性代码审查（`git diff main...分支` + 实测） | 独立人设不迎合；🔴🟡⚪ 分级 |
| Merge-reconciler | merge-reconciler | 合并冲突中立裁决 | worker 不自裁冲突 |

## 分支与提交

- 主线 `main`；功能线分支 `feat/v0.8.0-<name>`，从 main 最新切出
- 提交前缀循项目惯例（英文）：`feat:` / `fix:` / `docs:` / `chore:` / `revert:` / `merge:`
- 每 Task 至少一次提交；完成 = 代码 + 自测 + `git push -u origin <分支>`
- 合并用本地 `git merge --no-ff`，merge message 标 `Merge PR #N: <名>`；**分支保留不删**（回查用）
- 合并序列（orchestrator 执行）：`git status --short` 查工作区（先提交 reports/ 尾巴）→ checkout main → merge → push → 勾选进度

## 门禁流程（每线必经）

1. Developer 完成推送 → Orchestrator 只读快验（git log/文件产物）
2. **Tester 独立验证** → `reports/verify-<线>.md`（🔴🟡⚪ + 证据）
3. **Code-reviewer 审查** → `reports/review-<线>.md`（对抗性实测）
4. 🟡 代码级 → 回 developer 修复（同分支追加提交）→ 复审；文档级一行 🟡 → orchestrator 直办
5. 🔴 驳回 → 修复后 reviewer 重跑原探针才算闭环
6. 合并授权（钦定）：低风险且 tester 全绿、无 🔴 → orchestrator 自主合入（事后汇报）；core/高风险每批亲批
7. 合并消息记录残留 🟡（承接项）与夫君拍板决策

## 报告纪律

- 验证/审查报告落 `reports/`，随合并入库
- 凭证一律脱敏 `<REDACTED>`（合并前 orchestrator 复查清零）
- 数据/结果数字必须实证，禁止编造

## 本版 core 清单（亲批范围）

- `internal/content/renderer.go` — 全站渲染管线（shortcode 扩展影响所有页面）
- `internal/content/shortcode*.go` — 新渲染路径
- `internal/build/builder.go` — 构建主流程挂载（provider 注入 + RSS）
- `internal/theme/engine.go` — 模板查找
- `go.mod` — 依赖变更（cgo chai2010/webp）

## 环境要点

- Go 项目：`go test ./internal/...` + `go build -o chenhai ./cmd/chenhai/` 为验证基线
- cgo 依赖（chai2010/webp）需要 C 工具链（macOS CLT / ubuntu gcc）
- 验证基线保持全绿；每 Task 完成后跑一次
