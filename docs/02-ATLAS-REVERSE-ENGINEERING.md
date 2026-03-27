# 官方 Atlas 本地逆向结论

## 1. 安装器层

目标：

- `Install_ChatGPT_Atlas.dmg`
- 挂载后得到 `Install ChatGPT Atlas.app`

结论：

- 安装器是原生 macOS 应用，不是 Electron
- 主二进制是 `arm64 only`
- `CFBundleIdentifier = com.openai.atlas.installer`
- 版本号 `2.0`
- 构建号 `20260221030225000`

关键事实：

- 安装器内部可见真实下载地址：
  - `https://persistent.oaistatic.com/atlas/public/ChatGPT_Atlas.dmg`

## 2. 主应用层

目标：

- `ChatGPT_Atlas.dmg`
- 挂载后得到 `ChatGPT Atlas.app`

结论：

- 主应用不是 Electron `app.asar` 结构
- 主程序、关键 framework、helper 大量为 `arm64 only`
- 不能像 `Codex` 那样通过“替换 Electron x64 runtime + patch app.asar”转成 Intel 可运行版

关键元数据：

- `CFBundleIdentifier = com.openai.atlas`
- `CFBundleShortVersionString = 1.2026.63.11`
- `CFBundleVersion = 20260324211111000`
- `SUFeedURL = https://persistent.oaistatic.com/atlas/public/sparkle_public_appcast.xml`

## 3. 三层结构判断

本地证据显示 Atlas 是三层混合架构：

1. 原生宿主层
   - `ChatGPT Atlas.app`
   - 负责窗口、更新、权限、导入、系统集成
2. 内嵌浏览器层
   - `Contents/Support/ChatGPT Atlas.app`
   - 明显是 Chromium app-mode 风格封装
   - `CFBundleIdentifier = com.openai.atlas.web`
   - `CrProductDirName = com.openai.atlas.web`
   - Chromium 版本 `146.0.7680.165`
3. 业务模块层
   - 大量 `Aura*` 与 `ChatGPT*` Swift 模块

## 4. 关键模块面

从 `Aura.framework` 的 Swift 符号可确认以下业务域存在：

- `AuraAgents`
  - Agent 调度、DOM action、WebSocket、overlay、tab group、watchdog
- `AuraImport`
  - Chrome/Safari 数据导入
- `AuraBrowserMemories`
  - 浏览器记忆服务
- `AuraTabs` / `AuraTabsUI`
  - 标签页、侧边栏、搜索、自动整理
- `AuraBookmarksBar`
  - 书签栏
- `AuraSettings` / `AuraSettingsImplementation`
  - 浏览器设置、成熟内容、默认浏览器、skills、web browsing settings
- `AuraSideChat`
  - 网页侧边栏上下文问答
- `AuraWindow`
  - 应用窗口、更新、下载、设置、导航

还可确认 `ChatGPT*` 系列模块：

- `ChatGPTProjects`
- `ChatGPTConnectors`
- `ChatGPTCodex`
- `ChatGPTMemory`
- `ChatGPTMessages`
- `ChatGPTVoice`
- `ChatGPTImageGen`
- `ChatGPTSettings`
- `ChatGPTPairWithAI`

## 5. 本地运行层结论

在 Intel x64 Mac 上：

- 直接运行官方宿主壳会报 `bad CPU type in executable`
- 直接运行内嵌 `Support/ChatGPT Atlas.app` 也同样失败
- 官方原生宿主壳无法直接复用

## 6. 已验证的最低替代形态

已验证本机 `Google Chrome` 可以用独立 app 模式打开：

```text
https://chatgpt.com/atlas?get-started
```

但这只解决“进入入口页”，不能解决原生浏览器功能缺失问题。
