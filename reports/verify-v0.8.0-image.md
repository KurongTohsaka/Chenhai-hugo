# 验证报告：v0.8.0 线 A 贴图工作流（feat/v0.8.0-image）

| 项目 | 内容 |
|------|------|
| 验证对象 | Chenhai-hugo v0.8.0 线 A（imageproc 包 + image 命令族） |
| 分支 | `feat/v0.8.0-image`（本地 worktree：Chenhai-hugo-worktrees/image） |
| HEAD | `f8690b0` feat: chenhai image add — paste-image workflow with webp compression |
| 验证者 | tester（独立第三方） |
| 日期 | 2026-08-13 |
| 验收依据 | `docs/superpowers/plans/2026-08-13-v0.8.0-writing-experience.md` Task 1-3 + CONTRIBUTING.md |
| 测试命令 | `go test -count=1 ./internal/...`；`go build -o chenhai ./cmd/chenhai/`；功能实测（compress/resize/add 全链路） |

## 结论摘要

| 维度 | 结果 | 证据 |
|------|------|------|
| 结构（diff vs plan Task 1-3） | ✅ 吻合 | 7 文件逐一对应，见下表 |
| 提交基线 | ✅ 干净 | main..HEAD 仅 3 个 feat 提交，无混杂 |
| 依赖 | ✅ 正确 | chai2010/webp v1.4.0 + golang.org/x/image v0.23.0，go.sum 锁定 |
| 单元测试 | ✅ 全绿 | 9 包 ok（-count=1 强制重跑），imageproc 5/5，cli 含新增 TestPostToImgDir |
| 构建 | ✅ 成功 | 22MB 二进制，站点 build 13ms 完成 |
| 功能实测 | ✅ 17/17 过 | 见实测表 |
| 🔴 阻塞 | 0 | — |
| 🟡 建议 | 2 | 退出码恒 0；--force 语义与文档不符 |
| ⚪ 可选 | 1 | static/ 产物目录未入 gitignore |

## 结构核对（git diff main...HEAD）

| plan 文件 | 实际 diff | 状态 |
|-----------|-----------|------|
| `internal/imageproc/imageproc.go`（Task1 新建） | A | ✅ |
| `internal/imageproc/imageproc_test.go`（Task1 新建） | A（117 行） | ✅ |
| `go.mod`（Task1 修改 +依赖） | M（+chai2010/webp v1.4.0、x/image v0.23.0） | ✅ |
| `go.sum`（Task1 随依赖提交） | M（4 行新增） | ✅ |
| `internal/cli/image.go`（Task2 新建） | A（256 行） | ✅ |
| `internal/cli/root.go`（Task2 注册命令） | M（+1 行 AddCommand） | ✅ |
| `internal/cli/image_test.go`（Task3 新建） | A（17 行 TestPostToImgDir） | ✅ |

- 提交基线：`9175664`（imageproc）→ `0366ff4`（compress/resize）→ `f8690b0`（add），与 plan Task 1-3 提交顺序一一对应。
- Task 4-8（shortcode/RSS/文档）**不在本分支**，符合任务范围（本验证仅 Task 1-3）。

## 测试输出

```
go test -count=1 ./internal/... → 9 包全 ok（build/config/content/imagehost/imageproc/index/server/theme/cli）
imageproc: 5/5 PASS
  TestEncodeWebP_RoundTrip / TestDecode_File / TestResize_Downscale / TestResize_NoOp / TestNextImageName
cli: 13/13 PASS（含新增 TestPostToImgDir；原 12 个既有测试无回归）
go build -o chenhai ./cmd/chenhai/ → BUILD OK（22,130,402 字节）
```

## 功能实测（独立对抗验证，全部实际执行）

| # | 场景 | 预期 | 实测结果 | 状态 |
|---|------|------|----------|------|
| 1 | compress 批量（2 张 PNG → WebP，q75） | 转换 2 跳过 0，保留原文件 | a.png→a.webp(946B)、b.png→b.webp(934B)，原文件保留 | ✅ |
| 2 | compress 重复跑（已含 webp） | webp 跳过计数 | 转换 2 张，跳过 2 张 | ✅ |
| 3 | compress 单文件 | 转换 1 跳过 0 | 转换 1 张，跳过 0 张 | ✅ |
| 4 | compress 损坏图（坏 PNG） | 报 ✗ 不中断 | `✗ ...: png: invalid format: bad filter type`，计入 skipped，流程继续 | ✅ |
| 5 | resize --width 400 | a_w400.png 400x300 | 独立读取 PNG 头：(400,300) | ✅ |
| 6 | resize --height 200 | b_h200.png 等比 | (600,200) | ✅ |
| 7 | resize 无 --width/--height | 报错 | `Error: 请指定 --width 或 --height` | ✅ |
| 8 | resize webp 输入 | webp 保持 webp | a_w300.webp，RIFF/VP8 300x225 | ✅ |
| 9 | add --post posts/hello.md | section 前缀剥离 → static/img/hello/img1.webp + `![](/img/hello/img1.webp)` | 输出与落盘完全一致 | ✅ |
| 10 | add 再跑（自动递增） | img2.webp | img2.webp 生成 | ✅ |
| 11 | add --dir img/gallery | 独立计数 img1.webp | img1.webp | ✅ |
| 12 | add --name custom.jpg | 扩展名替换为 webp | custom.webp（防 WebP 数据写进 .jpg） | ✅ |
| 13 | add --name plain（无扩展名） | 补 .webp | plain.webp | ✅ |
| 14 | 站点根检测（/tmp 无 config.yaml） | 报错 | `Error: 未找到 config.yaml——请在站点根目录运行 image add` | ✅ |
| 15 | --dir 与 --post 同传/同缺 | 报错 | `Error: 必须且只能指定 --post 或 --dir 之一` | ✅ |
| 16 | add GIF | 原样拷贝不转码 | 字节级一致（cmp 零差异），输出 img1.gif | ✅ |
| 17 | add --force 覆盖同名 | 覆盖成功 | img1.webp 被新图覆盖（936B） | ✅ |
| 18 | webp 落盘有效性 | 可解码且尺寸正确 | 独立 Go 程序 webp.Decode：img1/custom/gallery 全 800x600 | ✅ |
| 19 | chenhai build（含 add 产物） | 构建不报错 | 13ms 完成，public/ 生成 | ✅ |

## 问题清单

### 🔴 阻塞（0）

无。

### 🟡 建议（2）

1. **错误路径退出码恒为 0**（`cmd/chenhai/main.go:9`）
   - 证据：`main.go` 调 `cli.Execute()` 丢弃返回值；实测 5 条错误路径（resize 无参数、--dir/--post 互斥、站点根检测、源文件不存在、源为目录）`echo $?` 全为 0。
   - 影响：脚本化使用（CI / `&&` 链 / shell 判断）会静默误判"命令成功"。
   - 建议：`if err := cli.Execute(); err != nil { os.Exit(1) }`。影响面为全 CLI 既有命令，改动 1 行，风险低。

2. **--force 语义与文档不符**（`internal/cli/image.go` runImageAdd）
   - 证据：plan 声明 "--force 覆盖已存在的同名文件"；实现中 force 仅控制"是否跳过自动命名"（`outName == "" && !force` 分支），**无任何"文件已存在"检查**——实测不带 --force 传 `--name img1.webp` 静默覆盖已有文件且 exit=0。
   - 影响：CLI 契约与文档口径不符；自动命名路径无风险（NextImageName 递增不会撞名），仅 --name 显式路径覆盖无保护。
   - 建议：二选一——(a) 加 `os.Stat(outPath)` 存在检查，非 force 时报错；(b) 文档改口径为"force 时跳过自动命名"。与 developer 对齐意图后定。

### ⚪ 可选（1）

1. **`testdata/example-blog/static/` 未入 .gitignore**（`.gitignore:38,49` 只覆盖 public/）
   - 证据：实测按 plan 步骤在 example-blog 跑 image add 后工作树出现 `?? testdata/example-blog/static/`（未跟踪）；`.gitignore` 无 static 规则（check-ignore exit=1）。
   - 影响：开发者/验证者按 plan 手动验证后工作树留脏；不影响功能。
   - 建议：.gitignore 补 `testdata/example-blog/static/`。

## Final Assessment

- 结构 ✅ / 测试 ✅ / 构建 ✅ / 功能实测 19/19 项全过（含 17 项任务指定场景 + 2 项补充：损坏图容错、GIF 字节级一致）。
- 无 🔴，无回归（cli 既有 12 测试全过）。
- 🟡-1（退出码）为全 CLI 既有行为（非本分支引入），建议作为承接项由 orchestrator 定夺；🟡-2（force 语义）属本分支文档口径问题，建议回 developer 澄清后一行修复或改文档。
- 按 CONTRIBUTING 门禁：tester 验证 ✅ 通过，可流转 code-reviewer；低风险且无 🔴，orchestrator 可自主合入（🟡 承接项记入 merge message）。
- 验证产物清理：实测生成的 /tmp/imgtest、example-blog/static、tmp_verify 均已删除，工作树 `git status` 干净。
