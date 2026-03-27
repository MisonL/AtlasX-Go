# AtlasX 交接文档

## 当前状态

当前已完成的是“项目级分析与总体设计收口”，不是正式开发中期状态。

已沉淀内容：

- 官方 Atlas 本地逆向结论
- 官网公开资料功能面整理
- `AtlasX` 的产品蓝图
- Go 主导的总体架构
- 分阶段路线图

## 关键结论

1. 官方 `ChatGPT Atlas.app` 不能在 Intel x64 Mac 上直接运行
2. 官方 Atlas 是三层混合架构：
   - 原生宿主
   - 内嵌 Chromium 浏览器壳
   - 业务功能模块
3. 最合理的重建方向是：
   - Go 控制面
   - Chromium 数据面
   - 少量 macOS 原生桥接

## 已验证事实

- 官方主宿主与关键 framework 为 `arm64 only`
- `Support/ChatGPT Atlas.app` 为 Chromium app-mode 风格壳
- `Aura*` 与 `ChatGPT*` 模块证明官方产品能力面远超普通网页壳
- 用本机 Chrome 打开 `https://chatgpt.com/atlas?get-started` 只能作为最低 fallback

## 下一位接手者应做的第一步

1. 在新项目根目录初始化正式代码仓库
2. 把 `atlasx` 的 Go 骨架迁入新仓库主线
3. 先把 `Phase 1` 的控制面做实：
   - profile store
   - config store
   - diagnostics
   - launcher management
4. 再推进 `Phase 2`：
   - CDP
   - 历史、书签、下载镜像
   - 导入链路

## 不要做的事

- 不要继续尝试补丁官方 `arm64` 原生 Atlas
- 不要把 Wails 或系统 WebView 当最终浏览器底座
- 不要把网页 fallback 当项目完成态

## 参考来源

- 本目录全部文档
- 公开资料：
  - <https://openai.com/index/introducing-chatgpt-atlas/>
  - <https://help.openai.com/en/articles/12591856-chatgpt-atlas-release-notes>
  - <https://help.openai.com/en/articles/12628199-using-ask-chatgpt-sidebar-and-chatgpt-agent-on-atlas>
  - <https://help.openai.com/en/articles/12625059-web-browsing-settings-on-chatgpt-atlas>
  - <https://help.openai.com/en/articles/12574142-chatgpt-atlas-data-controls-and-privacy>
  - <https://help.openai.com/en/articles/12603091-chatgpt-atlas-for-enterprise>
  - <https://openai.com/index/hardening-atlas-against-prompt-injection/>
