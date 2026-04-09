# RELEASE-CHECKLIST

- 日期: 2026-04-07
- 适用范围: 当前 `T001-T137` 完成后的 AtlasX-Go 仓库

## 发布前必须执行

- `cd atlasx && bash scripts/release_evidence.sh`
- `cd atlasx && go test ./...`
- `cd atlasx && bash scripts/e2e_gate.sh`
- `cd atlasx && go run ./cmd/atlasd --once`

## 发布前必须核对

- `tasks.csv` 中没有 `未开始` 或 `进行中` 条目
- gate 脚本退出码为 `0`
- 本次发布涉及的功能链路没有落在 `UNCOVERED` 中
- 若本次发布涉及 runtime 安装链路：
  - 已准备 install plan
  - 已明确是否设置 `ATLASX_E2E_ALLOW_INSTALL=1`
- 若本次发布涉及真实 sidebar/provider：
  - `sidebar_qa_ready=true`
  - 当前存在受管 page target

## 证据留存

- 优先使用 `cd atlasx && bash scripts/release_evidence.sh`
- 默认会把 `go test ./...`、`bash scripts/e2e_gate.sh`、`go run ./cmd/atlasd --once` 的输出统一落到单一输出目录
- 输出目录至少应包含：
  - `go-test.log`
  - `e2e-gate.log`
  - `atlasd-once.log`
  - `SUMMARY.md`
- `SUMMARY.md` 应自动包含：
  - `runtime_manifest_version`
  - `runtime_manifest_channel`
  - `chrome_source`
  - `sidebar_default_provider`
  - `atlasd_ready`
  - `runtime_manifest_present`
  - `mirror_present`
  - `managed_session_live`
  - `sidebar_qa_ready`
  - `uncovered_count`
  - `uncovered_items`
  - `tasks_total`
  - `tasks_done`
  - `tasks_doing`
  - `tasks_todo`
  - `release_ready`
  - `release_blockers`
  - `release_prerequisites`

## 阻断条件

- `go test ./...` 失败
- `bash scripts/e2e_gate.sh` 退出码非 `0`
- 本次发布必须覆盖的链路仍在 `UNCOVERED`
- `go run ./cmd/atlasd --once` 暴露的状态与发布声明不一致

## 恢复入口

- runtime 问题:
  - `atlasx/docs/RUNBOOK.md` 中的 Runtime 操作章节
- browser/session 问题:
  - `atlasx/docs/RUNBOOK.md` 中的 浏览器与数据面恢复 章节
- provider 问题:
  - `atlasx/docs/RUNBOOK.md` 中的 Sidebar 恢复 章节

## 发布后首轮检查

- 再次执行 `cd atlasx && go run ./cmd/atlasd --once`
- 若发布影响 runtime 或 sidebar，重新执行 `cd atlasx && bash scripts/e2e_gate.sh`
- 若出现新漂移，先回写 `docs/reviews/CR-ISSUES-CURRENT-STATE-2026-04-07.md`
