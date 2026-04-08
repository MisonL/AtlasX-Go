# CR-ISSUES-CURRENT-STATE

- 日期: 2026-04-07
- 目标: 记录当前仓库、当前开发机与当前 gate 的事实状态，供后续继续迭代直接接管

## 任务状态

- `T001-T091`
  - 当前全部已完成
- 当前任务源事实
  - `tasks.csv` 中没有剩余 `未开始` 或 `进行中` 条目

## 当前代码能力

- fallback、控制面、browser capability、managed runtime、sidebar intelligence、memory、E2E gate、runbook 已全部落地
- 当前项目已经具备：
  - 统一 CLI 与 HTTP 控制面
  - 统一 settings 只读入口：`atlasctl settings` 与 `/v1/settings`
  - 统一 memory 只读入口：`atlasctl memory list` 与 `/v1/memory`
  - 统一 memory 检索入口：`atlasctl memory search` 与 `/v1/memory/search`
  - 统一 sidebar status 入口：`atlasctl sidebar status` 与 `/v1/sidebar/status`
  - `atlasctl sidebar ask` 与 `/v1/sidebar/ask`
  - `atlasctl sidebar selection-ask` 与 `/v1/sidebar/selection/ask`
  - `atlasctl sidebar summarize` 与 `/v1/sidebar/summarize`
  - `atlasctl tabs suggest` 与 `/v1/tabs/suggestions`
  - `atlasctl tabs memories` 与 `/v1/tabs/memories`
  - `atlasctl tabs recommend-context` 与 `/v1/tabs/context-recommendations`
  - `atlasctl tabs extract-context` 与 `/v1/tabs/semantic-context`
  - `atlasctl tabs windows` 与 `/v1/tabs/windows`
  - `atlasctl tabs search` 与 `/v1/tabs/search`
  - `atlasctl tabs open-in-window` 与 `/v1/tabs/open-in-window`
  - `atlasctl tabs merge-window` 与 `/v1/tabs/merge-window`
  - `atlasctl tabs open-devtools` 与 `/v1/tabs/open-devtools`
  - `atlasctl tabs close-duplicates` 与 `/v1/tabs/close-duplicates`
  - `atlasctl tabs activate-window` 与 `/v1/tabs/activate-window`
  - `atlasctl tabs close-window` 与 `/v1/tabs/close-window`
  - `atlasctl tabs set-window-state` 与 `/v1/tabs/window-state`
  - `atlasctl tabs set-window-bounds` 与 `/v1/tabs/window-bounds`
  - `atlasctl tabs open-window` 与 `/v1/tabs/open-window`
  - `atlasctl tabs emulate-device` 与 `/v1/tabs/emulate-device`
  - `atlasctl tabs organize` 与 `/v1/tabs/organize`
  - `atlasd` 默认仅回环监听，非回环监听需显式危险开关
  - `atlasctl tabs selection|devtools` 与 `/v1/tabs/selection|devtools`
  - 标签页搜索、structured tabs capture、DOM 结构化上下文提取、原生文本选区捕获、按标签页 DevTools URL 解析、固定设备预设模拟、browser websocket 新窗口打开、窗口内打开、窗口合并、DevTools 新窗口打开、只读窗口分组、重复页清理、显式窗口激活、显式窗口关闭、显式窗口状态控制与显式窗口 bounds 控制
  - browser-data open
  - mirror/chrome import 来源目录白名单校验
  - managed runtime stage/verify/install/rollback
  - sidebar 多 provider、页内总结、DOM 结构化上下文提取、原生选区提问、本地 memory 轻量增强、按标签页聚合的 Browser memories、结构化页面建议、结构化上下文推荐、标签整理建议与固定设备预设模拟
  - 项目级 gate 与发布/恢复手册

## 当前开发机观测事实

以下事实来自 `cd atlasx && go run ./cmd/atlasd --once`：

- `ready=true`
- `chrome_status=ok`
- `chrome_source=system_auto`
- `managed_session_live=false`
- `mirror_present=true`
- `chrome_import_present=true`
- `memory_present=false`
- `runtime_manifest_present=false`
- `sidebar_qa_ready=false`

当前解释：

- 本机系统 Chrome 可发现
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

## 注意事项

- 不要把当前开发机的 `UNCOVERED` 项误判成产品缺陷
- 后续若要继续新增功能，先更新 `tasks.csv`
- 后续若代码事实再变化，应优先回写本文件和最新 stage alignment，而不是继续依赖旧日期文档
