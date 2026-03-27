# AtlasX 交接报告

## Summary

本次工作已完成 `ChatGPT Atlas` 在 Intel x64 macOS 上的可行性分析、公开资料梳理、本地逆向归纳，以及 `AtlasX` 新生产品的总体设计收口。

当前结论明确：

- 官方 `ChatGPT Atlas.app` 不能在 Intel x64 Mac 上直接运行
- 官方产品不能按 `Codex` 的 Electron 重打包路径处理
- 可持续的方向是新建 `AtlasX`，采用 `Go + Chromium + 少量 macOS 原生桥接`

## State Estimate / Root Cause

### 官方 Atlas 的真实形态

本地逆向显示，官方 Atlas 不是单层应用，而是三层混合架构：

1. 原生宿主层
   - `ChatGPT Atlas.app`
   - 负责窗口、更新、权限、导入、系统集成
2. 内嵌浏览器层
   - `Contents/Support/ChatGPT Atlas.app`
   - Chromium app-mode 风格浏览器壳
3. 业务功能层
   - `Aura*` 与 `ChatGPT*` 模块
   - 覆盖标签页、侧边栏、Agent、导入、记忆、设置、语音、项目、连接器等能力

### 关键阻断

阻断点不在网页，而在官方宿主层和关键浏览器运行时为 `arm64 only`：

- 主宿主二进制为 `arm64 only`
- `Aura.framework` 为 `arm64 only`
- 内嵌 `ChatGPT Atlas Framework.framework` 为 `arm64 only`
- 各类 helper 也为 `arm64 only`

因此：

- 无法直接在 x64 Mac 上执行
- 无法像 `Codex` 一样通过 Electron runtime 替换和 `app.asar` 补丁完成迁移

### 产品判断

官方 Atlas 是一个完整浏览器产品，而不是普通“网页外壳 + 聊天入口”。

这意味着：

- 用普通 WebView 壳不能复原主要能力
- 用 Go 纯 GUI 也不能承载浏览器级能力
- 最合理的重建基座必须是 Chromium

## Changes

本轮已完成的资产沉淀如下：

### 1. 文档体系

已在新项目目录 [`AtlasX-Go/docs`](/Volumes/Work/code/AtlasX-Go/docs) 下建立文档集：

- [`00-INDEX.md`](/Volumes/Work/code/AtlasX-Go/docs/00-INDEX.md)
- [`01-CONTROL-CONTRACT.md`](/Volumes/Work/code/AtlasX-Go/docs/01-CONTROL-CONTRACT.md)
- [`02-ATLAS-REVERSE-ENGINEERING.md`](/Volumes/Work/code/AtlasX-Go/docs/02-ATLAS-REVERSE-ENGINEERING.md)
- [`03-PUBLIC-SOURCES-FEATURE-MATRIX.md`](/Volumes/Work/code/AtlasX-Go/docs/03-PUBLIC-SOURCES-FEATURE-MATRIX.md)
- [`04-ATLASX-PRODUCT-BLUEPRINT.md`](/Volumes/Work/code/AtlasX-Go/docs/04-ATLASX-PRODUCT-BLUEPRINT.md)
- [`05-ATLASX-GO-ARCHITECTURE.md`](/Volumes/Work/code/AtlasX-Go/docs/05-ATLASX-GO-ARCHITECTURE.md)
- [`06-IMPLEMENTATION-ROADMAP.md`](/Volumes/Work/code/AtlasX-Go/docs/06-IMPLEMENTATION-ROADMAP.md)
- [`07-HANDOFF.md`](/Volumes/Work/code/AtlasX-Go/docs/07-HANDOFF.md)

### 2. Go 主导初始骨架

已在当前仓库中落过一版 `atlasx` 骨架，用于验证 `Go` 作为主控制面的可行性，主要包括：

- `atlasctl`
- `atlasd`
- blueprint 模块
- Chrome runtime 探测与 fallback 启动器

这部分当前仍在原仓库内，尚未迁入新项目目录。

### 3. 可运行 fallback 形态

在本机 Intel x64 Mac 上，已验证以下最小运行形态成立：

- 用本机 `Google Chrome` 以 app 模式打开
  - `https://chatgpt.com/atlas?get-started`

已生成可点击启动器：

- [ChatGPT Atlas x64.app](/Users/mison/Applications/ChatGPT%20Atlas%20x64.app)
- [ChatGPT Atlas x64 (共用Chrome配置).app](/Users/mison/Applications/ChatGPT%20Atlas%20x64%20%28%E5%85%B1%E7%94%A8Chrome%E9%85%8D%E7%BD%AE%29.app)

注意：

- 这只是最低 fallback
- 它不等于官方原生 Atlas

## Verification

### 已做验证

本地逆向验证：

- DMG 挂载与结构确认
- `Info.plist`、`codesign`、`otool -L`、`file`、`lipo -info`
- Sparkle appcast 检查
- 模块符号粗分类

运行验证：

- 官方宿主壳在 x64 Mac 上直接失败
- 内嵌浏览器壳在 x64 Mac 上直接失败
- `Google Chrome --app=https://chatgpt.com/atlas?get-started` 可运行

文档验证：

- 新项目目录存在
- 文档索引和链接路径已复核

### 当前未做验证

- 尚未在 `AtlasX-Go` 目录内初始化正式 Go 工程
- 尚未把 `atlasx` 骨架迁入 `AtlasX-Go`
- 尚未做 CDP、导入链路、标签/下载/历史接管验证

## Recovery Evidence

当前恢复策略明确：

- 如果完整浏览器产品重建暂时不可推进，最低可用形态仍然保留为 Chrome app-mode fallback
- 任意阶段失败时，不继续在不稳定方案上叠功能，而是回退到上一个已验证阶段

当前可恢复锚点：

1. Web app fallback
2. 文档化的产品蓝图
3. Go 控制面初始骨架

## Observability Evidence

关键观测证据来自三类：

1. 二进制结构证据
   - `file`
   - `lipo`
   - `otool`
   - `codesign`
2. 应用元数据证据
   - `Info.plist`
   - Sparkle `SUFeedURL`
   - appcast 内容
3. 运行时证据
   - 官方壳报 `bad CPU type in executable`
   - Chrome app-mode 进程可启动并稳定驻留

这些证据足以支持当前裁决：

- 不应继续处理官方 `arm64` 二进制
- 应转向 `AtlasX` 重建路线

## Residual Risks / Gate Boundary

### 高风险区

- 自管 Chromium runtime 的维护成本
- Agent 模式的安全治理
- Browser memories 的质量和策略设计
- 默认浏览器、权限、导入等高耦合原生桥接

### 无法等价复原的部分

- 官方闭源服务端能力
- 官方实验开关体系
- 官方 prompt injection 对抗体系
- 官方账号风控与某些订阅策略

### 下一阶段门禁

下一位接手者在真正进入实现前，至少应先完成：

1. 在 [`AtlasX-Go`](/Volumes/Work/code/AtlasX-Go) 下初始化正式代码仓库
2. 迁移 `atlasx` Go 骨架
3. 做实 `Phase 1`
   - profile
   - config
   - diagnostics
   - launcher management
4. 再推进 `Phase 2`
   - CDP
   - 历史、书签、下载
   - 导入链路

## Final Handoff Statement

当前阶段已经完成“方向裁决 + 证据闭环 + 文档交接”。

项目从这里开始，不再是“研究官方 Atlas 能不能修”，而是：

- 明确转入 `AtlasX` 新生产品重建
- 以 Go 为主控制面
- 以 Chromium 为主数据面
- 以分阶段演进方式逐步逼近 Atlas 的核心能力面
