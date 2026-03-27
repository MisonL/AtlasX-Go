# AtlasX 分阶段路线图

## Phase 0: Fallback

目标：

- 在 x64 Mac 上形成最低可运行形态

交付：

- 独立 app 模式启动
- 隔离 profile
- 共享 profile 两种入口
- 基础诊断工具

门禁：

- `atlasctl doctor`
- `atlasctl launch-webapp`

## Phase 1: Go 控制面

目标：

- 让 `AtlasX` 具备自己的控制平面

交付：

- `atlasd`
- profile store
- 配置持久化
- 日志、健康检查、诊断
- 启动治理

门禁：

- 配置与 profile 可独立治理
- Chrome runtime 探测稳定

## Phase 2: 浏览器能力接管

目标：

- 让 `AtlasX` 不再只是启动网页

交付：

- CDP 接入
- 标签管理
- 历史、书签、下载镜像
- Chrome/Safari 导入
- 侧边栏问答基础版

门禁：

- 至少 1 条真实导入链路成功
- 标签、历史、下载可观测

## Phase 3: 自管 Chromium Runtime

目标：

- 脱离外部 Chrome 依赖

交付：

- 自管 Chromium/CEF runtime
- 统一版本治理
- 内置 DevTools
- 独立壳层

门禁：

- 不依赖用户本机安装的 Chrome
- 浏览器主链稳定

## Phase 4: 智能层深化

目标：

- 向 Atlas 的智能浏览体验靠拢

交付：

- Agent 模式基础版
- Browser memories 基础版
- 页面建议、标签整理、上下文任务

门禁：

- Agent 动作链可控、可观测、可回退

## 长期风险

- Chromium 升级成本
- 本地桥接与权限回归
- Agent 安全策略成本
- 数据面和控制面耦合风险
