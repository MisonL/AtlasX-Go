# AtlasX 文档索引

本目录用于承接 `ChatGPT Atlas` 在 Intel x64 macOS 上重建项目的分析、规划与交接资料。

文档顺序建议如下：

1. [01-CONTROL-CONTRACT.md](/Volumes/Work/code/AtlasX-Go/docs/01-CONTROL-CONTRACT.md)
   - 项目目标、边界、验收口径、风险控制
2. [02-ATLAS-REVERSE-ENGINEERING.md](/Volumes/Work/code/AtlasX-Go/docs/02-ATLAS-REVERSE-ENGINEERING.md)
   - 对官方安装器、主应用、内嵌浏览器壳的本地逆向结论
3. [03-PUBLIC-SOURCES-FEATURE-MATRIX.md](/Volumes/Work/code/AtlasX-Go/docs/03-PUBLIC-SOURCES-FEATURE-MATRIX.md)
   - 官网与帮助文档可确认的公开功能面
4. [04-ATLASX-PRODUCT-BLUEPRINT.md](/Volumes/Work/code/AtlasX-Go/docs/04-ATLASX-PRODUCT-BLUEPRINT.md)
   - 新生产品定义、能力分层与复原策略
5. [05-ATLASX-GO-ARCHITECTURE.md](/Volumes/Work/code/AtlasX-Go/docs/05-ATLASX-GO-ARCHITECTURE.md)
   - Go 主导的总体设计与模块边界
6. [06-IMPLEMENTATION-ROADMAP.md](/Volumes/Work/code/AtlasX-Go/docs/06-IMPLEMENTATION-ROADMAP.md)
   - 分阶段落地路线与门禁
7. [07-HANDOFF.md](/Volumes/Work/code/AtlasX-Go/docs/07-HANDOFF.md)
   - 当前状态、已验证事实、下一位接手者的起步动作
8. [08-HANDOFF-REPORT.md](/Volumes/Work/code/AtlasX-Go/docs/08-HANDOFF-REPORT.md)
   - 正式交接报告，适合直接转交下一位接手者
9. [CR-STAGE-ALIGNMENT-2026-04-07.md](/Volumes/Work/code/AtlasX-Go/docs/reviews/CR-STAGE-ALIGNMENT-2026-04-07.md)
   - 当前阶段对齐、已完成能力、冻结边界与后续开发入口
10. [CR-ISSUES-CURRENT-STATE-2026-04-07.md](/Volumes/Work/code/AtlasX-Go/docs/reviews/CR-ISSUES-CURRENT-STATE-2026-04-07.md)
   - 当前开发机与当前 gate 事实状态
11. [RELEASE-CHECKLIST-2026-04-07.md](/Volumes/Work/code/AtlasX-Go/docs/reviews/RELEASE-CHECKLIST-2026-04-07.md)
   - 发布前统一门禁与证据清单

当前一句话结论：

- 官方 `ChatGPT Atlas.app` 不能在 Intel x64 Mac 上直接运行。
- 但可以重建一个 `Atlas-like` 产品。
- 最合理的方向是 `Go + Chromium + 少量原生桥接`，而不是继续修官方 `arm64` 闭源二进制。
