# ds-rescue：DeepSeek 驱动的跨家族代码 Review 工具

> ds-rescue 是一个单一二进制 CLI，使用 DeepSeek API 为 Claude Code（及其他 LLM 编码工具）提供独立的第二意见——review 方案、方案比选和提交，彻底消除回音壁。

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00ADD8.svg)](https://go.dev)
[![DeepSeek v4-pro](https://img.shields.io/badge/DeepSeek-v4--pro-blue.svg)](https://api-docs.deepseek.com/)
[![PRs welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

[English](README.md) | 中文

---

## TL;DR

- **是什么**：单二进制 CLI，支持跨家族 LLM review（plan / scheme / execution 三种模式）
- **给谁用**：Claude Code / opencode / Continue / Aider 用户，需要非 Anthropic 系的第二意见
- **为什么**：同家族 LLM 自审 = 回音壁。跨家族可多发现 1.7–2.3 倍的 bug
- **怎么用**：`go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest && ds-rescue --check`
- **成本**：DeepSeek v4-pro ~$0.01/次，v4-flash 更便宜

---

## 三种 review 模式

### plan（默认）

Review 实施方案文档：任务 DAG、架构决策、风险清单。发现缺失假设、排序 bug、回滚漏洞、未处理的失败分支。

```bash
ds-rescue docs/plans/my-feature.md           # 自动检测为 plan 模式
ds-rescue --mode plan docs/plans/my-feature.md
ds-rescue --plan docs/plans/my-feature.md    # 别名
```

### scheme

Review 方案比选表：A vs B vs C 分析、选项权衡、技术选型矩阵。对每个选项施加对抗性压力，挑战未声明的假设。

```bash
ds-rescue --mode scheme architecture-choices.md
ds-rescue --scheme architecture-choices.md
```

### execution

Review 最后一次提交或已暂存变更的 git diff。二进制内部自动执行 `git diff HEAD~1 HEAD`，无需手动传入 diff。

```bash
ds-rescue --mode execution
ds-rescue --execution
```

---

## 为什么跨家族 review 能发现更多 bug？

你写完一份方案，Claude review 了——架构清晰，没有明显漏洞。你上线了。

三天后发现 retry 逻辑里有竞态条件，Claude 当时完全没发现。为什么？因为这份方案本来就是 Claude 写的。让同一家族的模型 review 自己的产出，结构上就是**回音壁**——在相近数据分布上训练的模型共享系统性盲点。

真实案例：

```
Claude 自审一份三服务部署方案：5/5 通过，未发现问题。
DeepSeek review 同一份方案：
  - DB migration 步骤缺少幂等性保护（retry 时有数据损坏风险）
  - 回滚顺序违反依赖图（服务 B 在服务 A 之前被拆除）
  - 在受限 VM 上，health-check timeout 对冷启动来说太短
  - 第 4 步成功、第 5 步超时时会发生什么，方案完全没提
```

四个发现。Claude 一个都没发现。同一份方案。

代码审查研究持续表明：背景不同的 reviewer 发现的缺陷数量是背景相同时的 1.7–2.3 倍。同样的原理适用于 LLM。目标不是取代 Claude，而是在工作流中加入一个来自不同家族的独立第二 reviewer。

---

## 为什么选 DeepSeek？

[codex-plugin-cc](https://github.com/anthropics/codex-plugin-cc) 开创了把非 Claude reviewer 引入 Claude Code 工作流的先例。ds-rescue-cc 在这个基础上继续前行。

ds-rescue 选择 DeepSeek，是因为跨家族的价值需要一个与产出内容训练分布实质不同的模型。DeepSeek 的定价让习惯性使用成为可能——每次 review 的边际成本低到你不需要再计算，直接跑就行。

- **跨家族互补**：训练分布与 Anthropic 模型不同，盲点不重叠
- **成本高效**：低到足以让每次计划和提交都跑 review 成为习惯，而非预算决策
- **中文领域推理深度**：制造/物流/政策问题的分析质量明显更强
- **强大的编程基准性能**：足以抓住真实的 bug

---

## 快速上手

### 安装

```bash
# 全平台推荐（绕过 Windows Smart App Control）
go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest

# Linux / macOS 预构建二进制
curl -fsSL https://raw.githubusercontent.com/zred0627/ds-rescue-cc/main/scripts/install.sh | sh

# Windows PowerShell（含 SAC 检测，自动回退到 go install）
iwr https://raw.githubusercontent.com/zred0627/ds-rescue-cc/main/scripts/install.ps1 | iex
```

### 设置 API Key

```bash
export DEEPSEEK_API_KEY=sk-your-key-here
# 或持久化：mkdir -p ~/.ds-rescue && echo "sk-your-key-here" > ~/.ds-rescue/key
```

在 [platform.deepseek.com](https://platform.deepseek.com) 获取 DeepSeek API Key。

### 第一次 review

```bash
ds-rescue --check                             # 自检：key、二进制、SKILL.md、模型
ds-rescue docs/plans/my-feature.md           # plan 模式
ds-rescue --scheme architecture-choices.md   # scheme 模式
ds-rescue --execution                        # review 最后一次提交
```

---

## Claude Code 插件

```
/plugin marketplace add zred0627/ds-rescue-cc
/plugin install ds-rescue
```

安装后，直接在 Claude Code 提示符里使用：

```
/ds-rescue --plan path/to/plan.md
/ds-rescue --scheme path/to/scheme-comparison.md
/ds-rescue --execution
```

插件是对同一个 CLI 二进制的薄薄集成层——在会话中内联展示 review 结果，并在启动时检查二进制是否存在。

---

## 常见问题

### ds-rescue 和 codex-plugin-cc 有什么区别？

| 维度 | codex-plugin-cc | ds-rescue |
|---|---|---|
| 语言 / 运行时 | Node.js | Go（单二进制，无运行时依赖） |
| 模型家族 | OpenAI | DeepSeek（v4-pro / v4-flash） |
| 中文推理质量 | 标准 | 中文领域问题更强 |
| Windows 分发 | npm install | `go install`（SAC 安全）或预构建二进制 |
| 开创者认可 | 是——建立了这个模式 | 在这个基础上继续前行 |

### 为什么 Smart App Control 会拦截 Windows 下载？

Windows 11 SAC 会拦截带有 MOTW 标记的未签名二进制。我们的二进制未签名——EV 代码签名每年需要 $300–500。三种解决方式，按推荐优先级排列：

```powershell
# 1. 本地编译——无 MOTW，SAC 安全（推荐）
go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest

# 2. 解除已下载文件的拦截
Unblock-File "$env:USERPROFILE\bin\ds-rescue.exe"

# 3. 添加排除项（需要管理员权限）
Add-MpPreference -ExclusionPath "$env:USERPROFILE\bin\ds-rescue.exe"
```

### 支持哪些 DeepSeek 模型？

| 别名 | 模型 | 特点 |
|---|---|---|
| `pro`（默认） | deepseek-chat（v4-pro） | 质量最佳，约 $0.01/次 |
| `flash` | deepseek-chat-fast | 更快，约 $0.002/次 |

### 可以定制 review persona 吗？

可以——fork 或编辑 `plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md`，该文件定义三种模式的全部 review 行为。让 ds-rescue 指向你的版本：

```bash
export DS_RESCUE_SKILL_PATH=/path/to/your/SKILL.md
# 或单次调用：ds-rescue --skill /path/to/your/SKILL.md --plan my-plan.md
```

### 我的代码会发送到 DeepSeek 服务器吗？

是的，与任何 LLM API 调用一样。如果组织政策禁止将代码发送到外部 LLM API，该限制同样适用于 ds-rescue。

---

## 配置

### API Key 解析优先级

1. `$DEEPSEEK_API_KEY` 环境变量
2. `$XDG_CONFIG_HOME/ds-rescue/key`（Linux/macOS）
3. `%APPDATA%\ds-rescue\key`（Windows）
4. `~/.ds-rescue/key`（兜底）

### SKILL.md 路径解析优先级

1. `--skill <path>` CLI flag
2. `$DS_RESCUE_SKILL_PATH` 环境变量
3. `$CLAUDE_PLUGIN_ROOT/skills/ds-plan-challenger/SKILL.md`（由 Claude Code 插件自动设置）
4. `~/.ds-rescue/skills/ds-plan-challenger/SKILL.md`

### CLI 参数参考

| 参数 | 默认值 | 说明 |
|---|---|---|
| `--mode {plan\|scheme\|execution}` | 自动检测 | review 模式 |
| `--plan` / `--scheme` / `--execution` | — | 模式别名 |
| `--skill <path>` | （见上方） | SKILL.md 路径 |
| `--model {pro\|flash}` | `pro` | DeepSeek 模型别名 |
| `--api-timeout <seconds>` | `90` | API 调用超时 |
| `--exec-timeout <seconds>` | `30` | 每次工具调用超时 |
| `--no-tools` | false | 禁用 agentic 工具调用循环 |
| `--verbose` | false | 把每次工具调用记录到 stderr |
| `--version` | — | 输出版本和 git commit |
| `--check` | — | 自检：key、二进制、SKILL.md、模型响应 |

---

## 贡献

参见 [CONTRIBUTING.md](CONTRIBUTING.md)。特别欢迎 `SKILL.md` prompt 定义的 PR——无需任何 Go 知识。

---

## 许可证

MIT — 参见 [LICENSE](./LICENSE)

Copyright (c) 2026 Daniel

---

## 相关工作

- [codex-plugin-cc](https://github.com/anthropics/codex-plugin-cc) — 将跨家族 LLM review 引入 Claude Code 的先驱
- [DeepSeek API 文档](https://api-docs.deepseek.com/)
- [Smart App Control 概述](https://learn.microsoft.com/zh-cn/windows/apps/develop/smart-app-control/overview)

---

由 [Daniel (zred0627)](https://github.com/zred0627) 维护 · v0.1.0（2026-05）
