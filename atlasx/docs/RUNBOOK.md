# AtlasX Runbook

## 适用范围

本文档只覆盖当前仓库已经实现的控制面事实：

- support root 布局
- managed runtime 与 install plan 操作
- provider 配置与恢复
- 本机真实 gate 入口
- 发布前检查与恢复动作

本文档不宣称以下未实现能力：

- 自动写入真实 provider 密钥
- 真实生产发布流水线
- 自动构造回滚失败场景的 live smoke

## Support Root Layout

AtlasX 当前所有本地状态都落在：

```text
~/Library/Application Support/AtlasX/
```

当前代码已使用的目录和文件：

```text
config.json
memory/
  events.jsonl
mirrors/
  browser-data.json
imports/
  chrome/Default/
    Bookmarks.json
    Preferences.json
    report.json
  safari/
    Bookmarks.json
    report.json
runtime/
  Chromium.app
  manifest.json
  install-plan.json
state/
  webapp-session.json
  sidebar-qa-status.json
  chrome-import-status.json
  safari-import-status.json
```

各路径当前语义：

- `config.json`
  - AtlasX 主配置文件
- `memory/events.jsonl`
  - 只保存 `page_capture` 与 `qa_turn` 两类追加事件
- `mirrors/browser-data.json`
  - mirror-scan 落盘的历史/下载/书签镜像
- `imports/chrome/Default/`
  - Chrome 导入结果与 report
- `imports/safari/`
  - Safari 导入结果与 report
- `runtime/manifest.json`
  - 当前 staged managed runtime 事实
- `runtime/install-plan.json`
  - runtime install plan 事实源
- `state/webapp-session.json`
  - 受管浏览器会话状态
- `state/sidebar-qa-status.json`
  - sidebar runtime 状态、最近 trace 与最近错误
- `state/chrome-import-status.json`
  - 最近一次 Chrome 导入结果
- `state/safari-import-status.json`
  - 最近一次 Safari 导入结果

## 配置文件

配置文件路径：

```bash
cd atlasx
go run ./cmd/atlasd --once | rg '^config_file='
```

当前配置字段：

```json
{
  "chrome_binary": "",
  "default_profile": "isolated",
  "listen_addr": "127.0.0.1:17537",
  "web_app_url": "https://chatgpt.com/atlas?get-started",
  "sidebar_provider": "",
  "sidebar_model": "",
  "sidebar_base_url": "",
  "sidebar_default_provider": "primary",
  "sidebar_providers": [
    {
      "id": "primary",
      "provider": "openai",
      "model": "gpt-5.4",
      "base_url": "https://api.openai.com/v1",
      "api_key_env": "OPENAI_API_KEY"
    }
  ]
}
```

配置约束：

- 优先使用 `sidebar_default_provider + sidebar_providers`
- `api_key_env` 只保存环境变量名，不保存真实密钥
- 旧字段 `sidebar_provider/sidebar_model/sidebar_base_url` 仅保留兼容桥接

常用 provider 环境变量示例：

- `OPENAI_API_KEY`
- `OPENROUTER_API_KEY`

导出密钥后再启动 atlasd：

```bash
export OPENAI_API_KEY='your-real-key'
cd atlasx
go run ./cmd/atlasd --once
```

## Runtime 操作

检查当前 runtime 状态：

```bash
cd atlasx
go run ./cmd/atlasctl runtime status
go run ./cmd/atlasctl runtime verify
```

用本地 bundle 恢复 staged runtime：

```bash
cd atlasx
go run ./cmd/atlasctl runtime stage --bundle-path /Applications/Google\ Chrome.app --version 136.0.7103.114
go run ./cmd/atlasctl runtime verify
```

清理损坏的 staged runtime：

```bash
cd atlasx
go run ./cmd/atlasctl runtime clear
```

使用 install plan 执行安装：

```bash
cd atlasx
go run ./cmd/atlasctl runtime plan status
go run ./cmd/atlasctl runtime install
go run ./cmd/atlasctl runtime verify
```

恢复判断原则：

- `runtime verify` 失败且本机有可用本地 bundle时，优先 `runtime stage`
- staged runtime 明显损坏且无法验证时，先 `runtime clear`
- 只有 install plan 已准备好且允许修改 runtime 时才执行 `runtime install`

## 浏览器与数据面恢复

检查受管浏览器状态：

```bash
cd atlasx
go run ./cmd/atlasctl status
go run ./cmd/atlasctl tabs list
```

显式停止陈旧受管会话：

```bash
cd atlasx
go run ./cmd/atlasctl stop-webapp
```

检查启动参数但不真正启动：

```bash
cd atlasx
go run ./cmd/atlasctl launch-webapp --dry-run
```

抓当前页面上下文：

```bash
cd atlasx
go run ./cmd/atlasctl tabs capture <target-id>
```

检查镜像和导入数据：

```bash
cd atlasx
go run ./cmd/atlasctl mirror-scan
go run ./cmd/atlasctl import-chrome
go run ./cmd/atlasctl import-safari
go run ./cmd/atlasctl history list
go run ./cmd/atlasctl bookmarks list
go run ./cmd/atlasctl downloads list
```

## Sidebar 恢复

检查 sidebar readiness：

```bash
cd atlasx
go run ./cmd/atlasd --once | rg '^sidebar_qa_'
```

真实问答前必须满足：

- `sidebar_qa_ready=true`
- 当前存在受管 page target
- 对应 `api_key_env` 已在当前 shell 导出

如果 `sidebar_qa_ready=false`：

- 先检查 `config.json` 中的 `sidebar_default_provider` 和 `sidebar_providers`
- 再检查对应环境变量是否已导出
- 再检查 `base_url`、`model`、`provider` 是否完整

## 控制面监听边界

默认启动：

```bash
cd atlasx
go run ./cmd/atlasd
```

当前安全约束：

- `atlasd` 默认只允许回环监听地址，例如 `127.0.0.1:17537`、`localhost:17537`、`[::1]:17537`
- 若传入 `0.0.0.0:17537` 或其他非回环地址，进程会显式失败
- 只有显式传入 `--allow-remote-control` 时，才允许非回环监听

危险示例：

```bash
cd atlasx
go run ./cmd/atlasd --listen 0.0.0.0:17537 --allow-remote-control
```

仅在你明确接受“无鉴权控制面会被远程网络访问”的风险时才使用该模式。

## 真实 Gate 操作

当前统一 gate 入口：

```bash
cd atlasx
bash scripts/e2e_gate.sh
```

允许执行真实 runtime install smoke：

```bash
export ATLASX_E2E_ALLOW_INSTALL=1
cd atlasx
bash scripts/e2e_gate.sh
```

gate 结果解释：

- 脚本退出码非 0
  - 离线强制 gate 或已具备前置条件的 smoke 失败
- 脚本退出码为 0 且出现 `UNCOVERED`
  - 当前机缺少真实前置条件，未覆盖项需要在发布前单独补齐

## 发布检查单

- 运行 `go test ./...`
- 运行 `bash scripts/e2e_gate.sh`
- 检查 `UNCOVERED` 列表是否符合本次发布边界
- 若要覆盖真实 runtime install，先确认 install plan 已存在，再设置 `ATLASX_E2E_ALLOW_INSTALL=1`
- 若要覆盖真实 sidebar ask，先确认 `sidebar_qa_ready=true` 且当前存在受管 page target
- 记录本次使用的 runtime 版本、provider 配置 id 和 gate 结果

## 故障恢复顺序

推荐顺序：

1. `go run ./cmd/atlasd --once`
2. `go run ./cmd/atlasctl runtime status`
3. `go run ./cmd/atlasctl status`
4. `bash scripts/e2e_gate.sh`
5. 按问题类型执行 `runtime stage/clear/install`、`stop-webapp`、`import-chrome/import-safari`

优先级原则：

- 先恢复 runtime 与受管会话，再做 tabs/browser-data/sidebar
- 先暴露真实失败，再决定是否执行 `runtime clear` 或 `runtime install`
- provider 相关问题只修配置和环境变量，不把真实密钥写回仓库文件
