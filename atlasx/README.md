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
go run ./cmd/atlasctl runtime stage --bundle-path /path/to/Chromium.app --version 123.0.0
go run ./cmd/atlasctl runtime stage --bundle-path /Applications/Google\\ Chrome.app --version 136.0.7103.114
go run ./cmd/atlasctl runtime status
go run ./cmd/atlasctl runtime clear
go run ./cmd/atlasctl mirror-scan
go run ./cmd/atlasctl tabs list
go run ./cmd/atlasctl tabs open https://openai.com
go run ./cmd/atlasctl tabs activate <target-id>
go run ./cmd/atlasctl tabs close <target-id>
go run ./cmd/atlasctl tabs capture <target-id>
go run ./cmd/atlasctl import-chrome
go run ./cmd/atlasctl import-safari
go run ./cmd/atlasctl history list
go run ./cmd/atlasctl history open <index>
go run ./cmd/atlasctl downloads list
go run ./cmd/atlasctl downloads open <index>
go run ./cmd/atlasctl bookmarks list
go run ./cmd/atlasctl bookmarks open <index>
go run ./cmd/atlasctl launch-webapp --dry-run
go run ./cmd/atlasctl stop-webapp
go run ./cmd/atlasd --once
```

## 当前边界

- `launch-webapp` 只会启动 Atlas Web 入口，不等于官方原生 Atlas。
- 当前控制面只覆盖离线诊断、配置、profile 和本地健康检查。
- 受管 launcher management 当前只覆盖隔离 profile 模式；共享 profile 模式明确视为非受管。
- 当前已提供受管隔离 profile 的 CDP 入口探测，可从 `status` / `doctor` / `atlasd --once` 读取 DevTools endpoint。
- 当前 Chrome runtime 探测已区分 `system_auto` 与 `managed_auto` 来源；若 `Application Support/AtlasX/runtime/Chromium.app` 下存在可执行 bundle，诊断口径会优先识别为 managed runtime。
- 当前已提供 managed runtime manifest 骨架；`doctor` 与 `atlasd --once` 会输出 manifest 的 path/present/version/channel/bundle 状态，但这不等于 runtime 已可启动。
- 当前 `launch-webapp`、`status` 与受管 session state 已输出 `runtime_source`，可直接判断当前命中的是 `system_auto`、`managed_auto` 还是 `config`。
- 当前已提供 `runtime stage`，可把本地 `Chromium.app` 或兼容的 `Google Chrome.app` 显式复制到 support root/runtime 并写入 manifest，形成不依赖下载器的 managed runtime 装入链路。
- 当前 managed runtime 检测已优先读取 manifest 中记录的 `binary_path`，不再硬编码为 `Chromium.app/Contents/MacOS/Chromium`。
- 当前已提供 `runtime status` 和 `runtime clear`，可查看 staged runtime/manifest/binary 状态，并显式清理 support root/runtime 下的本地 managed runtime。
- 当前 `atlasd` 已提供 `GET /v1/runtime/status` 与 `POST /v1/runtime/stage`，可服务化查询 managed runtime 状态并触发本地 bundle stage。
- 当前 `atlasd` 已提供 `POST /v1/runtime/clear`，可服务化回退 staged runtime；clear 后 `GET /v1/runtime/status` 会回到 manifest/bundle/binary 全部缺失状态。
- 当前 `/v1/status` 已额外导出 `runtime_bundle_present`、`runtime_binary_present`、`runtime_binary_executable`，避免把 manifest present 误判成 runtime ready。
- 当前已提供 `mirror-scan`，会把历史/书签/下载的 source metadata 写入 `Application Support/AtlasX/mirrors/browser-data.json`。
- 当前已提供最小标签页链路：`tabs list` 读取页面级 targets，`tabs open <url>` 可通过 CDP HTTP 入口创建新标签页。
- 当前已提供标签页控制增强：`tabs activate <id>` 和 `tabs close <id>` 可操作已存在的页面级标签。
- 当前已提供 `tabs navigate <id> <url>`，通过 DevTools websocket 在现有 page target 内导航。
- 当前已提供 `tabs capture <id>`，可抓取受管 page target 的标题、URL 和最小正文文本。
- 当前已提供 Chrome 默认 profile 导入基线：`import-chrome` 会复制书签与 Preferences，并记录 History source metadata。
- 当前已提供 Safari 导入基线：`import-safari` 会导出 Safari 书签到 `Application Support/AtlasX/imports/safari/Bookmarks.json`，并记录 History.db source metadata。
- 当前已提供浏览器数据查询与动作：`history list/open`、`downloads list/open`、`bookmarks list/open` 可读取已落盘的 mirror/import 数据，并将选中 URL 打开到受管标签。
- 当前 `atlasd` 的 `/v1/status` 与 `/healthz` 已输出 launcher、mirror、import 与 sidebar QA 骨架状态，并额外提供 `/v1/history`、`/v1/downloads`、`/v1/bookmarks`、`/v1/history/open`、`/v1/downloads/open`、`/v1/bookmarks/open`、`/v1/tabs`、`/v1/tabs/context`、`/v1/sidebar/status`、`/v1/sidebar/ask` 以及 `/v1/mirror/scan`、`/v1/import/chrome`、`/v1/import/safari` 等 API。
- 当前侧边栏问答只提供控制面骨架：未配置时 `/v1/sidebar/ask` 显式返回 `503`，即使已配置 provider/model/base URL 但后端未实现，也会显式返回 `501`。
- 真正的产品目标是逐步替换为自管 Chromium runtime 与 Go 控制面。
