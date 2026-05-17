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
- **给谁用**：使用 Claude Code / opencode / Continue / Aider 的开发者，希望获得非 Anthropic 的第二意见
- **为什么**：同家族 LLM 自我 review = 回音壁。跨家族可多发现 1.7–2.3 倍的 bug。
- **怎么用**：`go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest && ds-rescue --check`
- **成本**：DeepSeek v4-pro ~$0.01/次，v4-flash 更便宜

---

## ds-rescue 是什么？

ds-rescue 是一个独立 Go 二进制，支持三种 review 模式——plan、scheme、execution——每种模式均由 DeepSeek v4-pro 模型家族驱动。它不是聊天助手，不生成代码，也不回答问题。它是单一用途的对抗性 reviewer：给它一份文档或 diff，从一个没有参与生成被 review 工件的模型那里获得结构化发现。

这个工具被设计为习惯性工具。每次 review 调用约 $0.01，没有任何实质性的成本门槛阻止你在每份方案、每个架构决策、每次提交上都跑一遍。review 行为定义在一个纯 Markdown 文件（`SKILL.md`）中，你可以读取、fork 并定制，完全不需要动任何 Go 代码。

ds-rescue 在 Linux、macOS 和 Windows 上均可作为独立 CLI 使用，也可选择性地安装为 Claude Code 插件，在编码会话中内联展示 review 结果。

---

## 我应该在什么时候用 ds-rescue？

### 适合使用的场景

- 你写完一份方案，在开始实施前想要第二意见
- 你正在多个架构方案之间选择，想对每个选项施加对抗性压力
- 你完成了一个工作单元，想在 push 前 review diff
- 你使用 Claude Code，想在工作流中加入非 Anthropic 的 reviewer
- 你的工作涉及中文，或处理中文技术文献相关问题（物流、制造、政策分析）
- 你想要一个与生成工件的模型具有不同训练先验的 reviewer

### 不适合的场景

- 你需要聊天助手或代码生成器——直接用 Claude / Gemini
- 你的组织政策禁止将代码发送到外部 LLM API（这是所有 LLM 工具的通用约束，不是 ds-rescue 特有的）
- 你需要本地/离线 review——ds-rescue 始终调用 DeepSeek API
- 你希望 review 来自与作者相同的模型家族（设计上永远不会这样）

---

## 三种 review 模式

### 模式：plan（默认）

Review 实施方案文档：任务 DAG、架构决策、风险清单、迭代计划。发现缺失假设、排序 bug、回滚漏洞和未处理的失败分支。

```bash
ds-rescue docs/plans/my-feature.md          # 自动检测为 plan 模式
ds-rescue --mode plan docs/plans/my-feature.md
ds-rescue --plan docs/plans/my-feature.md   # 别名
```

在开始功能开发或迭代之前使用。

### 模式：scheme

Review 方案比选表格：A vs B vs C 分析、选项权衡、技术选型矩阵。对每个选项施加对抗性压力，标记缺失的评估标准，挑战未声明的假设。

```bash
ds-rescue --mode scheme architecture-choices.md
ds-rescue --scheme architecture-choices.md  # 别名
```

在多个架构备选方案之间做选择时使用。

### 模式：execution

Review 最后一次提交或已暂存变更的 git diff。二进制内部自动执行 `git diff HEAD~1 HEAD`，无需手动传入 diff。

```bash
ds-rescue --mode execution
ds-rescue --execution                        # 别名
```

完成一个工作单元、push 之前使用。

---

## 跨家族 review 为什么能发现更多 bug？

你写完一份方案，交给 Claude review。Claude 说看起来很好——架构清晰、思路合理、没有明显漏洞。你照着去做了。

三天后发现 retry 逻辑里有竞态条件，Claude 当时没有提。为什么？因为这份方案本来就是 Claude 帮你写的。让同一家族的模型 review 自己的输出，结构上就是一个**回音壁**。

这不是批评 Claude，而是 LLM 工作方式的固有属性：在相近数据分布上训练的模型，共享系统性盲点。对一份隐含边界条件 bug 的方案信心满满的模型，在重新阅读时同样会以高置信度忽略那个 bug。失败模式是相关的，不是独立的。

真实 dogfooding 案例：

```
Claude 自我 review 一份三服务部署方案：5/5 通过，未发现问题。
DeepSeek review 同一份方案：
  - DB migration 步骤缺少幂等性保护（retry 时有数据损坏风险）
  - 回滚顺序违反依赖图（服务 B 在服务 A 之前被拆除）
  - 在资源受限 VM 上，health-check timeout 对冷启动来说太短
  - 第 4 步成功、第 5 步超时的情况下会发生什么，方案完全没提
```

四个发现。Claude 一个都没发现。同一份方案。

### 工程价值

软件工程里有一个被充分研究的类比：**代码 review 效果随 reviewer 多样性而提升**。关于结对 review 和审查程序的研究持续表明，当 reviewer 背景不同、先验知识不同时，发现的缺陷数量是同质 reviewer 的 1.7–2.3 倍。

同样的原则适用于 LLM review。来自不同训练家族的模型对"正常"有不同的偏见，对哪些边界条件值得提及有不同的先验，有不同的失败模式，与作者的失败模式不容易重叠。DeepSeek v4-pro 在数据分布和架构选择上与 Claude 有实质性差异。这种互补性就是这里要释放的工程价值。

目标不是替代 Claude，而是在工作流中加入一个来自不同家族的独立第二 reviewer——就像你不会让 PR 的作者同时成为它唯一的 reviewer 一样。

---

## 为什么选 DeepSeek？

[codex-plugin-cc](https://github.com/anthropics/codex-plugin-cc) 开创了把非 Claude reviewer 引入 Claude Code 工作流的先例，理应得到完整的认可。ds-rescue-cc 的存在因那项工作而成为可能。

ds-rescue 选择 DeepSeek，是因为跨家族的价值需要一个与产出内容训练分布实质不同的模型。DeepSeek 的定价让习惯性使用成为可能——每次审查的边际成本低到你不需要再计算，直接跑就行。

### 为什么具体是 DeepSeek

- **跨家族互补性**：在与 Anthropic 模型不同的数据分布上训练，盲点不重叠
- **成本高效**：低到足以让每次计划和提交都跑审查成为习惯，而非预算决策
- **中文领域推理深度**：对于使用中文工作、或处理制造/物流/政策问题的团队，DeepSeek 的中文推理质量明显更强
- **强大的编程基准性能**：足以抓住真实的 bug，如上面的案例所示

---

## 快速上手（60 秒）

### 全平台推荐：通过 Go 安装

```bash
go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest
```

这是首选方式的原因：

- 在 Linux、macOS、Windows 上完全一致
- **绕过 Windows Smart App Control**——本地编译的二进制，没有 Mark-of-the-Web 标记
- 始终从你固定的 Go 工具链构建
- 需要 Go 1.22+（从 [go.dev/dl](https://go.dev/dl/) 安装）

### Linux / macOS：curl 一键安装预构建二进制

```bash
curl -fsSL https://raw.githubusercontent.com/zred0627/ds-rescue-cc/main/scripts/install.sh | sh
```

### Windows：PowerShell 安装脚本（含 SAC 检测）

```powershell
iwr https://raw.githubusercontent.com/zred0627/ds-rescue-cc/main/scripts/install.ps1 | iex
```

> **注意**：如果 Windows Smart App Control 处于开启状态，安装脚本会自动检测并提供 `go install` 的替代指引。详见[技术深度解析](#为什么-smart-app-control-会拦截我们的二进制技术深度解析)章节。

### 设置 API Key

```bash
# 方式 A：环境变量（推荐用于 CI）
export DEEPSEEK_API_KEY=sk-你的-key

# 方式 B：key 文件（重启 shell 后依然有效）
mkdir -p ~/.ds-rescue && echo "sk-你的-key" > ~/.ds-rescue/key && chmod 600 ~/.ds-rescue/key
```

在 [platform.deepseek.com](https://platform.deepseek.com) 获取 DeepSeek API Key。

### 第一次 review

```bash
ds-rescue --check                              # 自检：key、二进制、SKILL.md、模型响应
ds-rescue docs/plans/my-feature.md            # plan 模式（自动检测）
ds-rescue --scheme architecture-choices.md    # scheme 模式
ds-rescue --execution                         # review 最后一次提交
```

---

## Claude Code 插件安装

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

插件注册了一个 `SessionStart` hook，在启动时检查 ds-rescue 二进制是否存在，如缺失则优先推荐 `go install`。它还会自动设置 `$CLAUDE_PLUGIN_ROOT`，让二进制能找到 `SKILL.md`，并将 review 结果内联展示在对话中。

底层二进制与独立 CLI 完全相同——插件是薄薄的集成层，不是单独的产品。

---

## 工作原理

```
  你的方案 / 方案比选 / git diff
          │
          ▼
    ds-rescue CLI
    ─────────────────────────────────────────────────────
    │  1. 加载 SKILL.md（定义 review persona 和输出 schema）
    │  2. 根据所选模式组合 prompt
    │  3. 调用 DeepSeek API（默认 v4-pro）
    │  4. 可选 agentic 工具调用循环（read_file / write_file / bash_exec）
    │     最多 20 轮，拒绝模式列表保护
    │  5. 将结构化发现输出到 stdout
    ─────────────────────────────────────────────────────
          │
          ▼
    对抗性发现（stdout）
    诊断信息（stderr）
```

三种模式的 review 行为完全定义在 `plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md` 中。该文件在运行时从磁盘加载，不嵌入二进制。你可以在不动 Go 代码的情况下定制它。

---

## 常见问题（FAQ）

### ds-rescue 和 codex-plugin-cc 有什么区别？

| 维度 | codex-plugin-cc | ds-rescue |
|---|---|---|
| 语言 / 运行时 | Node.js | Go（单二进制，无运行时依赖） |
| 模型家族 | OpenAI | DeepSeek（v4-pro / v4-flash） |
| 中文推理质量 | 标准 | 在中文领域问题上更强 |
| Windows 分发 | npm install | `go install`（SAC 安全）或预构建二进制 |
| 开创者认可 | 是——建立了这个模式 | 在这个基础上继续前行 |

两个工具都属于这个生态系统，互为补充，不是竞争关系。

### ds-rescue 支持 OpenAI / Claude / Gemini 吗？

不支持，这是设计决策。跨家族 review 的价值，正来自于使用与生成工件的模型**不同**的模型家族。如果你想对 Claude 撰写的方案做 DeepSeek review，ds-rescue 做到了。

### 为什么 Smart App Control 会拦截 Windows 下载的二进制？

Windows 11（25H2 及以后版本）默认开启 Smart App Control（SAC）。SAC 会拦截通过互联网下载的、带有 Mark of the Web（MOTW）区域标记的未签名二进制。

我们的预构建二进制没有代码签名。EV（Extended Validation）代码签名证书每年需要 $300–500，还需要组织验证——对一个由单个开发者维护的免费开源项目来说不可行。

三种解决方式，按推荐优先级排列：

1. **推荐——本地编译**（无 MOTW 标记，SAC 默认信任）：
   ```powershell
   go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest
   ```

2. **解除下载文件的锁定**（右键 → 属性 → 取消封锁，或 PowerShell）：
   ```powershell
   Unblock-File "$env:USERPROFILE\bin\ds-rescue.exe"
   ```

3. **添加排除项**（需要管理员权限）：
   ```powershell
   Add-MpPreference -ExclusionPath "$env:USERPROFILE\bin\ds-rescue.exe"
   ```

install.ps1 脚本会自动检测你的 SAC 状态，并打印相应的操作指引。

### 日常使用成本是多少？

v4-pro ~$0.01/次——每天高频使用，每月也只是几美元。v4-flash（~$0.002/次）适合高频迭代场景，费用更低。

### 我可以定制 review persona 吗？

可以。Fork 或编辑 `SKILL.md`，路径为 `plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md`。该文件有三个 section（`## Plan Mode`、`## Scheme Mode`、`## Execution Mode`），每个包含对应 review 类型的系统 prompt 和输出 schema。

让 ds-rescue 指向你的定制版本：

```bash
export DS_RESCUE_SKILL_PATH=/path/to/your/SKILL.md
```

或每次调用时传入：

```bash
ds-rescue --skill /path/to/your/SKILL.md --plan my-plan.md
```

定制方向举例：金融/投资备忘录审计、安全审计视角、基础设施即代码 review、中文推理深度增强模式。

### 我的代码会发送到 DeepSeek 的服务器吗？

是的，就像任何 LLM API 调用一样。方案文档、方案比选或 git diff 会发送到 DeepSeek 的 API 服务器进行处理。这与使用 Codex、Claude API 或 Gemini API 时发生的数据共享完全相同。如果你的组织政策禁止将专有代码发送到外部 LLM API，该约束适用于 ds-rescue，也同样适用于所有 LLM 工具。

### 需要联网吗？

是的。ds-rescue 始终调用 DeepSeek API，没有离线或本地模型模式。

### 支持哪些 DeepSeek 模型？

| 别名 | 模型 | 特点 |
|---|---|---|
| `pro`（默认） | deepseek-chat（v4-pro） | 质量最佳，约 $0.01/次 |
| `flash` | deepseek-chat-fast | 更快，约 $0.002/次 |

使用 `--model pro` 或 `--model flash` 选择。

### 不用 Claude Code 也能用吗？

可以——ds-rescue 是通用 CLI。其他工具使用示例：

```bash
# opencode
ds-rescue --mode plan docs/plans/current-plan.md

# Aider
ds-rescue --mode execution && aider --message "修复 ds-rescue 发现的问题"

# 裸 shell / pre-push hook
ds-rescue --mode execution --no-tools || echo "Review 发现了问题，请检查 stderr"
```

### 怎么更新？

```bash
# 重新执行安装命令即可拉取最新版本
go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest

# 如果使用 Claude Code 插件
/plugin update ds-rescue
```

---

## 通用 CLI 用法（不依赖 Claude Code）

ds-rescue 是标准 CLI 二进制，可以在任何能执行 shell 命令的环境中工作。

**opencode：**

```bash
ds-rescue --mode plan docs/plans/current-plan.md
```

**Continue（VS Code / JetBrains）——添加自定义命令：**

```json
{
  "name": "ds-rescue review",
  "command": "ds-rescue --mode execution",
  "description": "对当前变更运行 DeepSeek 对抗性 review"
}
```

**Aider：**

```bash
ds-rescue --mode execution && aider --message "修复 ds-rescue 发现的问题"
```

**裸 shell / CI pre-push hook：**

```bash
ds-rescue --mode execution --no-tools || echo "Review 发现了问题，请检查 stderr"
```

二进制从 `stdin` 或文件参数读取，把发现写到 `stdout`，把诊断信息写到 `stderr`，可以与任何工具链组合使用。

---

## 定制 review persona

三种模式的 review 行为定义在单一 Markdown 文件中：

```
plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md
```

这个文件有三个 section——`## Plan Mode`、`## Scheme Mode`、`## Execution Mode`——每个包含对应 review 类型的系统 prompt 和输出 schema。二进制在运行时从磁盘加载这个文件，它不嵌入在二进制里。

```bash
# fork 仓库，按你的需求编辑 SKILL.md
vim plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md

# 让 ds-rescue 指向你的定制版本
export DS_RESCUE_SKILL_PATH=/path/to/your/SKILL.md
ds-rescue --mode plan my-plan.md

# 或者每次调用时传入
ds-rescue --skill /path/to/your/SKILL.md --mode plan my-plan.md
```

定制方向举例：

- **金融/投资备忘录 review**：标记缺失假设、检查 DCF 输入、挑战收入预测
- **安全审计视角**：标记注入面、认证边界、密钥处理
- **基础设施即代码 review**：标记缺失回滚路径、爆炸半径、依赖排序
- **中文推理深度增强**：为制造、物流、政策等中文背景场景提供更深入的分析

一份改进 `SKILL.md` 的 PR 与改进 Go 代码的 PR 同等有价值，无需任何 Go 知识。

---

## 配置

### API Key 解析优先级

1. `$DEEPSEEK_API_KEY` 环境变量（最高优先级）
2. `$XDG_CONFIG_HOME/ds-rescue/key`（Linux/macOS XDG 标准）
3. `%APPDATA%\ds-rescue\key`（Windows）
4. `~/.ds-rescue/key`（所有平台的 fallback）

### SKILL.md 路径解析优先级

1. `--skill <path>` CLI flag（最高优先级）
2. `$DS_RESCUE_SKILL_PATH` 环境变量
3. `$CLAUDE_PLUGIN_ROOT/skills/ds-plan-challenger/SKILL.md`（由 Claude Code 插件自动设置）
4. `~/.ds-rescue/skills/ds-plan-challenger/SKILL.md`（独立安装）
5. 如果都找不到：二进制退出并列出所有尝试过的路径

### CLI 参数参考

| 参数 | 默认值 | 说明 |
|---|---|---|
| `--mode {plan\|scheme\|execution}` | 自动检测 | review 模式 |
| `--plan` | — | `--mode plan` 的别名 |
| `--scheme` | — | `--mode scheme` 的别名 |
| `--execution` | — | `--mode execution` 的别名 |
| `--skill <path>` | （见上方） | SKILL.md 路径 |
| `--model {pro\|flash}` | `pro` | DeepSeek 模型别名 |
| `--api-timeout <seconds>` | `90` | API 调用超时 |
| `--exec-timeout <seconds>` | `30` | 每次工具调用超时 |
| `--no-tools` | false | 禁用 agentic 工具调用循环 |
| `--verbose` | false | 把每次工具调用记录到 stderr |
| `--version` | — | 输出版本和 git commit |
| `--check` | — | 自检：key、二进制、SKILL.md、模型响应 |

---

## 为什么 Smart App Control 会拦截我们的二进制（技术深度解析）

Windows 11 25H2 引入了 Smart App Control（SAC）作为默认开启的安全功能。SAC 在每个可执行文件运行前对其进行评估。对于无法通过 Microsoft 信誉服务或受信任代码签名证书验证的可执行文件，SAC 会检查 Mark of the Web（MOTW）区域标记。

任何从互联网下载的文件——通过浏览器、`curl`、`Invoke-WebRequest` 或 GitHub Releases 下载——都会在 NTFS 备用数据流中自动获得 `ZoneId=3`（互联网区域）标记。SAC 会拦截这类文件的执行，除非：

- 文件使用 Microsoft 信任的 EV 证书签名，或
- 文件已建立 Microsoft 信誉评分（通常需要数百万次下载），或
- MOTW 标记被移除（通过 Unblock-File 或属性对话框）。

我们的二进制未签名，因为 EV 代码签名证书每年需要 $300–500 并需要组织验证——对一个免费开源工具来说不可行。

**最简洁的解决方案是 `go install`**：当 Go 在本地编译二进制时，生成的可执行文件被写入本地 `$GOPATH/bin`，不带任何 MOTW 标记。SAC 没有理由拦截它。这就是为什么 `go install` 是 Windows 用户的推荐安装方式。

三种完整解决方案：

```powershell
# 1. 本地编译——无 MOTW，SAC 安全（推荐）
go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest

# 2. 解除下载文件锁定（从特定文件移除 MOTW 标记）
Unblock-File "$env:USERPROFILE\bin\ds-rescue.exe"

# 3. 添加排除项（需要管理员权限，影响 Defender + SAC）
Add-MpPreference -ExclusionPath "$env:USERPROFILE\bin\ds-rescue.exe"
```

`install.ps1` 脚本在安装时检查 `(Get-MpComputerStatus).SmartAppControlState`，如果 SAC 处于开启状态则打印相应指引。

---

## Agentic 工具调用

默认模式下，ds-rescue 在 **agentic 工具调用循环**中运行 DeepSeek 模型（最多 20 轮）。模型在 review 过程中可以调用三个工具：

| 工具 | 功能 |
|---|---|
| `read_file` | 读取指定路径的文件（用于检查被引用的代码或配置） |
| `write_file` | 写入文件（用于生成结构化输出或修复建议） |
| `bash_exec` | 执行 shell 命令（用于运行 `git log`、`grep` 等） |

工具调用循环受拒绝模式列表保护。匹配 `sudo`、`rm -rf /`、fork bomb 或通过 curl 泄露凭证等模式的命令，在执行前会被拒绝。使用 `--no-tools` 可以完全禁用 agentic 模式，以单次调用方式运行。

---

## 贡献

参见 [CONTRIBUTING.md](CONTRIBUTING.md) 了解贡献规范、提交消息格式和安全规则。特别欢迎针对 `SKILL.md` prompt 定义的 PR——你不需要写任何 Go 代码就能对 ds-rescue 做出有意义的改进。

---

## 许可证

MIT — 参见 [LICENSE](./LICENSE)

Copyright (c) 2026 Daniel

---

## 相关工作与引用

- [codex-plugin-cc](https://github.com/anthropics/codex-plugin-cc) — 将跨家族 LLM review 引入 Claude Code 的先驱
- [DeepSeek API 文档](https://api-docs.deepseek.com/) — 官方 API 参考，含模型和定价信息
- [Smart App Control 概述](https://learn.microsoft.com/zh-cn/windows/apps/develop/smart-app-control/overview) — Microsoft 官方 SAC 文档
- [Generative Engine Optimization](https://en.wikipedia.org/wiki/Generative_engine_optimization) — 面向 AI 搜索可见性的 README 结构原则

---

由 [Daniel (zred0627)](https://github.com/zred0627) 维护 · v0.1.0（2026-05）
