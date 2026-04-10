# CR-ISSUES-CURRENT-STATE

- 日期: 2026-04-10
- 目标: 记录当前仓库、当前开发机与当前 gate 的事实状态，供后续继续迭代直接接管

## 任务状态

- `T001-T101`
  - 当前全部已完成
- `T102`
  - 当前已完成
- `T103`
  - 当前已完成
- `T104`
  - 当前已完成
- `T105`
  - 当前已完成
- `T106`
  - 当前已完成
- `T107`
  - 当前已完成
- `T108`
  - 当前已完成
- `T109`
  - 当前已完成
- `T110`
  - 当前已完成
- `T111`
  - 当前已完成
- `T112`
  - 当前已完成
- `T113`
  - 当前已完成
- `T114`
  - 当前已完成
- `T115`
  - 当前已完成
- `T116`
  - 当前已完成
- `T117`
  - 当前已完成
- `T118`
  - 当前已完成
- `T119`
  - 当前已完成
- `T120`
  - 当前已完成
- `T121`
  - 当前已完成
- `T122`
  - 当前已完成
- `T123`
  - 当前已完成
- `T124`
  - 当前已完成
- `T125`
  - 当前已完成
- `T126`
  - 当前已完成
- `T127`
  - 当前已完成
- `T128`
  - 当前已完成
- `T129`
  - 当前已完成
- `T130`
  - 当前已完成
- `T131`
  - 当前已完成
- `T132`
  - 当前已完成
- `T133`
  - 当前已完成
- `T134`
  - 当前已完成
- `T135`
  - 当前已完成
- `T136`
  - 当前已完成
- `T137`
  - 当前已完成
- `T138`
  - 当前已完成
- `T139`
  - 当前已完成
- `T140`
  - 当前已完成
- `T141`
  - 当前已完成
- `T143`
  - 当前已完成
- 当前任务源事实
  - `tasks.csv` 中没有剩余 `未开始` 或 `进行中` 条目

## 当前代码能力

- fallback、控制面、browser capability、managed runtime、sidebar intelligence、memory、E2E gate、runbook 已全部落地
- 当前项目已经具备：
  - 统一 CLI 与 HTTP 控制面
  - 统一 settings 只读入口：`atlasctl settings` 与 `/v1/settings`
  - 统一默认浏览器只读入口：`atlasctl default-browser status` 与 `/v1/default-browser`
  - 统一默认浏览器设置入口：`atlasctl default-browser set` 与 `/v1/default-browser/set`
  - 统一日志状态只读入口：`atlasctl logs status` 与 `/v1/logs`
  - 统一更新状态只读入口：`atlasctl updates status` 与 `/v1/updates`
  - 统一结构化 doctor 诊断入口：`atlasctl doctor --json` 与 `/v1/doctor`
  - 统一 profile 状态只读入口：`atlasctl profile status` 与 `/v1/profile`
  - 统一企业策略只读入口：`atlasctl policy status` 与 `/v1/policy`
  - 统一权限边界只读入口：`atlasctl permissions status` 与 `/v1/permissions`
  - 统一 memory 只读入口：`atlasctl memory list` 与 `/v1/memory`
  - 统一 memory 检索入口：`atlasctl memory search` 与 `/v1/memory/search`
  - 统一 memory 数据控制入口：`atlasctl memory controls`、`atlasctl memory set-persist`、`atlasctl memory set-page-visibility` 与 `/v1/memory/controls`
  - 统一 memory 按站点页面可见性控制入口：`atlasctl memory set-site-visibility` 与扩展后的 `/v1/memory/controls`
  - memory site visibility 参数错误在 HTTP 控制面显式返回 `400`，不再误报为 `500`
  - 统一 sidebar status 入口：`atlasctl sidebar status` 与 `/v1/sidebar/status`
  - `atlasctl sidebar ask` 与 `/v1/sidebar/ask`
  - `atlasctl sidebar selection-ask` 与 `/v1/sidebar/selection/ask`
  - `atlasctl sidebar summarize` 与 `/v1/sidebar/summarize`
  - `atlasctl tabs suggest` 与 `/v1/tabs/suggestions`
  - `atlasctl tabs agent-plan` 与 `/v1/tabs/agent-plan`
  - `atlasctl tabs agent-execute` 与 `/v1/tabs/agent-execute`
  - `atlasctl tabs memories` 与 `/v1/tabs/memories`
  - `atlasctl tabs auth-mode` 与 `/v1/tabs/auth-mode`
  - `atlasctl tabs recommend-context` 与 `/v1/tabs/context-recommendations`
  - `atlasctl tabs extract-context` 与 `/v1/tabs/semantic-context`
  - `atlasctl tabs windows` 与 `/v1/tabs/windows`
  - `atlasctl tabs groups` 与 `/v1/tabs/groups`
  - `atlasctl tabs search` 与 `/v1/tabs/search`
  - `atlasctl tabs open-in-window` 与 `/v1/tabs/open-in-window`
  - `atlasctl tabs move-to-window` 与 `/v1/tabs/move-to-window`
  - `atlasctl tabs move-to-new-window` 与 `/v1/tabs/move-to-new-window`
  - `atlasctl tabs organize-group-to-window` 与 `/v1/tabs/organize-group-to-window`
  - `atlasctl tabs organize-group-into-window` 与 `/v1/tabs/organize-group-into-window`
  - `atlasctl tabs organize-to-windows` 与 `/v1/tabs/organize-to-windows`
  - `atlasctl tabs organize-into-window` 与 `/v1/tabs/organize-into-window`
  - `atlasctl tabs organize-window-to-windows` 与 `/v1/tabs/organize-window-to-windows`
  - `atlasctl tabs organize-window-into-window` 与 `/v1/tabs/organize-window-into-window`
  - `atlasctl tabs organize-window` 与 `/v1/tabs/organize-window`
  - `atlasctl tabs organize-window-group-to-window` 与 `/v1/tabs/organize-window-group-to-window`
  - `atlasctl tabs organize-window-group-into-window` 与 `/v1/tabs/organize-window-group-into-window`
  - `atlasctl tabs merge-window` 与 `/v1/tabs/merge-window`
  - `atlasctl tabs open-devtools` 与 `/v1/tabs/open-devtools`
  - `atlasctl tabs open-devtools-in-window` 与 `/v1/tabs/open-devtools-in-window`
  - `atlasctl tabs open-devtools-panel` 与 `/v1/tabs/open-devtools-panel`
  - `atlasctl tabs open-devtools-panel-in-window` 与 `/v1/tabs/open-devtools-panel-in-window`
  - `atlasctl tabs open-devtools-panel-window-into-window` 与 `/v1/tabs/open-devtools-panel-window-into-window`
  - `atlasctl tabs open-devtools-panel-window-to-windows` 与 `/v1/tabs/open-devtools-panel-window-to-windows`
  - `atlasctl tabs open-devtools-window-into-window` 与 `/v1/tabs/open-devtools-window-into-window`
  - `atlasctl tabs devtools-panel` 与 `/v1/tabs/devtools-panel`
  - `atlasctl tabs close-duplicates` 与 `/v1/tabs/close-duplicates`
  - `atlasctl tabs activate-window` 与 `/v1/tabs/activate-window`
  - `atlasctl tabs close-window` 与 `/v1/tabs/close-window`
  - `atlasctl tabs set-window-state` 与 `/v1/tabs/window-state`
  - `atlasctl tabs set-window-bounds` 与 `/v1/tabs/window-bounds`
  - `atlasctl tabs set-title` 与 `/v1/tabs/set-title`
  - `atlasctl tabs open-window` 与 `/v1/tabs/open-window`
  - `atlasctl tabs emulate-device` 与 `/v1/tabs/emulate-device`
  - `atlasctl tabs organize` 与 `/v1/tabs/organize`
  - `atlasd` 默认仅回环监听，非回环监听需显式危险开关
  - policy 只读视图统一汇总回环监听默认限制、危险远程控制开关、shared profile 非受管、sidebar `api_key_env` 名称与镜像/导入白名单
  - permissions 只读视图统一汇总当前未实现真实 TCC 探测、未实现权限提示、未实现授权写路径，以及 OS 权限失败显式冒泡的代码事实
  - `atlasctl tabs selection|devtools` 与 `/v1/tabs/selection|devtools`
  - 标签页搜索、structured tabs capture、DOM 结构化上下文提取、原生文本选区捕获、按标签页的 `inferred=true` 登录模式推断、按标签页 DevTools URL 解析、按指定 panel 的只读 DevTools URL 解析、固定设备预设模拟、browser websocket 新窗口打开、窗口内打开、单标签跨窗口迁移、单标签拆到新窗口、按建议组整理到新窗口、按建议组整理到指定窗口、批量按建议组整理到多窗口、批量按建议组整理到指定现有窗口、按指定窗口建议组拆到多新窗口、按指定窗口建议组整理到指定现有窗口、按指定窗口单建议组整理到新窗口、按指定窗口单建议组整理到指定现有目标窗口、窗口级只读整理建议、推导式标签页分组视图、窗口合并、标签页标题重命名、DevTools 新窗口打开、DevTools 定向打开到指定现有窗口、按指定 panel 的 DevTools 新窗口打开、按指定 panel 定向打开到指定现有窗口、只读窗口分组、重复页清理、显式窗口激活、显式窗口关闭、显式窗口状态控制与显式窗口 bounds 控制
  - browser-data open
  - mirror/chrome import 来源目录白名单校验
  - managed runtime stage/verify/install/rollback
  - sidebar 多 provider、页内总结、DOM 结构化上下文提取、原生选区提问、本地 memory 轻量增强、page-scoped memory snippets 全局可见性控制、按标签页聚合的 Browser memories、结构化页面建议、结构化上下文推荐、标签整理建议与固定设备预设模拟
  - 只读 Agent 预演计划，可基于当前页上下文、页面建议、上下文推荐与 memory 生成多步计划，不执行动作，并导出 step 级 `executable/execution_path/requires_provider` 元数据
  - 显式确认的 Agent 执行入口，当前支持单步执行与显式有界链式执行（`step_ids + max_steps`）；支持执行 sidebar 类型计划步骤、`related_tab` 单步激活与 `memory_snippet` 单步问答，不写 memory
  - 项目级 gate 与发布/恢复手册
  - 发布证据自动采集脚本 `bash scripts/release_evidence.sh`
  - 发布证据摘要自动汇总 `runtime_manifest_version`、`runtime_manifest_channel` 与 `sidebar_default_provider`
  - 发布证据摘要自动汇总 `uncovered_count` 与 `uncovered_items`
  - 发布证据摘要自动汇总 `tasks_total/tasks_done/tasks_doing/tasks_todo` 与 `release_ready/release_blockers`
  - 发布证据摘要自动汇总 `atlasd_ready/managed_session_live/sidebar_qa_ready`
  - 发布证据摘要自动汇总 `release_prerequisites`
  - 发布证据摘要自动汇总 `chrome_source/runtime_manifest_present/mirror_present`
  - 发布证据摘要自动汇总 `chrome_import_present/memory_present/logs_present/logs_file_count/updates_plan_present/updates_plan_pending`

## 当前开发机观测事实

以下事实来自 `cd atlasx && go run ./cmd/atlasd --once`：

- `ready=true`
- `chrome_status=ok`
- `chrome_source=managed_auto`
- `managed_session_live=false`
- `mirror_present=true`
- `chrome_import_present=true`
- `memory_present=false`
- `runtime_manifest_present=true`
- `runtime_manifest_version=146.0.7680.178`
- `runtime_manifest_channel=local`
- `sidebar_qa_ready=false`
- `default_browser_http_bundle_id=org.mozilla.firefox`
- `default_browser_https_bundle_id=org.mozilla.firefox`
- `default_browser_consistent=true`
- `logs_present=false`
- `logs_file_count=0`
- `updates_manifest_present=false`
- `updates_plan_present=false`
- `updates_plan_pending=false`
- `doctor_json_chrome_status=ok`
- `doctor_json_chrome_source=system_auto`
- `profile_default_profile=isolated`
- `profile_selected_mode=isolated`

以下事实来自 `cd atlasx && go run ./cmd/atlasctl policy status`:

- `policy_default_listen_addr=127.0.0.1:17537`
- `policy_loopback_only_default=true`
- `policy_remote_control_flag=--allow-remote-control`
- `policy_shared_profile_managed=false`
- `policy_sidebar_secrets_persisted=false`

以下事实来自 `cd atlasx && go run ./cmd/atlasctl permissions status`:

- `permissions_source=codebase_boundary`
- `permissions_granted_state_observable=false`
- `permissions_accessibility_probe_supported=false`
- `permissions_permission_prompt_supported=false`
- `permissions_permission_write_supported=false`
- `permissions_os_policy_failures_surface=true`

当前解释：

- 本机系统 Chrome 可发现
- 本机 LaunchServices 当前 `http/https` 默认浏览器一致，均指向 `org.mozilla.firefox`
- 本机当前 `Application Support/AtlasX/logs` 尚不存在，日志状态入口返回 `present=false`
- 本机当前已有 staged managed runtime，但仍没有 install plan，顶层更新状态入口返回 `manifest_present=true`、`plan_present=false`
- 本机 `atlasctl doctor --json` 可返回结构化诊断结果，当前 `ChromeStatus=ok`、`Chrome.Source=system_auto`
- 本机 `atlasctl profile status` 返回 `default_profile=isolated`、`selected_mode=isolated`，并确认 isolated profile 目录已存在
- 本机 `atlasctl policy status` 返回当前治理护栏视图，默认仅允许回环监听，远程控制危险开关名称为 `--allow-remote-control`，shared profile 当前显式视为非受管
- 本机 `atlasctl permissions status` 返回的是代码边界事实而不是真实 TCC 授权态；当前代码库未实现权限探测、权限提示或授权写路径
- 本机直接运行 `atlasctl tabs agent-plan <target-id>` 仍依赖受管浏览器会话；当前无 managed session 时会显式失败 `no managed browser session`
- 本机 `atlasctl tabs agent-execute --confirm [--max-steps <n>] <target-id> <step-id>...` 同样依赖受管浏览器会话；执行 `sidebar_*` 与 `memory_snippet` 步骤时还依赖 sidebar provider readiness，执行 `related_tab` 步骤不依赖 provider；多步请求超过 `max_steps` 时会显式失败；当前无 managed session 时会显式失败
- 当前没有受管浏览器会话
- mirror/import 已有历史落盘
- 当前没有 staged managed runtime
- 当前没有本地 memory 事件
- 当前没有配置好真实 provider 凭据或 provider registry

## 当前 Gate 结果

以下事实来自 `cd atlasx && bash scripts/e2e_gate.sh`：

- 离线强制 gate 通过
- 当前 `UNCOVERED` 项：
  - `runtime verify smoke`
  - `runtime install smoke`
  - `tabs capture smoke`
  - `browser-data open smoke`
  - `sidebar ask real smoke`

当前解释：

- 这些未覆盖项不是代码失败，而是本机当前缺少 staged runtime、受管浏览器会话和真实 provider readiness
- 当前 gate 已进一步区分 `browser-data open smoke` 的两类阻断：
  - 若没有已落盘 history/bookmarks/downloads 数据，会显式返回“当前都没有可打开的已落盘数据”
  - 若已有落盘数据但没有受管浏览器会话，会显式返回“已有落盘数据但当前没有受管浏览器会话”
- 当前 `bash scripts/release_evidence.sh /tmp/atlasx-release-evidence` 生成的 `SUMMARY.md` 元数据事实：
  - `runtime_manifest_version=none`
  - `runtime_manifest_channel=none`
  - `chrome_source=system_auto`
  - `sidebar_default_provider=none`
  - `runtime_manifest_present=false`
  - `mirror_present=true`
  - `tasks_total=137`
  - `tasks_done=137`
  - `tasks_doing=0`
  - `tasks_todo=0`
  - `release_ready=false`
  - `release_blockers` 当前包含 `uncovered_items_present`
  - `release_prerequisites` 当前覆盖 staged runtime、install plan、managed session、browser-data open 和 provider readiness 五类前置条件
- 若后续要做真实 smoke，需要先按 `atlasx/docs/RUNBOOK.md` 补齐对应前置条件

## 当前推荐入口

- 继续开发入口:
  - `docs/reviews/CR-STAGE-ALIGNMENT-2026-04-07.md`
- 发布与恢复入口:
  - `atlasx/docs/RUNBOOK.md`
- gate 入口:
  - `atlasx/scripts/e2e_gate.sh`
- release checklist:
  - `docs/reviews/RELEASE-CHECKLIST-2026-04-07.md`
- release evidence:
  - `atlasx/scripts/release_evidence.sh`

## 注意事项

- 不要把当前开发机的 `UNCOVERED` 项误判成产品缺陷
- 后续若要继续新增功能，先更新 `tasks.csv`
- 后续若代码事实再变化，应优先回写本文件和最新 stage alignment，而不是继续依赖旧日期文档
