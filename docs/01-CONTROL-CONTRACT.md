# AtlasX 控制合同

## Primary Setpoint

在 Intel x64 macOS 上重建一个 `Atlas-like` 桌面产品，逐步覆盖官方 `ChatGPT Atlas` 的核心浏览器能力、侧边栏问答、Agent、记忆、导入与设置能力。

## Acceptance

- L0
  - 架构、功能面、技术选型、风险边界文档齐全
- L1
  - 可启动一个独立的 `Atlas` Web 入口作为最低可用形态
  - Go 控制面可完成诊断、配置、启动与 profile 管理
- L2
  - 形成自管 Chromium runtime
  - 补齐标签页、历史、书签、下载、导入、侧边栏与部分 Agent 能力

## Guardrails

- 不依赖官方 `arm64` 宿主壳作为运行时基础
- 不把网页入口误判成产品完成态
- 主语言固定为 `Go`
- 浏览器数据面固定站在 `Chromium` 上，不用系统 WebView 作为最终壳

## Known Delays

- Chromium 运行时接管属于慢回路
- 本地原生桥接与权限系统属于慢回路
- Agent 与 browser memories 的效果验证需要真实环境和长周期观察

## Recovery Target

- 任意阶段都保留 `webapp fallback` 作为最低可运行形态
- 任意阶段失败时，回到上一阶段已验证形态，不在不稳定版本上继续叠加功能

## Constraints

- 目标平台优先 `darwin/amd64`
- 不承诺等价复刻官方闭源服务端行为
- 不承诺完整复刻官方原生外壳体验

## Risks

1. Chromium 自管 runtime 的维护成本高
2. Agent 与 browser memories 无法与官方效果完全等价
3. 默认浏览器、导入、系统权限等原生桥接是高耦合区域
