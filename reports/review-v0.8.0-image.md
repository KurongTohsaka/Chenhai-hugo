# 对抗性审查报告：v0.8.0 线 A 贴图工作流（feat/v0.8.0-image）

| 项目 | 内容 |
|------|------|
| 审查对象 | Chenhai-hugo v0.8.0 线 A（imageproc 包 + image 命令族 add/compress/resize） |
| 分支 | `feat/v0.8.0-image`（本地 worktree：Chenhai-hugo-worktrees/image） |
| **审查时 HEAD** | **`5ffcf9c`** feat: enforce --force semantics for image add overwrite |
| 审查者 | code-reviewer（对抗性实测，独立复跑） |
| 日期 | 2026-08-13 |
| 依据 | `docs/superpowers/plans/2026-08-13-v0.8.0-writing-experience.md` Task 1-3 + CONTRIBUTING.md + tester verify 报告 |
| 方法 | 全量 diff 阅读 + 依赖库源码核对（chai2010/webp v1.4.0）+ 14 组独立探针实测 |

> ⚠️ 提交集注记：tester verify 报告记录 HEAD 为 `f8690b0`（3 提交）；审查开始时分支已前进到 `5ffcf9c`（developer 新增 force 语义修复提交）。本报告以最终 HEAD `5ffcf9c` 为准复探；`5ffcf9c` 内容（overwrite guard + 负向单测）已单独核验。

---

## 1. 结论

**🟡 有条件通过** —— 无 🔴 阻断；3 项 🟡 建议本批修复（路径穿越输入校验、quality 边界、并发撞名），修复后重跑对应探针即可闭环；⚪ 承接项 7 条记录在案。

- 功能面完整且实测可用（编码三格式 + GIF 拷贝 + 命名递增 + guard 修复均实证过）
- 核心风险集中在**用户可控路径/参数输入未校验**（`--dir`/`--post`/`--name` 三面穿越、`--quality` 越界静默失败），以及**命名分配的非原子性**（并发撞名静默覆盖）
- tester 🟡-2（force 语义）已由 developer 修复并验证完整（见 §4 复审节）；tester 🟡-1（退出码恒 0）为全 CLI 既有行为（`cmd/chenhai/main.go` 不在本分支 diff 内），承接 orchestrator

## 2. 符合项 evidence table（实测证明能工作的部分）

| # | 项 | 证据 |
|---|----|------|
| 1 | 单元测试基线 | `go test -count=1 ./internal/...` → 9 包全 ok（build/config/content/imagehost/imageproc/index/server/theme/cli），imageproc 5/5、cli 含新增 TestRunImageAdd_OverwriteGuard 无回归 |
| 2 | 静态检查 | `go vet ./internal/imageproc/ ./internal/cli/` → 0 告警 |
| 3 | 构建 | `go build -o /tmp/review-img/chenhai ./cmd/chenhai/` → 22,130,402 字节 |
| 4 | jpg 输入（YCbCr 路径） | `compress src.jpg --quality 75` → ✓ src.webp（`adjustImage` 对 `*image.YCbCr` 走 `NewRGBImageFrom`，库源码核对无误） |
| 5 | 标准 PNG（NRGBA 路径） | PIL 生成 color-type-2 PNG → ✓ std.webp（`*image.NRGBA → toRGBAImage` 兜底） |
| 6 | 带 alpha PNG | ✓ alpha.webp（透明通道保留路径正常） |
| 7 | GIF 拷贝字节一致 | `add anim.gif --name clip.png` → 输出强制替换为 `clip.gif`，`cmp` 零差异 |
| 8 | overwrite guard 修复有效（复审正向） | 非 force + `--name shot.webp` 撞已存在文件 → 拒绝「输出文件已存在（--force 覆盖）」且哨兵内容未被破坏；force → 覆盖成功（单测 + 独立复跑均过） |
| 9 | 超大 `--width 9223372036854775807` 不崩溃 | 输出 10x10 原图（NoOp 守卫 + `dstW<1` 双兜底），exit=0；无 panic 面 |
| 10 | NextImageName 超大编号安全方向 | `img99999999999999999999.webp` 存在时 Sscanf 溢出报错不计数，返回 img1.webp，无碰撞 |
| 11 | symlink src 输入 | `add link.png`（symlink→tiny.png）正常解码转码 |
| 12 | 站点根检测 / `--post`-`--dir` 互斥 | 非站点根报错、互斥报错（tester 实测 + 本审查复验错误路径存在） |
| 13 | 扩展名强制替换 | `--name plain`（无扩展名）→ 补 `.webp`；`--name custom.jpg` → 替换为 `.webp`（防 WebP 数据写进 .jpg，实测 tester 场景 12/13 逻辑复核一致） |

## 3. 问题清单

### 🔴 阻断（0）

无。

### 🟡 风险（3）

---

**🟡-1. `--dir`/`--post`/`--name` 三输入面路径穿越，文件静默写出 static/ 且引用行 URL 错误**
- 位置：`internal/cli/image.go:202`（`staticDir := filepath.Join("static", relDir)`）、`:227`（`outPath := filepath.Join(staticDir, outName)`）、`:168-175`（postToImgDir）
- 问题：`--dir`/`--post`/`--name` 均为用户可控输入，直接拼入路径，无任何 `..` 段/绝对路径/分隔符校验。`filepath.Join` 的 Clean 语义会把 `..` 段向上归并，导致落盘位置逃出 `static/` 甚至站点根；同时第 5 步引用行 `fmt.Sprintf("![](/%s/%s)", relDir, outName)`（`:260`）原样打印含 `..` 的 URL——**URL 与实际落盘不一致，粘贴即裂图**（贴图工作流的核心交付物错误）。
- 证据（实测，站点根 /tmp/review-img/site 仅含 config.yaml）：
  ```
  $ chenhai image add tiny.png --dir ../../escape
    ✓ ../escape/img1.webp
  ![](/../../escape/img1.webp)          # 落盘 /tmp/review-img/escape/img1.webp（逃出 static 两层）
  $ chenhai image add tiny.png --post a/../../../../etc/evil.md
    ✓ ../../etc/evil/img1.webp           # 落盘站点根再上两级（/private/tmp/etc/evil/），find 于站点内不可见
  ![](/../../../etc/evil/img1.webp)
  $ chenhai image add tiny.png --dir img/x --name ../../../evil.webp
    ✓ evil.webp                          # 落盘站点根 /tmp/review-img/site/evil.webp
  ![](/img/x/../../../evil.webp)
  ```
  三例 exit=0，无任何警告。
- 依据：flag 帮助文本契约「--dir 目标目录（相对 static/，如 img/xxx）」「--post 文章路径（相对 content/，如 posts/...）」——相对语义未 enforcement；postToImgDir 的 docstring 只描述正常形态，无负向处理。`os.WriteFile` 覆盖语义使穿越目标可为任意已有文件（数据覆写面）。
- 威胁模型：本地单用户 CLI，非多租户攻击面，故定 🟡 非 🔴；但「输出错误引用行」是功能性缺陷（核心交付物错误），且从教程/历史命令粘贴 `--dir` 参数可静默落错位置。
- 改法：(a) `relDir` 解析后 `filepath.Clean` + 断言 `!strings.HasPrefix(rel, "..")` 且非绝对路径；`outName` 拒绝含 `/`、`\` 及 `..` 段（name 语义=文件名，非路径）；(b) `postToImgDir` 对非预期前缀（非 `posts/`/`about/` 等合法 section）同样校验；(c) 单测补负向用例（现有 TestPostToImgDir 4 例全为正常形态，零穿越用例）。

---

**🟡-2. `--quality` 无 0-100 边界校验，越界值静默失败（compress）**
- 位置：`internal/cli/image.go:87`（compress）、`:157`（add）
- 问题：flag 帮助文本声称「WebP 质量 0-100」但无任何 range 校验。chai2010/webp `Options.Quality`（库 writer.go:23 注释 0~100）直接透传 libwebp，越界时 `WebPEncode*` 返回失败。compress 路径把 encode 失败吞入 `skipped`（`:66-67`）并 exit=0——**用户得到「命令成功、0 张转换」的静默成功外观**，错误信息 `webpEncodeRGBA: failed` 不提示 quality 越界，批量场景下难以定位。
- 证据（实测）：
  ```
  $ chenhai image compress src.png --quality 999
    ✗ /tmp/review-img/src.png: encode webp: webpEncodeRGBA: failed
  完成：转换 0 张，跳过 1 张
  exit=0
  $ chenhai image compress tiny.png --quality -1
    ✗ ...: encode webp: webpEncodeRGBA: bad arguments
  exit=0
  $ chenhai image compress tiny.png --quality 0   → ✓ 正常
  $ chenhai image compress tiny.png --quality 100 → ✓ 正常
  ```
  （add 路径越界为响亮错误但错误文案重复「encode webp: encode webp:」，见 ⚪-5）
- 依据：flag 帮助文本「WebP 质量 0-100」即契约；越界行为与契约不符且失败被静默吸收。
- 改法：`RunE` 入口校验 `if quality < 0 || quality > 100 { return fmt.Errorf("quality 必须在 0-100 之间: %d", quality) }`（compress/add 各一处）；或 pflag 无内置 range，用 cobra 校验函数。顺带把 encode 失败从 skipped 计数中拆出独立失败计数（见 ⚪-6）。

---

**🟡-3. 并发 add 撞名静默覆盖（NextImageName + overwrite guard 均 check-then-act 非原子）**
- 位置：`internal/imageproc/imageproc.go:117-137`（NextImageName 读目录取 maxN+1）、`internal/cli/image.go:230`（guard 先 Stat 后写）
- 问题：自动命名「读目录 → 算 imgN+1 → WriteFile」与 guard「Stat 不存在 → WriteFile」均为非原子两步。并发进程同时通过检查、拿到同名，后写覆盖先写——**静默丢图无感知**（引用行已打印给用户，但文件内容只剩最后写入者）。
- 证据（实测，10 进程 xargs -P10 同时 add 到空目录）：
  ```
  $ ls static/img/conc/   → img1.webp img2.webp ... img7.webp   # 10 进程仅 7 文件
  $ grep 引用行计数        → ![](/img/conc/img1.webp) ×3         # 3 进程拿到 img1，2 次静默覆盖
  ```
- 依据：NextImageName docstring 无并发承诺，但 add 是「贴图工作流」，多终端/脚本并行贴图场景真实存在；静默覆盖是错数据路径（比响亮失败更需优先）。
- 改法：写入改用原子独占创建 `os.OpenFile(outPath, O_WRONLY|O_CREATE|O_EXCL, 0644)`，`EEXIST` 时重试递增（自动命名路径）或报「已存在」（`--name` 路径）；guard 可保留作为快速失败提示，但最终一致性由 O_EXCL 保证。

---

### ⚪ 可选（7）

**⚪-1. `--force` 无 `--name` 时固定写 `img1.webp`，不递增**
- 位置：`internal/cli/image.go:218-219`（`outName = "img1." + outExt`）
- 证据：目录已有 img1-img3.webp 时 `add --force` → 输出 `static/img/f/img1.webp`（覆盖 img1），ls 仍为 img1/img2/img3。
- 评价：与注释「force 模式下 NextImageName 会撞名，故跳过自动命名」一致，但文档口径「--force 覆盖已存在的同名文件」在无 name 时语义含糊（同名=硬编码 img1？）。建议文档注明或 force 时仍递增。记录承接。

**⚪-2. compress / resize 输出无 overwrite guard，静默覆盖既有文件**
- 位置：`internal/cli/image.go:76-79`（compress 输出 `*.webp`）、`:111-127`（resize 输出 `<name>_w<W>.<ext>`）
- 证据：`compress` 二次运行重写既有 `cov.webp`（md5 变化对比前哨）；`resize` 输出 `tiny_w100.png` 已存在（内容为 SENTINEL 标记）→ 静默替换为 PNG，exit=0。
- 评价：与 add（有 guard）行为不一致；重压幂等语义可辩护，但批量 compress 目录中用户手工放置的既有 webp 会被静默重写。建议 compress 加「已存在则跳过/提示」或文档注明覆盖语义。记录承接。

**⚪-3. resize 超大 `--width` 输出尺寸与文件名不符**
- 位置：`internal/imageproc/imageproc.go:83`（`srcW <= maxW` 即 NoOp）
- 证据：`resize tiny.png --width 9223372036854775807` → 输出 10x10 原图，文件名 `tiny_w9223372036854775807.png`。仅缩小不放大且无提示（正向：无崩溃，NoOp 守卫 + dstW<1 双兜底）。
- 评价：文档「等比缩放图片」未说明仅缩小；文件名声称尺寸与实际不符。建议放大方向提示或文档注明。记录承接。

**⚪-4. NextImageName 前缀匹配过宽（Sscanf 半匹配）**
- 位置：`internal/imageproc/imageproc.go:132`
- 证据：目录含 `img10abc.webp` + `img2.webp` → 自动命名返回 `img11.webp`（img10abc 被解析为 n=10，10 号空置浪费）。
- 评价：无碰撞风险（guard 兜底，安全方向），仅编号浪费。可改完整匹配（解析后比对 `"img"+strconv.Itoa(n)+"."+ext == name`）或正则 `^img(\d+)\.`。记录承接。

**⚪-5. add 错误文案双重前缀**
- 位置：`internal/imageproc/imageproc.go:72`（`fmt.Errorf("encode webp: %w", ...)`）+ `internal/cli/image.go:252`（`fmt.Errorf("encode webp: %w", ...)`）
- 证据：`add --quality 999` → `Error: encode webp: encode webp: webpEncodeRGBA: failed`。一处包装即可。顺手修正。

**⚪-6. compress 失败计入 skipped，统计口径误导**
- 位置：`internal/cli/image.go:66-67`
- 证据：`--quality 999` 输出「跳过 1 张」（实为失败 1 张），与「已含 webp 跳过」混同。
- 评价：与 🟡-2 联动；建议拆独立 failed 计数。记录承接。

**⚪-7. `static/` 为符号链接时跟随写入外部目录**
- 位置：`internal/imageproc/imageproc.go:141-144`（MkdirAll/WriteFile 跟随链接）
- 证据：`static → /tmp/review-img/external` 时 add 正常写入外部目录，无提示。
- 评价：标准 Unix 语义，非缺陷；仅记录（若站点根被部署工具做成 symlink，属预期行为）。

---

## 4. 复审节（developer 修复提交 `5ffcf9c`）

针对 tester 🟡-2（--force 语义与文档不符），developer 提交 `5ffcf9c`（enforce --force semantics）添加：

- `internal/cli/image.go:229-234` — overwrite guard：非 force 且目标存在 → 拒绝「输出文件已存在（--force 覆盖）」；Stat 其他错误 → 报错。
- `internal/cli/image_test.go:55-102` — TestRunImageAdd_OverwriteGuard：负向（非 force 拒绝且哨兵内容未破坏）+ 正向（force 覆盖成功）。

**独立复探结果**：
- ✅ 已修：负向路径实测——非 force + `--name shot.webp` 撞已存在文件返回错误且哨兵内容未被破坏（`string(got) != "old"` 断言通过）；force 覆盖成功。
- ✅ 无半修：guard 双分支（`err == nil && !force` / `err != nil && !IsNotExist`）均处理，未发现只修一子句；新增测试自带负向用例（此前 tester 建议的正向用例缺口已补）。
- ✅ 无副作用：guard 位于命名确定之后、写入之前，自动命名路径（NextImageName 递增）不触发拒绝（guard 仅对 `--name` 显式撞名生效），既有 12 个 cli 测试无回归。
- 遗留：force 无 name 固定 img1（⚪-1）未在本次修复范围，按承接记录。

## 5. 合并前检查单

| 项 | 处置 | 说明 |
|----|------|------|
| 🟡-1 路径穿越输入校验 | **本批修复**（回 developer） | 三输入面（--dir/--post/--name）+ 负向单测；修复后 reviewer 重跑 P2 探针 |
| 🟡-2 quality 0-100 校验 | **本批修复**（回 developer） | compress/add 各一处；修复后重跑 P1 探针 |
| 🟡-3 并发撞名（O_EXCL 原子写） | **本批修复**（回 developer） | 或显式承接至下批（记录位置：Task 4-8 分支）；修复后重跑 P4 探针 |
| ⚪-1 force 无 name 固定 img1 | 记录承接 | merge message 注明 |
| ⚪-2 compress/resize 无 guard | 记录承接 | merge message 注明 |
| ⚪-3 仅缩小不放大 | 记录承接 | 文档注记即可 |
| ⚪-4 Sscanf 半匹配 | 记录承接 | 低优先级 |
| ⚪-5 双重错误前缀 | 记录承接（顺手 1 行） | — |
| ⚪-6 失败计入 skipped | 随 🟡-2 修复一并处理 | — |
| ⚪-7 static symlink 跟随 | 记录承接 | 非缺陷 |
| tester ⚪（testdata static/ 未入 .gitignore） | 记录承接 | tester 已报，orchestrator 直办 |
| tester 🟡-1（退出码恒 0） | 记录承接 | 全 CLI 既有（main.go 不在本分支 diff），orchestrator 定夺 |
| 审查时提交集 | 记录 | HEAD `5ffcf9c`（verify 报告当时为 `f8690b0`，审查已覆盖增量提交） |
| 探针环境 | 已清理 | /tmp/review-img 全部删除；worktree 仅 reports/ 未跟踪 |

## 6. 脱敏自检

报告内无凭据引用；探针输出仅含 /tmp 路径与命令输出。`grep -c` 凭据形态自检：0。

---

## 7. 复审（R2 · developer 修复提交 `3e4df3f`）

> 审查时 HEAD：`3e4df3f`（feat: harden image add path/quality validation and atomic writes）。本节为 reviewer 独立复跑探针的结论，与 developer 自述无关（本次无 developer 自述小节，仅修复提交 + 单测）。

### 7.1 结论

**✅ 通过** —— 首轮 3 项 🟡 全部修复并经独立探针闭环验证，负向单测齐备，无 🔴；发现 1 项新 ⚪（`--dir .` 行为收紧）与 2 项既有承接项（exit=0、force 无 name 固定 img1）需在合并前记录。

### 7.2 逐项复核（独立复跑证据）

| 原问题 | 状态 | 证据（本轮实测） |
|--------|------|------------------|
| 🟡-1 路径穿越（--dir/--post/--name 三面） | ✅ 已修 | 三例全拒绝、零产物（见下） |
| 🟡-2 quality 越界静默失败 | ✅ 已修 | 999/-1 双入口报错；0/100 正常（见下） |
| 🟡-3 并发撞名静默覆盖 | ✅ 已修 | 10 进程 10 文件、引用行各 ×1（见下） |
| ⚪-5 双重错误前缀 | ✅ 顺手修复 | diff 确认 `EncodeWebP` 已带 `encode webp:` 上下文，调用点改裸 `return err` |
| ⚪-6 失败计入 skipped | ✅ 顺手修复 | 实测损坏文件 → 「失败 1 张」独立计数，skipped 语义（既有 webp）未破坏 |

**P2 路径穿越（黑盒实测，站点根 /tmp/review-r2/site 仅 config.yaml）：**

```
$ chenhai image add tiny.png --dir ../../escape
  Error: 目标目录不能包含 .. 段或绝对路径: ../../escape
$ chenhai image add tiny.png --post a/../../../../etc/evil.md
  Error: --post 路径无效: 目标目录不能包含 .. 段或绝对路径: ../../../etc/evil
$ chenhai image add tiny.png --dir img/x --name ../../../evil.webp
  Error: 输出文件名不能包含路径分隔符: ../../../evil.webp
```

- 附加变体全部拒绝：`--dir /tmp/abs`（绝对路径）、`--dir img/x --name a/b.webp`（分隔符）、`--name ..\evil.webp`（反斜杠）。
- 产物断言：站点根外无 `escape`/`etc`，站点根内仅 config.yaml，`static/` 不存在（零产物）。
- 单测：TestPostToImgDir 新增 4 个穿越 bad case（`a/../../etc/evil.md`、`../posts/x.md`、`../../escape.md`、`/abs/path.md`）全 PASS；TestRunImageAdd_PathTraversal 7 子用例全 PASS（含「穿越产物未逃出站点根」断言）。

**P1 quality 边界（黑盒实测）：**

```
$ chenhai image compress src.png --quality 999
  Error: quality 必须在 0-100 之间: 999
$ chenhai image compress tiny.png --quality -1
  Error: quality 必须在 0-100 之间: -1
$ chenhai image compress tiny.png --quality 0    → ✓ 转换 1 张，跳过 0 张，失败 0 张
$ chenhai image compress src.png --quality 100   → ✓ 正常
$ chenhai image add tiny.png --dir img/q --quality 999
  Error: quality 必须在 0-100 之间: 999          （add 入口同样拦截）
```

- 单测：TestCheckQuality（-1/101/1000 拒 + 0/50/100 过）+ TestRunImageAdd_QualityRange（断言错误文案含「quality 必须在 0-100 之间」）全 PASS。

**P4 并发撞名（10 进程 xargs -P10 复跑原探针）：**

```
文件数: 10（img1.webp ... img10.webp）
引用行计数: ![](/img/conc/imgN.webp) 各 ×1（零重复、零失败进程）
```

- 显式名并发兜底：10 进程同 `--name fixed.webp` → 文件数 1，成功 1、拒绝 9（O_EXCL 在 guard 竞态窗口后仍保证不覆盖）。
- 全部产物 webp 魔数校验通过（RIFF....WEBP）。
- 单测：TestWriteFileExclusive（首次成功、二次 fs.ErrExist、原内容未被破坏）+ TestRunImageAdd_ConcurrentAutoName（4 goroutine → 恰好 img1..img4）全 PASS。

### 7.3 新发现（1 项 ⚪）

**⚪-8. `--dir .` 被拒，static/ 根目录不可达（行为收紧，超出首轮改法建议）**
- 位置：`internal/cli/image.go` checkRelDir（`rel == "." || rel == ""` 分支）
- 证据（新旧对照实测）：旧版 `5ffcf9c` `--dir .` → 写 `static/img1.webp` + 引用行 `![](/./img1.webp)`；新版 → `Error: 目标目录无效: .`。`--dir img/..`（Clean 归并后 "."）同样被拒，static/ 根完全不可达。
- 评估：首轮改法建议只要求「Clean + 拒 `..` 前缀 + 拒绝对路径」，未要求拒 `.`；旧行为不逃逸（落 static/ 内），仅引用行含 `.` 段不够干净。收紧有正当性（引用行规范），但属超出建议的行为变化，帮助文本「--dir 目标目录（相对 static/，如 img/xxx）」未注明根目录不可达，且错误文案「目标目录无效」不含原因。
- 改法（二选一，orchestrator 定夺）：(a) 帮助文本注明「须为 static/ 下子目录」；(b) 允许 `.`（引用行改 `!/img1.webp` 形态）。合并前记录承接即可。

### 7.4 回归与卫生

- 全量单测：`go test -count=1 ./internal/...` → 9 包全 ok（cli 含全部新负向用例，无回归）。
- `go vet ./internal/imageproc/ ./internal/cli/` → 0 告警。
- GIF 拷贝走新 O_EXCL 统一写入路径：`--name clip.png` → `clip.gif`，`cmp` 字节一致（无回归）。
- 正常路径引用行与落盘一致：`--dir img/ok --name shot.webp` → `![](/img/ok/shot.webp)` + 文件确实落盘 `static/img/ok/shot.webp`。
- 修复未引入新死旋钮/半修：checkQuality/checkRelDir/checkOutName 三校验器均在写入前执行且互不依赖；O_EXCL 重试循环 `autoNamed` 判定在循环外计算（常量语义正确）；`--name` 显式撞名报错文案与 guard 一致（「输出文件已存在（--force 覆盖）」）。

### 7.5 遗留承接（合并前检查单）

| 项 | 处置 |
|----|------|
| ⚪-8 `--dir .` 收紧 | 合并前注明（文档或允许 `.`，见 7.3） |
| tester 🟡-1 错误路径 exit=0 | 承接 orchestrator（main.go 不在本分支 diff；本轮实测 P1/P2 拒绝后 exit 仍为 0，脚本化检查不可靠，但零产物语义已达成） |
| ⚪-1 force 无 name 固定 img1 | 承接（本轮确认行为未变） |
| ⚪-2 compress/resize 无 guard | 承接 |
| ⚪-3 仅缩小不放大 | 承接 |
| ⚪-4 Sscanf 半匹配 | 承接 |
| ⚪-7 static symlink 跟随 | 承接（非缺陷） |
| 探针环境 | /tmp/review-r2 保留至本会话结束；worktree 仅 reports/ 未跟踪（与首轮同惯例，未 commit） |

### 7.6 脱敏自检

本复审节无凭据引用；探针输出仅含 /tmp 路径与命令输出，无凭据形态串。
