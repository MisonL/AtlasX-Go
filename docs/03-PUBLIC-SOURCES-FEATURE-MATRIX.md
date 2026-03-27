# 官网公开资料与功能矩阵

## 主要公开来源

- 介绍页
  - <https://openai.com/index/introducing-chatgpt-atlas/>
- Release Notes
  - <https://help.openai.com/en/articles/12591856-chatgpt-atlas-release-notes>
- Ask ChatGPT Sidebar / Agent
  - <https://help.openai.com/en/articles/12628199-using-ask-chatgpt-sidebar-and-chatgpt-agent-on-atlas>
- Web browsing settings
  - <https://help.openai.com/en/articles/12625059-web-browsing-settings-on-chatgpt-atlas>
- Data controls and privacy
  - <https://help.openai.com/en/articles/12574142-chatgpt-atlas-data-controls-and-privacy>
- Enterprise
  - <https://help.openai.com/en/articles/12603091-chatgpt-atlas-for-enterprise>
- Prompt injection hardening
  - <https://openai.com/index/hardening-atlas-against-prompt-injection/>

## 公开功能面

### 浏览器壳

- 新标签页融合搜索与问答
- 地址栏与普通导航
- 标签页、分组、搜索、重命名、合并窗口、清理重复页
- 历史、下载、书签、导入
- 侧边栏与竖向标签
- 默认浏览器设置
- DevTools 与设备模拟

### ChatGPT 集成

- Ask ChatGPT 侧边栏
- Agent 模式
- Browser memories
- 项目、连接器、共享链接
- 图像、语音、文件、代码相关能力

### 安全与治理

- Logged-in / logged-out mode
- Page visibility for memories
- Data controls
- Prompt injection hardening
- Enterprise 管理与策略

## 功能复原判断

| 能力 | 公开资料可见 | 本地逆向可见 | 可复原性 |
| --- | --- | --- | --- |
| 标签页/历史/书签/下载 | 是 | 是 | 高 |
| Ask ChatGPT 侧栏 | 是 | 是 | 中高 |
| 导入 Chrome/Safari 数据 | 是 | 是 | 高 |
| 多 Profile | 是 | 是 | 高 |
| Agent 模式 | 是 | 是 | 中 |
| Browser memories | 是 | 是 | 中 |
| Prompt injection 防护 | 是 | 部分 | 低 |
| 实验开关/策略系统 | 部分 | 是 | 低 |
| 订阅、风控、内部服务 | 部分 | 部分 | 低 |

## 一句话归纳

公开资料与本地逆向是对得上的。Atlas 不是“带侧栏的网页浏览器”，而是一个完整的浏览器产品。
