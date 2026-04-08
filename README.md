# AtlasX-Go

`AtlasX-Go` 是一个面向 Intel x64 macOS 的 `Atlas-like` 重建项目。

当前仓库包含两部分：

- [`docs/`](docs/00-INDEX.md)：可行性分析、公开资料梳理、产品蓝图、架构与交接文档
- [`atlasx/`](atlasx/README.md)：Go 主导的最小控制面骨架

当前已具备的最小能力：

- `atlasctl blueprint`
- `atlasctl doctor`
- `atlasctl doctor --json`
- `atlasctl profile status`
- `atlasctl policy status`
- `atlasctl permissions status`
- `atlasctl settings`
- `atlasctl default-browser status`
- `atlasctl logs status`
- `atlasctl updates status`
- `atlasctl memory list`
- `atlasctl memory search <question>`
- `atlasctl sidebar status`
- `atlasctl sidebar ask <target-id> <question>`
- `atlasctl sidebar selection-ask <target-id> <question>`
- `atlasctl sidebar summarize <target-id>`
- `atlasctl status`
- `atlasctl runtime stage|status|verify|clear|install`
- `atlasctl runtime plan create|resolve|status|clear`
- `atlasctl mirror-scan`
- `atlasctl history|downloads|bookmarks list/open`
- `atlasctl tabs list|search|windows|open|open-window|open-in-window|move-to-window|move-to-new-window|merge-window|open-devtools|open-devtools-panel|close-duplicates|activate-window|close-window|set-window-state|set-window-bounds|activate|close|navigate|capture|extract-context|selection|suggest|memories|recommend-context|organize|organize-window|organize-group-to-window|organize-group-into-window|organize-to-windows|organize-into-window|organize-window-to-windows|organize-window-into-window|organize-window-group-to-window|organize-window-group-into-window|devtools|devtools-panel|emulate-device`
- `atlasctl tabs suggest <target-id>`
- `atlasctl tabs memories <target-id>`
- `atlasctl tabs recommend-context <target-id>`
- `atlasctl tabs windows`
- `atlasctl tabs search <query>`
- `atlasctl tabs open-in-window <window-id> <url>`
- `atlasctl tabs move-to-window <tab-id> <target-window-id>`
- `atlasctl tabs move-to-new-window <tab-id>`
- `atlasctl tabs organize-group-to-window <group-id>`
- `atlasctl tabs organize-group-into-window <group-id> <window-id>`
- `atlasctl tabs organize-to-windows`
- `atlasctl tabs organize-into-window <window-id>`
- `atlasctl tabs organize-window-to-windows <window-id>`
- `atlasctl tabs organize-window-into-window <source-window-id> <target-window-id>`
- `atlasctl tabs organize-window <window-id>`
- `atlasctl tabs organize-window-group-to-window <source-window-id> <group-id>`
- `atlasctl tabs organize-window-group-into-window <source-window-id> <group-id> <target-window-id>`
- `atlasctl tabs merge-window <source-window-id> <target-window-id>`
- `atlasctl tabs open-devtools <target-id>`
- `atlasctl tabs open-devtools-panel <target-id> <panel>`
- `atlasctl tabs devtools-panel <target-id> <panel>`
- `atlasctl tabs close-duplicates`
- `atlasctl tabs activate-window <window-id>`
- `atlasctl tabs close-window <window-id>`
- `atlasctl tabs set-window-state <window-id> <state>`
- `atlasctl tabs set-window-bounds <window-id> <left> <top> <width> <height>`
- `atlasctl tabs open-window <url>`
- `atlasctl tabs emulate-device <target-id> <preset>`
- `atlasctl tabs extract-context <target-id>`
- `atlasd /v1/history|downloads|bookmarks` 与对应 `/open` 动作 API
- `atlasd /v1/settings`
- `atlasd /v1/default-browser`
- `atlasd /v1/logs`
- `atlasd /v1/updates`
- `atlasd /v1/doctor`
- `atlasd /v1/profile`
- `atlasd /v1/policy`
- `atlasd /v1/permissions`
- `atlasd /v1/memory` 与 `/v1/memory/search`
- `atlasd /v1/tabs|search|windows|open|open-window|open-in-window|move-to-window|move-to-new-window|merge-window|open-devtools|open-devtools-panel|close-duplicates|activate-window|close-window|window-state|window-bounds|activate|close|navigate|context|semantic-context|selection|suggestions|memories|context-recommendations|organize|organize-window|organize-group-to-window|organize-group-into-window|organize-to-windows|organize-into-window|organize-window-to-windows|organize-window-into-window|organize-window-group-to-window|organize-window-group-into-window|devtools|devtools-panel|emulate-device`
- `atlasd /v1/runtime/status|stage|verify|clear|install`
- `atlasd /v1/runtime/plan` 与 `/v1/runtime/plan/clear`
- `atlasd /v1/sidebar/status`、`/v1/sidebar/ask`、`/v1/sidebar/selection/ask` 与 `/v1/sidebar/summarize`
- support root 下 managed Chromium runtime 的发现、manifest、verify、install plan 状态导出
- launcher dry-run 与 session status 的 `runtime_source` 可观测性
- `atlasctl launch-webapp --dry-run`
- `atlasctl stop-webapp`
- `atlasd --once`

当前安全边界：

- `atlasd` 默认只允许回环监听地址
- 若要显式监听非回环地址，必须传 `--allow-remote-control`
- 当前未实现真实 macOS 权限探测、权限提示或授权写路径

当前阶段对齐与实施边界以 [`docs/00-INDEX.md`](docs/00-INDEX.md) 和 [`docs/reviews/CR-STAGE-ALIGNMENT-2026-04-07.md`](docs/reviews/CR-STAGE-ALIGNMENT-2026-04-07.md) 为入口。
运行手册与发布检查单见 [`atlasx/docs/RUNBOOK.md`](atlasx/docs/RUNBOOK.md)，真实 gate 入口见 [`atlasx/docs/E2E-GATE.md`](atlasx/docs/E2E-GATE.md)。
