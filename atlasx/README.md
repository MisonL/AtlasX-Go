# AtlasX

`AtlasX` 是一个面向 Intel x64 macOS 的 Atlas-like 重建项目。

当前仓库这部分已经落下最小控制面骨架：

- `atlasctl` 诊断、蓝图展示、fallback 启动
- `atlasd` 一次性初始化与本地健康检查
- Profile、Config、Chrome runtime 探测
- 产品蓝图与阶段划分

## 命令

```bash
cd atlasx
go run ./cmd/atlasctl doctor
go run ./cmd/atlasctl blueprint
go run ./cmd/atlasctl launch-webapp --dry-run
go run ./cmd/atlasd --once
```

## 当前边界

- `launch-webapp` 只会启动 Atlas Web 入口，不等于官方原生 Atlas。
- 当前控制面只覆盖离线诊断、配置、profile 和本地健康检查。
- 真正的产品目标是逐步替换为自管 Chromium runtime 与 Go 控制面。
