# AtlasX Go 主导架构

## 总体判断

主语言使用 Go 是可行的，但不能把 Go 当成“浏览器前端渲染器”。

合理分工是：

- Go: 控制面、状态面、守护进程、CLI、索引、Agent 编排、更新治理
- Chromium: 浏览器运行时
- 少量原生桥接: 权限、默认浏览器、通知、分享

## 技术栈裁决

### 选用

- Go
  - 主语言
  - `atlasd` 守护进程
  - `atlasctl` CLI
  - 本地 IPC / HTTP API
  - Profile / 配置 / 导入 / 诊断
- Chromium / CDP
  - 浏览器主数据面
  - 标签、导航、历史、下载、DOM 自动化
- Web UI
  - 侧边栏和部分控制台界面

### 不选

- Wails 作为最终壳
  - 因为它基于 OS WebView，不是浏览器产品底座
- 纯 Go GUI
  - 无法承载浏览器级能力
- Electron 作为最终主壳
  - Atlas 的能力面更接近 Chromium 浏览器产品，不是普通桌面 Web 容器

## 控制拓扑

### 控制面

- `atlasd`
  - 生命周期与配置管理
  - Agent 调度
  - 记忆索引
  - 导入管线
  - 日志与诊断
  - 更新与门禁

### 数据面

- Chromium runtime
  - 第一阶段复用外部 Chrome
  - 第二阶段切到自管 Chromium/CEF

### 状态面

- Profile Store
- Memory Index
- Import Cache
- Downloads State
- History/Bookmarks Mirror

## 目录建议

```text
atlasx/
  cmd/
    atlasctl/
    atlasd/
  internal/
    blueprint/
    platform/
      chrome/
      macos/
    runtime/
    profile/
    memory/
    imports/
    agent/
    settings/
    updater/
```

## 阶段性实现策略

### Phase 0

- Go 只做诊断和启动
- Chromium 先复用本机安装版

### Phase 1

- 加入 profile、配置、日志、健康检查
- 形成真正控制面

### Phase 2

- 用 CDP 接管浏览器动作
- 补齐标签、下载、历史、书签、导入

### Phase 3

- 引入自管 Chromium runtime
- 增加原生桥接

### Phase 4

- 深化 Agent、memories、建议系统
