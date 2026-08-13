# 复审报告（R2 独立复跑）· Chenhai-hugo v0.8.0 线 A 贴图工作流

| 项目 | 内容 |
|------|------|
| 复审对象 | feat/v0.8.0-image（image add/compress 路径校验 + quality 校验 + O_EXCL 原子写修复） |
| 修复提交 | `3e4df3f` feat: harden image add path/quality validation and atomic writes |
| **审查时 HEAD** | **`3e4df3f`**（与 R2 首跑一致，分支未再移动；`git status` 除 reports/ 未跟踪外干净） |
| 审查者 | code-reviewer（独立复跑，未采信 developer 自述） |
| 日期 | 2026-08-13（R2 首跑 21:03-21:09；本文件为 peer 指定落盘名的独立副本，证据为 21:35 起 R3 新鲜复跑） |
| 审计链 | 首轮报告 `reports/review-v0.8.0-image.md` 第 7 节（R2 结论）与本文件同结论、同证据来源；两处并存，未改写首轮内容 |
| 依据 | `docs/superpowers/plans/2026-08-13-v0.8.0-writing-experience.md` Task 1-3 + CONTRIBUTING.md + 首轮审查报告 🟡-1/2/3 |

---

## 1. 结论

**✅ 通过** —— 首轮 3 项 🟡（路径穿越、quality 边界、并发撞名）全部修复，负向单测齐备，无 🔴，无半修；本轮 R3 新鲜复跑（HEAD 3e4df3f 重建二进制）逐项实证通过。新发现 1 项 ⚪（`--dir .` 行为收紧）与 2 项既有承接项（错误路径 exit=0、force 无 name 固定 img1）记录在案，不影响合入。

## 2. 逐项复核（本轮 R3 新鲜复跑证据，二进制 `/tmp/review-r3/chenhai` 为 HEAD 重建）

| 原问题 | 状态 | 证据（2026-08-13 21:35 实测） |
|--------|------|------------------|
| 🟡-1 路径穿越（--dir/--post/--name 三面） | ✅ 已修 | 三例全拒绝、站点根内外零产物（见 2.1） |
| 🟡-2 quality 越界静默失败 | ✅ 已修 | 999/-1 双入口（compress + add）报错「quality 必须在 0-100 之间」；0/100 正常转换（见 2.2） |
| 🟡-3 并发撞名静默覆盖 | ✅ 已修 | 10 进程 xargs -P10：10 文件 img1..img10、引用行各 ×1、零失败进程、全 webp 魔数合法（见 2.3） |
| ⚪-5 双重错误前缀 | ✅ 顺手修复 | 修复 diff 确认；本轮 add 越界报错文案仅一处前缀（见 2.2 P1-5 输出） |
| ⚪-6 失败计入 skipped | ✅ 顺手修复 | 本轮 P1-3/4 正常路径输出「转换 1 张，跳过 0 张，失败 0 张」独立计数口径 |

### 2.1 P2 路径穿越（黑盒，站点根 `/tmp/review-r3/site` 仅 config.yaml + tiny.png）

```
$ chenhai image add tiny.png --dir ../../escape
  Error: 目标目录不能包含 .. 段或绝对路径: ../../escape
$ chenhai image add tiny.png --post a/../../../../etc/evil.md
  Error: --post 路径无效: 目标目录不能包含 .. 段或绝对路径: ../../../etc/evil
$ chenhai image add tiny.png --dir img/x --name ../../../evil.webp
  Error: 输出文件名不能包含路径分隔符: ../../../evil.webp
$ chenhai image add tiny.png --dir /tmp/abs          （绝对路径附加变体）
  Error: 目标目录不能包含 .. 段或绝对路径: /tmp/abs
```

- 产物断言：站点根外 `escape`/`etc` 目录均不存在；全盘 `evil.webp` 0 个；站点根仅 config.yaml/static/tiny.png/tiny.webp（零逃逸产物）。
- 注：拒绝后 exit 仍为 0（tester 🟡-1，全 CLI 既有行为，main.go 不在本分支 diff），零产物语义已达成，承接 orchestrator。

### 2.2 P1 quality 边界（黑盒）

```
$ chenhai image compress tiny.png --quality 999
  Error: quality 必须在 0-100 之间: 999
$ chenhai image compress tiny.png --quality -1
  Error: quality 必须在 0-100 之间: -1
$ chenhai image compress tiny.png --quality 0    → ✓ tiny.png → tiny.webp，转换 1 张，跳过 0 张，失败 0 张
$ chenhai image compress tiny.png --quality 100  → ✓ 同上
$ chenhai image add tiny.png --dir img/q --quality 999
  Error: quality 必须在 0-100 之间: 999          （add 入口同样拦截，无双重前缀）
```

### 2.3 P4 并发撞名（10 进程 xargs -P10 复跑原探针）

```
文件数: 10（img1.webp ... img10.webp，sort -V 全列）
引用行计数（精确 pattern ![](/img/conc/imgN.webp)）: 各 ×1，共 10 行
webp 魔数检查: total 10, bad 0（RIFF....WEBP 全合法）
```

- 口径说明：粗 pattern `imgN.webp` 计数为 ×2/名，系「✓ 转换行 + 引用行」各含一次文件名所致；引用行本身各 ×1（10 进程 10 行，零重复零丢失）。
- 显式名兜底（R2 已验，本轮未复跑）：10 进程同 `--name fixed.webp` → 1 成 9 拒，O_EXCL 在 guard 竞态窗口后仍保证不覆盖。

## 3. 负向单测确认（代码层面核实 + 全量执行）

| 测试 | 覆盖 |
|------|------|
| `TestPostToImgDir`（image_test.go:15） | 含 4 个穿越 bad case（`a/../../etc/evil.md`、`../posts/x.md`、`../../escape.md`、`/abs/path.md`） |
| `TestCheckQuality`（image_test.go:46） | -1/101/1000 拒 + 0/50/100 过 |
| `TestRunImageAdd_PathTraversal`（image_test.go:161） | 7 子用例全 PASS（含穿越产物未逃出站点根断言） |
| `TestRunImageAdd_QualityRange`（image_test.go:196） | 错误文案断言「quality 必须在 0-100 之间」 |
| `TestRunImageAdd_ConcurrentAutoName`（image_test.go:209） | 4 goroutine → 恰好 img1..img4 |
| `TestWriteFileExclusive`（imageproc_test.go:124） | 首次成功、二次 fs.ErrExist、原内容未被破坏 |

执行：`go test -count=1 ./internal/...` → 9 包全 ok（build/config/content/imagehost/imageproc/index/server/theme/cli）；`go vet ./internal/imageproc/ ./internal/cli/` → 0 告警。本轮实测无回归。

## 4. 新发现（1 项 ⚪）

**⚪-8. `--dir .` 被拒，static/ 根目录不可达（行为收紧，超出首轮改法建议）**
- 位置：`internal/cli/image.go` checkRelDir（`rel == "." || rel == ""` 分支）
- 证据（R2 新旧对照实测）：旧版 `5ffcf9c` `--dir .` → 写 `static/img1.webp` + 引用行 `![](/./img1.webp)`；新版 → `Error: 目标目录无效: .`。`--dir img/..`（Clean 归并后 "."）同样被拒，static/ 根完全不可达。
- 评估：首轮改法建议只要求「Clean + 拒 `..` 段 + 拒绝对路径」，未要求拒 `.`；旧行为不逃逸（落 static/ 内），仅引用行含 `.` 段不够干净。收紧有正当性（引用行规范），但属超出建议的行为变化；帮助文本「--dir 目标目录（相对 static/，如 img/xxx）」未注明根目录不可达，错误文案「目标目录无效」不含原因。
- 处置：合并前二选一——(a) 帮助文本注明「须为 static/ 下子目录」；(b) 允许 `.`（引用行改 `!/img1.webp` 形态）。记录承接即可，不阻断。

## 5. 遗留承接（合并前检查单）

| 项 | 处置 |
|----|------|
| ⚪-8 `--dir .` 收紧 | 合并前注明（文档或允许 `.`，见 §4） |
| tester 🟡-1 错误路径 exit=0 | 承接 orchestrator（main.go 不在本分支 diff；本轮实测 P1/P2 拒绝后 exit 仍为 0，脚本化检查不可靠，但零产物语义已达成） |
| ⚪-1 force 无 name 固定 img1 | 承接（本轮行为未变） |
| ⚪-2 compress/resize 无 guard | 承接 |
| ⚪-3 仅缩小不放大 | 承接 |
| ⚪-4 Sscanf 半匹配 | 承接 |
| ⚪-7 static symlink 跟随 | 承接（非缺陷） |
| 探针环境 | /tmp/review-r3（本轮）+ /tmp/review-r2（R2）保留至会话结束；worktree 仅 reports/ 未跟踪（与首轮/R2 同惯例，未 commit） |

## 6. 脱敏自检

本报告无凭据引用；探针输出仅含 /tmp 路径与命令输出。落盘后凭据形态 grep 自检：0 命中（见报告尾部校验命令）。
