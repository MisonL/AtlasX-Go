# AtlasX-Go

`AtlasX-Go` 是一个面向 Intel x64 macOS 的 `Atlas-like` 重建项目。

当前仓库包含两部分：

- [`docs/`](docs/00-INDEX.md)：可行性分析、公开资料梳理、产品蓝图、架构与交接文档
- [`atlasx/`](atlasx/README.md)：Go 主导的最小控制面骨架

当前已具备的最小能力：

- `atlasctl blueprint`
- `atlasctl doctor`
- `atlasctl status`
- `atlasctl runtime stage|status|verify|clear|install`
- `atlasctl runtime plan create|status|clear`
- `atlasctl mirror-scan`
- `atlasctl history|downloads|bookmarks list/open`
- `atlasctl tabs list|open|activate|close|navigate|capture`
- `atlasctl tabs capture <target-id>`
- `atlasd /v1/history|downloads|bookmarks` 与对应 `/open` 动作 API
- `atlasd /v1/tabs|open|activate|close|navigate|context`
- `atlasd /v1/runtime/status|stage|verify|clear|install`
- `atlasd /v1/runtime/plan` 与 `/v1/runtime/plan/clear`
- `atlasd /v1/sidebar/status` 与 `/v1/sidebar/ask` 控制面骨架
- support root 下 managed Chromium runtime 的发现、manifest、verify、install plan 状态导出
- launcher dry-run 与 session status 的 `runtime_source` 可观测性
- `atlasctl launch-webapp --dry-run`
- `atlasctl stop-webapp`
- `atlasd --once`

当前阶段对齐与实施边界以 [`docs/00-INDEX.md`](docs/00-INDEX.md) 和 [`docs/reviews/CR-STAGE-ALIGNMENT-2026-04-06.md`](docs/reviews/CR-STAGE-ALIGNMENT-2026-04-06.md) 为入口。
