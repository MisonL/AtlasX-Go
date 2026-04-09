# AtlasX E2E Gate

`scripts/e2e_gate.sh` 是当前产品级 gate 入口。它把门禁拆成两层：

- 离线强制 gate：当前环境必须通过
- 本机真实 smoke：只有当前机具备 runtime、受管浏览器、真实 provider 等前置条件时才执行；否则显式标记 `UNCOVERED`

## 运行方式

```bash
cd atlasx
bash scripts/e2e_gate.sh
```

可选环境变量：

- `ATLASX_E2E_PORT`
  - atlasd 本地监听端口，默认 `17537`
- `ATLASX_E2E_ALLOW_INSTALL=1`
  - 允许脚本执行真实 `runtime install` smoke
  - 仅在本机已经准备好 install plan 且允许修改 managed runtime 时启用

## Gate Matrix

| gate 项 | 当前默认执行层 | 通过条件 | 未覆盖条件 |
| --- | --- | --- | --- |
| `go test ./...` | 离线强制 | 全仓回归通过 | 不适用 |
| `launch-webapp --dry-run` | 离线强制 | launcher plan 可生成 | 不适用 |
| `atlasd --once` | 离线强制 | bootstrap 成功 | 不适用 |
| `runtime verify` | 本机真实 smoke | 当前机已有 staged managed runtime，且 `atlasctl runtime verify` 通过 | 本机没有 staged runtime |
| `runtime install` | 本机真实 smoke | 设置 `ATLASX_E2E_ALLOW_INSTALL=1` 且 install plan 已存在，`atlasctl runtime install` 通过 | 未显式允许安装或本机没有 install plan |
| `runtime rollback` | 离线强制 | `internal/managedruntime` 的 rollback 测试通过 | 不适用 |
| `tabs capture` | 本机真实 smoke | 当前机存在受管浏览器会话且至少有一个 page target，`atlasctl tabs capture` 通过 | 无受管会话或无 page target |
| `browser-data open` | 本机真实 smoke | 当前机已有 history/bookmarks/downloads 落盘数据，且存在受管浏览器会话，至少一个 `open` 动作成功 | 当前机没有可打开的数据，或已有落盘数据但当前没有受管浏览器会话 |
| `sidebar ask` | 本机真实 smoke | `sidebar_qa_ready=true` 且存在 page target，`POST /v1/sidebar/ask` 成功 | 缺失真实 provider 凭据、配置未 ready 或无 page target |

## 设计约束

- 不把 fake provider 或测试 stub 冒充成真实 smoke 覆盖。
- 本机条件不满足时必须输出 `UNCOVERED`，而不是静默跳过。
- `runtime install` 是显式 opt-in，因为它会修改 support root/runtime。
- `runtime rollback` 当前用离线测试守住，因为真实回滚 smoke 需要故意制造安装失败，风险高于当前阶段允许范围。
