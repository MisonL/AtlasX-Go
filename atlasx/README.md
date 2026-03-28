# AtlasX

`AtlasX` 是一个面向 Intel x64 macOS 的 Atlas-like 重建项目。

当前仓库这部分已经落下最小控制面骨架：

- `atlasctl` 诊断、蓝图展示、fallback 启动
- `atlasd` 一次性初始化与本地健康检查
- Profile、Config、Chrome runtime 探测
- 产品蓝图与阶段划分

## 命令

```bash
cd atlasx
go run ./cmd/atlasctl doctor
go run ./cmd/atlasctl blueprint
go run ./cmd/atlasctl status
go run ./cmd/atlasctl mirror-scan
go run ./cmd/atlasctl tabs list
go run ./cmd/atlasctl tabs open https://openai.com
go run ./cmd/atlasctl tabs activate <target-id>
go run ./cmd/atlasctl tabs close <target-id>
go run ./cmd/atlasctl import-chrome
go run ./cmd/atlasctl import-safari
go run ./cmd/atlasctl history list
go run ./cmd/atlasctl downloads list
go run ./cmd/atlasctl bookmarks list
go run ./cmd/atlasctl launch-webapp --dry-run
go run ./cmd/atlasctl stop-webapp
go run ./cmd/atlasd --once
```

## 当前边界

- `launch-webapp` 只会启动 Atlas Web 入口，不等于官方原生 Atlas。
- 当前控制面只覆盖离线诊断、配置、profile 和本地健康检查。
- 受管 launcher management 当前只覆盖隔离 profile 模式；共享 profile 模式明确视为非受管。
- 当前已提供受管隔离 profile 的 CDP 入口探测，可从 `status` / `doctor` / `atlasd --once` 读取 DevTools endpoint。
- 当前已提供 `mirror-scan`，会把历史/书签/下载的 source metadata 写入 `Application Support/AtlasX/mirrors/browser-data.json`。
- 当前已提供最小标签页链路：`tabs list` 读取页面级 targets，`tabs open <url>` 可通过 CDP HTTP 入口创建新标签页。
- 当前已提供标签页控制增强：`tabs activate <id>` 和 `tabs close <id>` 可操作已存在的页面级标签。
- 当前已提供 Chrome 默认 profile 导入基线：`import-chrome` 会复制书签与 Preferences，并记录 History source metadata。
- 当前已提供 Safari 导入基线：`import-safari` 会导出 Safari 书签到 `Application Support/AtlasX/imports/safari/Bookmarks.json`，并记录 History.db source metadata。
- 当前已提供浏览器数据查询：`history list`、`downloads list`、`bookmarks list` 可读取已落盘的 mirror/import 数据。
- 当前 `atlasd` 的 `/v1/status` 与 `/healthz` 已输出 launcher、mirror、import 的统一状态，并额外提供 `/v1/history`、`/v1/downloads`、`/v1/bookmarks`。
- 真正的产品目标是逐步替换为自管 Chromium runtime 与 Go 控制面。
