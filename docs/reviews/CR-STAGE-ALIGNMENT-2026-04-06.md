# CR-STAGE-ALIGNMENT

- 日期: 2026-04-06
- 目标: 收口 AtlasX 当前阶段认知，统一 `tasks.csv`、代码事实、README 与后续开发切入点
- 结论: 当前代码事实与 `T001-T040` 一致，Phase 0-3 的控制面与状态面骨架已闭环，Phase 4 仍停留在显式占位控制面

## 当前阶段估计

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
  - config/profile/support root 基础状态面
- 当前边界:
  - 控制面主要是本地单机模式
  - 未引入长期后台作业编排和统一 release gate

### Phase 2 Browser Capability Takeover

- 已完成:
  - `mirror-scan`
  - `import-chrome`
  - `import-safari`
  - `history|downloads|bookmarks list/open`
  - `tabs list|open|activate|close|navigate|capture`
  - `/v1/history` `/v1/downloads` `/v1/bookmarks`
  - `/v1/history/open` `/v1/downloads/open` `/v1/bookmarks/open`
  - `/v1/tabs` `/v1/tabs/context` `/v1/tabs/open` `/v1/tabs/activate` `/v1/tabs/close` `/v1/tabs/navigate`
  - `/v1/mirror/scan` `/v1/import/chrome` `/v1/import/safari`
- 当前边界:
  - 页面上下文仍是最小抓取结构
  - mirror/import 还没有刷新新鲜度和最近执行状态面

### Phase 3 Managed Chromium Runtime

- 已完成:
  - `runtime stage|status|verify|clear`
  - `runtime plan create|status|clear`
  - `/v1/runtime/status`
  - `/v1/runtime/stage`
  - `/v1/runtime/verify`
  - `/v1/runtime/clear`
  - `/v1/runtime/plan`
  - `/v1/runtime/plan/clear`
  - manifest 驱动 binary 解析
  - install plan 状态机与状态面
- 当前边界:
  - 还没有真实下载执行器
  - 还没有 install/rollback 真正执行链
  - 还没有 catalog resolve

### Phase 4 Intelligence Layer

- 已完成:
  - `/v1/sidebar/status`
  - `/v1/sidebar/ask`
  - sidebar settings 骨架
- 当前边界:
  - 未配置时显式 `503`
  - 已配置但后端未实现时显式 `501`
  - 没有真实 provider adapter
  - 没有 memory 状态面

## 共享接口冻结边界

- CLI 公开面:
  - `atlasctl` 顶层命令和现有子命令名保持稳定
  - `runtime stage|status|verify|clear|plan`
  - `tabs list|open|activate|close|navigate|capture`
  - `history|downloads|bookmarks list|open`
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
  - `mirrors/browser-data.json`
  - `imports/*`

## 主要缺口

- 真实 runtime install/rollback 还不存在，Phase 3 尚未形成真正可安装闭环
- sidebar 仍无真实 provider，Phase 4 尚未进入数据面
- README 与项目入口文档此前落后于代码事实，需要同步维护
- 项目级 release gate、恢复演练和统一 checklist 还不存在

## 后续任务入口

- 当前建议按 `T041-T059` 顺序推进
- 先收口公开事实源和契约测试，再进入 runtime install 执行器与智能层
- 禁止绕过 `tasks.csv` 并行推进多个原子任务
