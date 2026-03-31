# AtlasX-Go

`AtlasX-Go` 是一个面向 Intel x64 macOS 的 `Atlas-like` 重建项目。

当前仓库包含两部分：

- [`docs/`](docs/00-INDEX.md)：可行性分析、公开资料梳理、产品蓝图、架构与交接文档
- [`atlasx/`](atlasx/README.md)：Go 主导的最小控制面骨架

当前已具备的最小能力：

- `atlasctl blueprint`
- `atlasctl doctor`
- `atlasctl status`
- `atlasctl mirror-scan`
- `atlasctl history|downloads|bookmarks list/open`
- `atlasctl tabs capture <target-id>`
- `atlasd /v1/history|downloads|bookmarks` 与对应 `/open` 动作 API
- `atlasd /v1/tabs/context?id=<target-id>`
- `atlasd /v1/sidebar/status` 与 `/v1/sidebar/ask` 控制面骨架
- support root 下 managed Chromium runtime 的发现骨架与 `chrome_source` 诊断口径
- managed runtime manifest 的本地读写与状态导出骨架
- launcher dry-run 与 session status 的 `runtime_source` 可观测性
- `atlasctl runtime stage` 的本地 managed Chromium 装入链路
- `atlasctl runtime status|clear` 的 staged runtime 查看与回退链路
- `atlasctl launch-webapp --dry-run`
- `atlasctl stop-webapp`
- `atlasd --once`

项目当前结论与实施边界以 [`docs/00-INDEX.md`](docs/00-INDEX.md) 为入口。
