# AtlasX-Go

`AtlasX-Go` 是一个面向 Intel x64 macOS 的 Atlas-like 控制面重建项目。

当前仓库已经不再是“最小骨架”阶段，而是一个可构建、可测试、可发布的 Go 控制面实现，主要由两部分组成：

- [`docs/`](docs/00-INDEX.md)：公开资料梳理、产品蓝图、架构决策、审计与发布记录
- [`atlasx/`](atlasx/README.md)：`atlasctl` / `atlasd`、managed Chromium runtime、sidebar QA、memory 与 tabs automation 的实现

## 当前能力

- `atlasctl` / `atlasd` 双入口
- managed Chromium runtime 的 stage、verify、install、plan 与状态导出
- 受管浏览器会话、mirror/import、history/downloads/bookmarks 打通
- memory controls、memory list/search、page-scoped snippet 注入控制
- sidebar provider registry、真实 `ask` / `selection-ask` / `summarize`
- tabs list/search/windows/groups、窗口整理、DevTools、capture、auth-mode、agent-plan / agent-execute
- 发布门禁与证据采集：`bash atlasx/scripts/e2e_gate.sh`、`bash atlasx/scripts/release_evidence.sh`

## 快速开始

```bash
cd atlasx
go test ./...
go run ./cmd/atlasctl doctor
go run ./cmd/atlasctl status
go run ./cmd/atlasd --once
```

如需真实发布门禁：

```bash
cd atlasx
bash scripts/e2e_gate.sh
bash scripts/release_evidence.sh
```

## 主要文档

- 总索引：[docs/00-INDEX.md](docs/00-INDEX.md)
- 实现与命令说明：[atlasx/README.md](atlasx/README.md)
- 运行与恢复：[atlasx/docs/RUNBOOK.md](atlasx/docs/RUNBOOK.md)
- E2E gate：[atlasx/docs/E2E-GATE.md](atlasx/docs/E2E-GATE.md)
- 发布与审计记录：[docs/reviews/](docs/reviews/)

## 安全边界

- `atlasd` 默认只允许回环监听地址
- 非回环监听必须显式传 `--allow-remote-control`
- sidebar 配置只落盘 `api_key_env`，不落盘真实密钥
- 当前仍未实现真实 macOS TCC 探测、权限提示或授权写路径
