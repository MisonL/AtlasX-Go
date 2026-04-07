# AtlasX

`AtlasX` 是一个面向 Intel x64 macOS 的 Atlas-like 重建项目。

当前仓库这部分已经落下最小控制面骨架：

- `atlasctl` 诊断、蓝图展示、fallback 启动
- `atlasctl settings` 当前有效配置读取
- `atlasd` 一次性初始化与本地健康检查
- Profile、Config、Chrome runtime 探测
- 产品蓝图与阶段划分

## 命令

```bash
cd atlasx
go run ./cmd/atlasctl doctor
go run ./cmd/atlasctl blueprint
go run ./cmd/atlasctl settings
go run ./cmd/atlasctl memory list
go run ./cmd/atlasctl memory search <question>
go run ./cmd/atlasctl status
go run ./cmd/atlasctl sidebar status
go run ./cmd/atlasctl sidebar ask <target-id> <question>
go run ./cmd/atlasctl sidebar selection-ask <target-id> <question>
go run ./cmd/atlasctl sidebar summarize <target-id>
go run ./cmd/atlasctl runtime stage --bundle-path /path/to/Chromium.app --version 123.0.0
go run ./cmd/atlasctl runtime stage --bundle-path /Applications/Google\\ Chrome.app --version 136.0.7103.114
go run ./cmd/atlasctl runtime status
go run ./cmd/atlasctl runtime verify
go run ./cmd/atlasctl runtime install
go run ./cmd/atlasctl runtime clear
go run ./cmd/atlasctl runtime plan create --version 123.0.0 --channel stable --url https://example.com/chromium.zip --sha256 deadbeef --archive-path /tmp/chromium.zip --bundle-path /tmp/Chromium.app
go run ./cmd/atlasctl runtime plan resolve --catalog /path/to/runtime-catalog.json --version 123.0.0 --channel stable
go run ./cmd/atlasctl runtime plan status
go run ./cmd/atlasctl runtime plan clear
go run ./cmd/atlasctl mirror-scan
go run ./cmd/atlasctl tabs list
go run ./cmd/atlasctl tabs open https://openai.com
go run ./cmd/atlasctl tabs navigate <target-id> https://openai.com
go run ./cmd/atlasctl tabs activate <target-id>
go run ./cmd/atlasctl tabs close <target-id>
go run ./cmd/atlasctl tabs capture <target-id>
go run ./cmd/atlasctl tabs selection <target-id>
go run ./cmd/atlasctl tabs suggest <target-id>
go run ./cmd/atlasctl tabs recommend-context <target-id>
go run ./cmd/atlasctl tabs organize
go run ./cmd/atlasctl tabs devtools <target-id>
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
go run ./cmd/atlasd --listen 127.0.0.1:17537
bash scripts/e2e_gate.sh
```

完整 gate 说明见 `docs/E2E-GATE.md`。脚本会先跑离线强制 gate，再根据当前本机是否具备 managed runtime、受管浏览器和真实 provider 条件执行 smoke；条件不足时会显式输出 `UNCOVERED`。
发布与恢复手册见 `docs/RUNBOOK.md`。其中收口了 support root 布局、runtime/provider 配置、恢复步骤和发布检查单。

当前安全边界：

- `atlasd` 默认只允许回环监听地址
- 若必须监听非回环地址，需显式传 `--allow-remote-control`
- 该危险开关会扩大无鉴权控制面的暴露面，只适合受控内网或本机调试环境

## 主要 HTTP API

- `GET /healthz`
- `GET /v1/status`
- `GET /v1/settings`
- `GET /v1/history`
- `GET /v1/downloads`
- `GET /v1/bookmarks`
- `POST /v1/history/open`
- `POST /v1/downloads/open`
- `POST /v1/bookmarks/open`
- `GET /v1/runtime/status`
- `POST /v1/runtime/stage`
- `POST /v1/runtime/verify`
- `POST /v1/runtime/clear`
- `POST /v1/runtime/install`
- `GET /v1/runtime/plan`
- `POST /v1/runtime/plan`
- `POST /v1/runtime/plan/clear`
- `GET /v1/tabs`
- `GET /v1/tabs/context`
- `GET /v1/tabs/selection`
- `GET /v1/tabs/suggestions`
- `GET /v1/tabs/context-recommendations`
- `GET /v1/tabs/organize`
- `GET /v1/tabs/devtools`
- `POST /v1/tabs/open`
- `POST /v1/tabs/activate`
- `POST /v1/tabs/close`
- `POST /v1/tabs/navigate`
- `GET /v1/sidebar/status`
- `POST /v1/sidebar/ask`
- `POST /v1/sidebar/summarize`
- `POST /v1/mirror/scan`
- `POST /v1/import/chrome`
- `POST /v1/import/safari`

## 当前边界

- `launch-webapp` 只会启动 Atlas Web 入口，不等于官方原生 Atlas。
- 当前控制面只覆盖离线诊断、配置、profile 和本地健康检查。
- 当前已提供 `atlasctl settings` 与 `GET /v1/settings`，可通过统一只读视图读取当前有效配置与 sidebar provider registry。
- 受管 launcher management 当前只覆盖隔离 profile 模式；共享 profile 模式明确视为非受管。
- 当前已提供受管隔离 profile 的 CDP 入口探测，可从 `status` / `doctor` / `atlasd --once` 读取 DevTools endpoint。
- 当前 Chrome runtime 探测已区分 `system_auto` 与 `managed_auto` 来源；若 `Application Support/AtlasX/runtime/Chromium.app` 下存在可执行 bundle，诊断口径会优先识别为 managed runtime。
- 当前已提供 managed runtime manifest 骨架；`doctor` 与 `atlasd --once` 会输出 manifest 的 path/present/version/channel/bundle 状态，但这不等于 runtime 已可启动。
- 当前 `launch-webapp`、`status` 与受管 session state 已输出 `runtime_source`，可直接判断当前命中的是 `system_auto`、`managed_auto` 还是 `config`。
- 当前已提供 `runtime stage`，可把本地 `Chromium.app` 或兼容的 `Google Chrome.app` 显式复制到 support root/runtime 并写入 manifest，形成不依赖下载器的 managed runtime 装入链路。
- 当前 managed runtime 检测已优先读取 manifest 中记录的 `binary_path`，不再硬编码为 `Chromium.app/Contents/MacOS/Chromium`。
- 当前已提供 `runtime status` 和 `runtime clear`，可查看 staged runtime/manifest/binary 状态，并显式清理 support root/runtime 下的本地 managed runtime。
- 当前已提供 `runtime verify`，可离线校验 manifest 与 staged bundle/binary/sha256 一致性。
- 当前已提供 `runtime plan create|resolve|status|clear`，可离线维护 install plan 文件；`resolve` 可从本地路径或 HTTPS catalog 自动解析 channel/version/platform 对应的 url、sha256 与 bundle 元数据。
- 当前已提供 `runtime install`，会按 install plan 执行下载、archive sha256 校验、本地 stage 与最终 verify，并把 phase/error 落回 install plan 状态面。
- 当前 `runtime install` 在失败时会显式清理 `archive_path` 与 `.part` 残留；若安装前已有有效 managed runtime，则会把 install plan phase 推进到 `rolled_back` 并恢复旧 bundle/manifest。
- 当前 `atlasd` 已提供 `GET /v1/runtime/status` 与 `POST /v1/runtime/stage`，可服务化查询 managed runtime 状态并触发本地 bundle stage。
- 当前 `atlasd` 已提供 `POST /v1/runtime/verify`、`POST /v1/runtime/clear` 与 `POST /v1/runtime/install`，可服务化执行 runtime 校验、回退与 install plan 驱动安装。
- 当前 `atlasd` 已提供 `GET /v1/runtime/plan`、`POST /v1/runtime/plan` 与 `POST /v1/runtime/plan/clear`，可服务化维护 install plan 状态面。
- 当前 `/v1/status` 已额外导出 `runtime_bundle_present`、`runtime_binary_present`、`runtime_binary_executable`，避免把 manifest present 误判成 runtime ready。
- 当前受管 session 状态面已增加 stale 检测与 CDP 自愈：无进程的陈旧 session file 会被自动清理；进程仍在但 CDP 长时间未恢复时会显式标记 stale，`/v1/status` 不再把这类会话误报为 live。
- 当前 `/v1/status` 已额外导出 mirror scan、Chrome import、Safari import 最近一次执行的时间、来源、结果与错误，便于判断数据面的新鲜度与最近失败。
- 当前已提供 `mirror-scan`，会把历史/书签/下载的 source metadata 写入 `Application Support/AtlasX/mirrors/browser-data.json`。
- 当前已提供最小标签页链路：`tabs list` 读取页面级 targets，`tabs open <url>` 可通过 CDP HTTP 入口创建新标签页。
- 当前已提供标签页控制增强：`tabs activate <id>` 和 `tabs close <id>` 可操作已存在的页面级标签。
- 当前已提供 `tabs navigate <id> <url>`，通过 DevTools websocket 在现有 page target 内导航。
- 当前已提供 `tabs capture <id>`，可抓取受管 page target 的标题、URL、正文文本以及 `captured_at`、`text_length`、`text_limit`、`text_truncated`、`capture_error` 等结构化上下文字段。
- 当前已提供 `tabs selection <id>`，可在 CLI 中抓取当前 page target 的浏览器原生选区文本与长度/截断元数据。
- 当前已提供 `tabs suggest <id>`，可在 CLI 中基于当前 page context 和本地 memory retrieval 生成结构化页面建议。
- 当前已提供 `tabs recommend-context <id>`，可在 CLI 中基于当前 page context、同 host 标签页与本地 memory retrieval 生成结构化上下文推荐。
- 当前已提供 `tabs organize`，可在 CLI 中基于当前 page targets 输出结构化分组建议，作为标签自动整理的最小只读入口。
- 当前已提供 `tabs devtools <id>`，可在 CLI 中输出当前 page target 对应的 `devtools_frontend_url`。
- 当前已提供 `GET /v1/tabs/selection?id=<target-id>`，可抓取当前 page target 的浏览器原生选区文本与长度/截断元数据，用于调试和验证选区链路。
- 当前已提供 `GET /v1/tabs/suggestions?id=<target-id>`，可基于当前页上下文和本地 memory retrieval 返回结构化页面建议，不依赖真实 provider。
- 当前已提供 `GET /v1/tabs/context-recommendations?id=<target-id>`，可基于当前页上下文、同 host 标签页与本地 memory retrieval 返回结构化上下文推荐，不直接改动浏览器状态或写 memory。
- 当前已提供 `GET /v1/tabs/organize`，可基于当前 page targets 返回结构化分组建议，不直接改动浏览器状态。
- 当前已提供 `GET /v1/tabs/devtools?id=<target-id>`，可按标签页解析并返回对应的 `devtools_frontend_url`，作为最小 DevTools 入口。
- 当前已提供 Chrome 默认 profile 导入基线：`import-chrome` 会复制书签与 Preferences，并记录 History source metadata。
- 当前已提供 Safari 导入基线：`import-safari` 会导出 Safari 书签到 `Application Support/AtlasX/imports/safari/Bookmarks.json`，并记录 History.db source metadata。
- 当前已提供浏览器数据查询与动作：`history list/open`、`downloads list/open`、`bookmarks list/open` 可读取已落盘的 mirror/import 数据，并将选中 URL 打开到受管标签。
- 当前 `atlasd` 的 `/v1/status` 与 `/healthz` 已输出 launcher、mirror、import 与 sidebar QA 骨架状态，并额外提供 `/v1/history`、`/v1/downloads`、`/v1/bookmarks`、`/v1/history/open`、`/v1/downloads/open`、`/v1/bookmarks/open`、`/v1/tabs`、`/v1/tabs/context`、`/v1/sidebar/status`、`/v1/sidebar/ask` 以及 `/v1/mirror/scan`、`/v1/import/chrome`、`/v1/import/safari` 等 API。
- 当前已提供 `GET /v1/memory`，可只读返回 memory root、events file、事件摘要与最近 N 条 memory events，作为 Browser memories 的最小读控制面。
- 当前已提供 `GET /v1/memory/search`，可按 `question` 以及可选 `tab_id/title/url/limit` 查询相关 memory snippets，复用既有 retrieval 排序逻辑。
- 当前侧边栏问答已接入 `openai/openai-compatible` 与 `openrouter` 两条 provider adapter：配置使用 `default provider + provider registry`，并只记录 `api_key_env` 这类环境变量名而不写真实密钥；`/v1/sidebar/ask` 会真实抓取当前 tab context 并返回 `answer/provider/model/context_summary/trace_id`，同时支持请求级 `provider_id` 覆盖默认 provider。当前还带默认 timeout、单次 retry、token budget 护栏，并把最近错误与最近 trace 暴露到 `/v1/sidebar/status` 和 `/v1/status`。未配置时显式返回 `503`，错误 provider id 显式返回 `400`，不支持的 provider 显式返回 `501`，上游 provider 失败显式返回 `502`。
- 当前已提供 `POST /v1/sidebar/selection/ask`，可通过显式 `selection_text + question` 发起结构化选区提问，也可在未传 `selection_text` 时自动读取当前 tab 的浏览器原生选区；仍复用既有 provider、memory 与 trace/runtime state 主链，无选区时显式失败。
- 当前已提供 `POST /v1/sidebar/summarize`，可对指定 tab 生成结构化页内总结，复用既有 provider、memory 与 trace/runtime state 主链。
- 当前已提供本地 memory v1 状态面：事件文件落在 `Application Support/AtlasX/memory/events.jsonl`，格式只定义 `page_capture` 与 `qa_turn` 两类显式事件；`/v1/status` 可直接观察 memory root/file、是否存在、事件数和最近事件时间/类型。
- 当前已提供 `atlasctl sidebar status`，可在 CLI 中读取 provider readiness、runtime 护栏与最近错误/trace。
- 当前已提供 `atlasctl memory list`，可在 CLI 中读取 memory root、events file、事件摘要与最近 memory events，并支持 `--limit` 只看最近 N 条。
- 当前已提供 `atlasctl memory search <question>`，可在 CLI 中按问题检索相关 memory snippets，并支持 `--tab-id`、`--title`、`--url`、`--limit` 过滤。
- 当前已提供 `atlasctl sidebar ask <id> <question>`，可在 CLI 中对指定 tab 发起侧边栏问答，并沿用既有 provider、runtime state 与 memory 主链。
- 当前已提供 `atlasctl sidebar selection-ask <id> <question>`，可在 CLI 中对指定 tab 发起选区问答，支持 `--selection-text` 显式输入与浏览器原生选区自动抓取，并沿用既有 provider、runtime state 与 memory 主链。
- 当前已提供 `atlasctl sidebar summarize <id>`，可在 CLI 中对指定 tab 执行页内总结，并沿用既有 provider、runtime state 与 memory 主链。
- 真正的产品目标是逐步替换为自管 Chromium runtime 与 Go 控制面。
