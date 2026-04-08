# CR-STAGE-ALIGNMENT

- 日期: 2026-04-08
- 目标: 将 `T001-T120` 的任务级 CR 收口为项目级阶段对齐事实，作为继续迭代前的统一入口
- 结论: `tasks.csv` 与代码事实当前一致，`T001-T120` 已完成，AtlasX 已具备可验证的本地控制面、浏览器能力面、managed runtime 闭环、智能层最小闭环，以及 profile、policy、permissions、默认浏览器、日志、更新和结构化 doctor 诊断入口与统一 gate/runbook 入口

## 阶段对齐

### Phase 0 Fallback

- 已完成:
  - `atlasctl doctor`
  - `atlasctl launch-webapp --dry-run`
  - `atlasctl status`
  - `atlasctl stop-webapp`
- 当前边界:
  - 仍然是 Atlas Web 入口，不是官方原生 Atlas
  - 共享 profile 仍明确视为非受管

### Phase 1 Go Control Plane

- 已完成:
  - `atlasd --once`
  - `/healthz`
  - `/v1/status`
  - `/v1/doctor`
  - `/v1/profile`
  - `/v1/policy`
  - `/v1/permissions`
  - `atlasctl settings`
  - `atlasctl profile status`
  - `atlasctl policy status`
  - `atlasctl permissions status`
  - `atlasctl default-browser status`
  - `atlasctl logs status`
  - `atlasctl updates status`
  - `atlasctl doctor --json`
  - `atlasctl memory list`
  - `atlasctl memory search`
  - `atlasctl sidebar status`
  - `/v1/memory`
  - `/v1/memory/search`
  - `/v1/settings`
  - `/v1/default-browser`
  - `/v1/logs`
  - `/v1/updates`
  - config/profile/support root 基础状态面
  - profile 只读状态面
  - 企业策略只读状态面
  - 权限能力边界只读状态面
  - 默认浏览器 LaunchServices 只读状态面
  - logs 目录只读状态面
  - 顶层更新状态只读视图
  - 结构化 doctor 诊断视图
  - runtime、mirror/import、memory、sidebar 的统一状态导出
  - `atlasd` 默认仅允许回环监听
- 当前边界:
  - 当前仍是本地单机模式
  - 远程控制监听需要显式危险开关 `--allow-remote-control`
  - 默认浏览器当前只提供 `http/https` handler 的只读观测，不提供写操作
  - logs 当前只提供目录扫描与最近文件只读观测，不负责日志生产或轮转治理
  - updates 当前只复用 managed runtime 与 install plan 状态做只读汇总，不触发下载、安装或 catalog 解析
  - doctor JSON 当前直接复用现有 diagnostics 主链，JSON 键名保持 Go 结构体默认命名，不额外引入第二套诊断口径
  - profile 状态当前只输出默认 profile、选中模式和 isolated 目录事实，不处理 profile 迁移或共享 profile 写路径
  - policy 状态当前只汇总本地治理护栏事实，不提供策略写入、动态执行态审计或新策略文件
  - permissions 状态当前只导出代码级权限边界事实，不探测真实 TCC 授权状态、不触发权限提示、也不提供授权写路径
  - 没有长期后台作业编排或多节点协调

### Phase 2 Browser Capability Takeover

- 已完成:
  - `mirror-scan`
  - `import-chrome`
  - `import-safari`
  - `history|downloads|bookmarks list/open`
  - `tabs list|search|windows|open|open-window|open-in-window|move-to-window|move-to-new-window|merge-window|open-devtools|open-devtools-in-window|open-devtools-panel|open-devtools-panel-in-window|open-devtools-window-into-window|close-duplicates|activate-window|close-window|set-window-state|set-window-bounds|activate|close|navigate|capture|extract-context|selection|suggest|organize|organize-window|organize-group-to-window|organize-group-into-window|organize-to-windows|organize-into-window|organize-window-to-windows|organize-window-into-window|organize-window-group-to-window|organize-window-group-into-window|devtools|devtools-panel|emulate-device`
  - `/v1/tabs/search`
  - `/v1/tabs/windows`
  - `/v1/tabs/open-in-window`
  - `/v1/tabs/move-to-window`
  - `/v1/tabs/move-to-new-window`
  - `/v1/tabs/merge-window`
  - `/v1/tabs/organize-group-to-window`
  - `/v1/tabs/organize-group-into-window`
  - `/v1/tabs/organize-to-windows`
  - `/v1/tabs/organize-into-window`
  - `/v1/tabs/organize-window-to-windows`
  - `/v1/tabs/organize-window-into-window`
  - `/v1/tabs/organize-window-group-to-window`
  - `/v1/tabs/organize-window-group-into-window`
  - `/v1/tabs/open-devtools`
  - `/v1/tabs/open-devtools-in-window`
  - `/v1/tabs/open-devtools-panel`
  - `/v1/tabs/open-devtools-panel-in-window`
  - `/v1/tabs/open-devtools-window-into-window`
  - `/v1/tabs/devtools-panel`
  - `/v1/tabs/close-duplicates`
  - `/v1/tabs/activate-window`
  - `/v1/tabs/close-window`
  - `/v1/tabs/window-state`
  - `/v1/tabs/window-bounds`
  - `/v1/tabs/semantic-context`
  - `/v1/tabs/selection`
  - `/v1/tabs/suggestions`
  - `/v1/tabs/organize`
  - `/v1/tabs/organize-window`
  - `/v1/tabs/devtools`
  - `/v1/tabs/emulate-device`
  - `/v1/tabs/open-window`
  - `/v1/history*` `/v1/downloads*` `/v1/bookmarks*`
  - `/v1/tabs*`
  - `/v1/mirror/scan` `/v1/import/chrome` `/v1/import/safari`
- 当前边界:
  - 页面上下文已支持标签页搜索、纯文本抓取、DOM 结构化语义提取与原生文本选区抓取，但仍不包含更深的 DOM 动作自动化
  - DevTools 当前已提供按标签页解析 frontend URL、按指定 panel 生成只读 DevTools URL、独立新窗口打开、DevTools 定向打开到指定现有窗口、按指定 panel 的 DevTools 新窗口打开、按指定 panel 定向打开到指定现有窗口、窗口级批量 DevTools 定向打开、窗口内打开、单标签跨窗口迁移、单标签拆到新窗口、按建议组整理到新窗口、按建议组整理到指定窗口、批量按建议组整理到多窗口、批量按建议组整理到指定现有窗口、按指定窗口建议组拆到多新窗口、按指定窗口建议组整理到指定现有窗口、按指定窗口单建议组整理到新窗口、按指定窗口单建议组整理到指定现有目标窗口、窗口级只读整理建议、窗口合并、固定设备预设模拟，以及最小多窗口打开、窗口分组、重复页清理、窗口激活、窗口关闭、窗口状态控制和窗口 bounds 控制入口，尚未提供完整内置面板壳层或更深层自动化窗口编排动作
  - mirror-scan 与 import-chrome 当前只接受受信 profile 根目录，不再支持任意本地目录输入
  - browser-data open 依赖已落盘 mirror/import 数据

### Phase 3 Managed Chromium Runtime

- 已完成:
  - `runtime stage|status|verify|clear|install`
  - `runtime plan create|resolve|status|clear`
  - `/v1/runtime/status|stage|verify|clear|install`
  - `/v1/runtime/plan` 与 `/v1/runtime/plan/clear`
  - manifest/install plan/catalog/rollback/cleanup
  - 项目级 E2E gate 脚本与运行手册
- 当前边界:
  - 真实 `runtime install` live smoke 仍是显式 opt-in
  - 真实回滚 smoke 仍通过离线测试守住，不主动在本机制造失败安装

### Phase 4 Intelligence Layer

- 已完成:
  - `/v1/sidebar/status`
  - `atlasctl sidebar ask`
  - `atlasctl sidebar selection-ask`
  - `atlasctl sidebar summarize`
  - `atlasctl sidebar status`
  - `/v1/sidebar/ask`
  - `/v1/sidebar/selection/ask`
  - `/v1/sidebar/summarize`
  - `openai/openai-compatible` 与 `openrouter` provider
  - timeout/retry/token budget/trace/last error 观测
  - `memory/events.jsonl` 状态面
  - `atlasctl memory list`
  - `atlasctl memory search`
  - `atlasctl tabs suggest`
  - `atlasctl tabs agent-plan`
  - `atlasctl tabs agent-execute`
  - `atlasctl tabs memories`
  - `atlasctl tabs recommend-context`
  - `atlasctl tabs extract-context`
  - `atlasctl tabs organize`
  - `/v1/memory`
  - `/v1/memory/search`
  - `/v1/tabs/suggestions`
  - `/v1/tabs/agent-plan`
  - `/v1/tabs/agent-execute`
  - `/v1/tabs/memories`
  - `/v1/tabs/context-recommendations`
  - `/v1/tabs/semantic-context`
  - `/v1/tabs/organize`
  - capture/ask 写 memory
  - selection ask/summarize 写 memory
  - 基于 memory 的轻量检索增强
  - 按标签页聚合的 Browser memories 只读入口
  - 基于 page context + memory 的结构化页面建议
  - 基于 page context + suggestions + context recommendations + memory 的结构化 Agent 只读预演计划（含 step 级 `executable/execution_path/requires_provider` 元数据）
  - 基于 agent plan 的显式确认执行入口（支持单步与显式有界链式执行：`step_ids + max_steps`）
  - 基于 page context + 同 host 标签页 + memory 的结构化上下文推荐
  - 基于 DOM 的结构化上下文提取
  - 基于 page targets 的结构化标签整理建议
- 当前边界:
  - 真实 provider smoke 依赖本机 `sidebar_qa_ready=true`
  - 仍未引入向量数据库、外部检索服务或真正多轮代理执行器；当前 agent execute 已支持 sidebar 类型步骤、`related_tab` 单步激活与 `memory_snippet` 单步问答，并支持用户显式提供 `step_ids` + `max_steps` 的有界链式执行，未支持的步骤类型会显式拒绝
  - DOM 结构化上下文当前只覆盖 headings、links、forms 摘要；设备模拟当前只覆盖固定预设与显式清除；多窗口当前覆盖显式新开窗口、窗口内打开、单标签跨窗口迁移、单标签拆到新窗口、按建议组整理到新窗口、按建议组整理到指定窗口、批量按建议组整理到多窗口、批量按建议组整理到指定现有窗口、按指定窗口建议组拆到多新窗口、按指定窗口建议组整理到指定现有窗口、按指定窗口单建议组整理到新窗口、按指定窗口单建议组整理到指定现有目标窗口、窗口级只读整理建议、窗口合并、DevTools 新窗口打开、DevTools 定向打开到指定现有窗口、按指定 panel 的 DevTools 新窗口打开、按指定 panel 定向打开到指定现有窗口、窗口级批量 DevTools 定向打开、按指定 panel 生成只读 DevTools URL、只读窗口分组、重复页清理、显式窗口激活、显式窗口关闭、显式窗口状态控制和显式 bounds 控制，尚未覆盖完整 DevTools 面板壳层或更深层自动化窗口编排

## 冻结边界

- CLI 公开面:
  - `atlasctl` 顶层命令及现有子命令名
- HTTP 公开面:
  - `/v1/status`
  - `/v1/default-browser`
  - `/v1/logs`
  - `/v1/updates`
  - `/v1/doctor`
  - `/v1/profile`
  - `/v1/policy`
  - `/v1/permissions`
  - `/v1/runtime/*`
  - `/v1/history*`
  - `/v1/downloads*`
  - `/v1/bookmarks*`
  - `/v1/tabs*`
  - `/v1/sidebar/*`
  - `/v1/mirror/scan`
  - `/v1/import/chrome`
  - `/v1/import/safari`
- 状态面:
  - `config.json`
  - `runtime/manifest.json`
  - `runtime/install-plan.json`
  - `state/webapp-session.json`
  - `state/sidebar-qa-status.json`
  - `state/chrome-import-status.json`
  - `state/safari-import-status.json`
  - `mirrors/browser-data.json`
  - `memory/events.jsonl`
  - `imports/*`

## 项目级 Gate 入口

- 运行手册:
  - `atlasx/docs/RUNBOOK.md`
- gate 脚本:
  - `atlasx/scripts/e2e_gate.sh`
- gate 说明:
  - `atlasx/docs/E2E-GATE.md`
- 当前状态文档:
  - `docs/reviews/CR-ISSUES-CURRENT-STATE-2026-04-07.md`
- release checklist:
  - `docs/reviews/RELEASE-CHECKLIST-2026-04-07.md`

## 后续默认入口

- 继续迭代前，先读本文件
- 再读 `docs/reviews/CR-ISSUES-CURRENT-STATE-2026-04-07.md`
- 发布前，执行 `bash atlasx/scripts/e2e_gate.sh`
- 所有新任务继续以 `tasks.csv` 为唯一任务源，维持单任务闭环
