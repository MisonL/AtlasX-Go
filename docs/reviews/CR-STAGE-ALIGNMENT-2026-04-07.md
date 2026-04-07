# CR-STAGE-ALIGNMENT

- 日期: 2026-04-07
- 目标: 将 `T001-T080` 的任务级 CR 收口为项目级阶段对齐事实，作为继续迭代前的统一入口
- 结论: `tasks.csv` 与代码事实当前一致，`T001-T080` 已完成，AtlasX 已具备可验证的本地控制面、浏览器能力面、managed runtime 闭环、智能层最小闭环，以及统一 gate/runbook 入口

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
  - `atlasctl settings`
  - `atlasctl memory list`
  - `atlasctl memory search`
  - `atlasctl sidebar status`
  - `/v1/memory`
  - `/v1/memory/search`
  - `/v1/settings`
  - config/profile/support root 基础状态面
  - runtime、mirror/import、memory、sidebar 的统一状态导出
  - `atlasd` 默认仅允许回环监听
- 当前边界:
  - 当前仍是本地单机模式
  - 远程控制监听需要显式危险开关 `--allow-remote-control`
  - 没有长期后台作业编排或多节点协调

### Phase 2 Browser Capability Takeover

- 已完成:
  - `mirror-scan`
  - `import-chrome`
  - `import-safari`
  - `history|downloads|bookmarks list/open`
  - `tabs list|open|activate|close|navigate|capture|extract-context|selection|suggest|organize|devtools|emulate-device`
  - `/v1/tabs/semantic-context`
  - `/v1/tabs/selection`
  - `/v1/tabs/suggestions`
  - `/v1/tabs/organize`
  - `/v1/tabs/devtools`
  - `/v1/tabs/emulate-device`
  - `/v1/history*` `/v1/downloads*` `/v1/bookmarks*`
  - `/v1/tabs*`
  - `/v1/mirror/scan` `/v1/import/chrome` `/v1/import/safari`
- 当前边界:
  - 页面上下文已支持纯文本抓取、DOM 结构化语义提取与原生文本选区抓取，但仍不包含更深的 DOM 动作自动化
  - DevTools 当前已提供按标签页解析 frontend URL 和固定设备预设模拟的最小入口，尚未提供内置面板壳层
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
  - `atlasctl tabs memories`
  - `atlasctl tabs recommend-context`
  - `atlasctl tabs extract-context`
  - `atlasctl tabs organize`
  - `/v1/memory`
  - `/v1/memory/search`
  - `/v1/tabs/suggestions`
  - `/v1/tabs/memories`
  - `/v1/tabs/context-recommendations`
  - `/v1/tabs/semantic-context`
  - `/v1/tabs/organize`
  - capture/ask 写 memory
  - selection ask/summarize 写 memory
  - 基于 memory 的轻量检索增强
  - 按标签页聚合的 Browser memories 只读入口
  - 基于 page context + memory 的结构化页面建议
  - 基于 page context + 同 host 标签页 + memory 的结构化上下文推荐
  - 基于 DOM 的结构化上下文提取
  - 基于 page targets 的结构化标签整理建议
- 当前边界:
  - 真实 provider smoke 依赖本机 `sidebar_qa_ready=true`
  - 仍未引入向量数据库、外部检索服务或多轮代理编排
  - DOM 结构化上下文当前只覆盖 headings、links、forms 摘要；设备模拟当前只覆盖固定预设与显式清除，尚未覆盖更深层 DevTools 面板壳层

## 冻结边界

- CLI 公开面:
  - `atlasctl` 顶层命令及现有子命令名
- HTTP 公开面:
  - `/v1/status`
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
