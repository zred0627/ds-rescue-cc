# ds-rescue-cc

> **针对你的方案和提交的跨家族对抗性 review — 由 DeepSeek 驱动，成本比 Codex 低 50 倍。**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00ADD8.svg)](https://go.dev)
[![Releases](https://img.shields.io/github/v/release/zred0627/ds-rescue-cc)](https://github.com/zred0627/ds-rescue-cc/releases)

[English README](README.md)

---

## 工程痛点：同家族模型 review 是回音壁

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

---

## 跨家族 review 为什么能找到更多 bug

软件工程里有一个被充分研究的类比：**代码 review 效果随 reviewer 多样性而提升**。关于结对 review 和审查程序的研究持续表明，当 reviewer 背景不同、先验知识不同时，发现的缺陷数量是同质 reviewer 的 1.7–2.3 倍。

同样的原则适用于 LLM review。来自不同训练家族的模型：

- 对什么是"正常"有不同的偏见
- 对哪些边界条件值得提及有不同的先验
- 有不同的失败模式，与作者的失败模式不容易重叠

DeepSeek v4-pro 在数据分布和架构选择上与 Claude 有实质性差异。当 Claude 漏掉某些东西，DeepSeek 往往能抓住——反之亦然。这种互补性就是这里要释放的工程价值。

目标不是替代 Claude，而是在工作流中加入一个来自不同家族的独立第二 reviewer——就像你不会让 PR 的作者同时成为它唯一的 reviewer 一样。

---

## 为什么不直接用 Codex？

[codex-plugin-cc](https://github.com/anthropics/codex-plugin-cc) 开创了把非 Claude reviewer 引入 Claude Code 工作流的先例。它作为开创者理应得到完整的认可，也使 ds-rescue-cc 的存在成为可能。

实际限制在于成本。按典型方案/提交体量，单次 Codex review 调用大约花费 $0.30–0.60 USD。偶尔使用——每周的架构 review、上线前审计——这个成本完全可以接受。但贵到让大多数开发者不会在每份方案、每次提交、每次方案比选时都跑一遍。**习惯无法形成**。

开发者工具的习惯形成，要求多用一次的边际成本低到可以忽略不计。$0.30–0.60 一次，这个门槛没有达到。

ds-rescue-cc 建立在一条不同的成本曲线上。

---

## 为什么是 DeepSeek：工程经济性 + 跨家族互补性

### 成本

按典型输入/输出体量，DeepSeek v4-pro 每次 review 调用大约只需 **$0.01**。这是等量 Codex review 成本的 1/50。

在 $0.01 一次 review 的情况下，习惯性使用的经济门槛实际上消失了。在每一次有意义的提交上跑一次 review 感觉是免费的——因为在那个价位上，它确实约等于免费。这是选择 DeepSeek 作为 review 引擎的首要工程论据。

### 质量

DeepSeek v4-pro 在 SWE-Bench 和编程基准测试上接近 GPT-4 级别的表现。对于 review 场景——分析方案文档或 git diff，找出问题——与最贵的替代品相比，质量差异并不显著。它足够好，能抓住真实的 bug，如上面的案例所示。

### 跨家族互补性

DeepSeek 在与 OpenAI 或 Anthropic 模型不同的数据分布和架构选择上训练而成。这正是它作为 reviewer 有价值的属性：它的盲点是不同的。与 Claude 撰写的方案结合，你得到的是真正的独立 review，而不是作者自己分析的改写版。

### 中文领域推理深度

对于使用中文工作、或处理中文技术文献相关问题的团队（制造系统、物流基础设施、政策分析），DeepSeek 的中文推理质量明显强于大多数西方来源的模型。这不是次要功能——在某些领域，这是首要差异化点。

---

## 三种模式：ds-rescue 做什么

ds-rescue 有三种模式，对应典型开发工作流中的三种 review 场景：

| 模式 | review 对象 | 使用时机 |
|------|-------------|----------|
| `plan` | 实施方案文档（任务 DAG、架构决策、风险清单） | 开始一个功能或迭代之前 |
| `scheme` | 方案比选表格（A vs B vs C 分析、选项权衡） | 在多个架构备选方案之间做选择时 |
| `execution` | 最后一次提交或已暂存变更的 git diff | 完成一个工作单元、push 之前 |

ds-rescue **不是**聊天工具的替代品。它不回答问题，也不生成代码。它是单一用途的 reviewer：给它一份文档或 diff，从一个没有参与生成被 review 工件的模型那里得到结构化发现。

每种模式的 prompt 行为定义在：

```
plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md
```

这个文件是 ds-rescue 说什么、怎么格式化输出的唯一真源——你可以读它、fork 它、按需定制，完全不需要动任何 Go 代码。

---

## 快速上手：60 秒完成第一次 review

### 第一步：安装二进制

**Linux / macOS:**

```bash
curl -L https://github.com/zred0627/ds-rescue-cc/releases/latest/download/ds-rescue-linux-amd64 \
  -o ds-rescue && chmod +x ds-rescue && sudo mv ds-rescue /usr/local/bin/
```

macOS (Apple Silicon):

```bash
curl -L https://github.com/zred0627/ds-rescue-cc/releases/latest/download/ds-rescue-darwin-arm64 \
  -o ds-rescue && chmod +x ds-rescue && sudo mv ds-rescue /usr/local/bin/
```

**Windows (PowerShell):**

```powershell
Invoke-WebRequest -Uri https://github.com/zred0627/ds-rescue-cc/releases/latest/download/ds-rescue-windows-amd64.exe `
  -OutFile ds-rescue.exe
Move-Item ds-rescue.exe "$env:LOCALAPPDATA\Programs\ds-rescue\ds-rescue.exe"
```

**Go install（任何平台，需 Go 1.22+）:**

```bash
go install github.com/zred0627/ds-rescue-cc/cmd/ds-rescue@latest
```

### 第二步：设置 DeepSeek API Key

```bash
# 方式 A：环境变量（推荐用于 CI）
export DEEPSEEK_API_KEY=sk-你的-key

# 方式 B：key 文件（推荐用于交互式使用，重启 shell 后依然有效）
mkdir -p ~/.ds-rescue && echo "sk-你的-key" > ~/.ds-rescue/key && chmod 600 ~/.ds-rescue/key
```

在 [platform.deepseek.com](https://platform.deepseek.com) 获取 DeepSeek API Key。

### 第三步：运行第一次 review

```bash
# review 一份方案文档
ds-rescue --mode plan path/to/your-plan.md

# review 一份方案比选
ds-rescue --mode scheme path/to/scheme-comparison.md

# review 你的最后一次提交
ds-rescue --mode execution
```

### 自检

```bash
ds-rescue --check
```

验证：key 已找到、二进制版本正常、SKILL.md 可加载、模型响应正常。

---

## Claude Code 插件：一键安装

如果你使用 Claude Code，可以把 ds-rescue 安装为插件，在你的会话中添加 `/ds-rescue` 斜杠命令：

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

插件还会注册一个 `SessionStart` hook，在启动时检查 ds-rescue 二进制是否存在，如缺失则提示安装。检查只是一次 `test -f` 调用，不会给 session 启动带来可感知的延迟。

### 插件在 CLI 基础上提供什么

插件封装层：
- 自动设置 `$CLAUDE_PLUGIN_ROOT`，让二进制能自动找到 SKILL.md
- 把发现内容内联展示在 Claude Code 对话中
- 让你可以在后续消息中引用发现，不需要离开会话

底层二进制与独立 CLI 完全相同——插件是薄薄的集成层，不是单独的产品。

---

## 通用 CLI：在任何环境下都能工作

ds-rescue 是标准 CLI 二进制，可以在任何能执行 shell 命令的环境中工作——不需要 Claude Code。

**opencode:**

```bash
ds-rescue --mode plan docs/plans/current-plan.md
```

**Continue (VS Code / JetBrains):**

在 Continue 配置中添加自定义命令：

```json
{
  "name": "ds-rescue review",
  "command": "ds-rescue --mode execution",
  "description": "对当前变更运行 DeepSeek 对抗性 review"
}
```

**Aider:**

```bash
# 在 aider 会话开始前 review 当前状态
ds-rescue --mode execution && aider --message "修复 ds-rescue 发现的问题"
```

**裸 shell / CI:**

```bash
# 在 pre-push hook 或 CI 步骤中
ds-rescue --mode execution --no-tools || echo "Review 发现了问题，请检查 stderr"
```

二进制从 `stdin` 或文件参数读取，把发现写到 `stdout`，把诊断信息写到 `stderr`。可以与任何工具链组合使用。

---

## 定制 Skill Prompt

三种模式的 review 行为定义在单一 Markdown 文件中：

```
plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md
```

这个文件有三个 section（`## Plan Mode`、`## Scheme Mode`、`## Execution Mode`），每个包含对应 review 类型的系统 prompt 和输出 schema。二进制在运行时读取这个文件——它不是嵌入在二进制里的。

**定制方法：**

```bash
# fork 仓库，按你的需求编辑 SKILL.md
vim plugins/ds-rescue/skills/ds-plan-challenger/SKILL.md

# 让 ds-rescue 指向你的定制版本
export DS_RESCUE_SKILL_PATH=/path/to/your/SKILL.md
ds-rescue --mode plan my-plan.md

# 或者每次调用时传入
ds-rescue --skill /path/to/your/SKILL.md --mode plan my-plan.md
```

**定制方向的想法：**

- 金融/投资备忘录 review（标记缺失假设、检查 DCF 输入）
- 安全审计视角（标记注入面、认证边界、密钥处理）
- 基础设施即代码 review（标记缺失回滚路径、爆炸半径）
- 中文推理增强模式（为中国市场背景提供更深入的分析）

由于 SKILL.md 是纯 Markdown 且二进制从磁盘加载，你可以维护一批专用 review prompt，通过一个 flag 在它们之间切换。改进 prompt 不需要任何 Go 知识——一份改进 SKILL.md 的 PR，与改进 Go 代码的 PR 同等有价值。

---

## SKILL.md 路径解析

二进制按以下优先级查找 SKILL.md：

1. `--skill <path>` CLI flag（最高优先级）
2. `$DS_RESCUE_SKILL_PATH` 环境变量
3. `$CLAUDE_PLUGIN_ROOT/skills/ds-plan-challenger/SKILL.md`（由 Claude Code 插件自动设置）
4. `~/.ds-rescue/skills/ds-plan-challenger/SKILL.md`（独立安装）
5. 如果都找不到：二进制退出并列出所有尝试过的路径

运行 `ds-rescue --check` 可以看到解析到哪条路径，并确认文件可加载。

---

## Agentic 工具调用

默认模式下，ds-rescue 在 **agentic 工具调用循环**中运行 DeepSeek 模型（最多 20 轮）。模型在 review 过程中可以调用三个工具：

| 工具 | 功能 |
|------|------|
| `read_file` | 读取指定路径的文件（用于检查被引用的代码或配置） |
| `write_file` | 写入文件（用于生成结构化输出或修复建议） |
| `bash_exec` | 执行 shell 命令（用于运行 `git log`、`grep` 等） |

工具调用循环受拒绝模式列表保护。匹配 `sudo`、`rm -rf /`、fork bomb 或通过 curl 泄露凭证等模式的命令，在执行前会被拒绝。使用 `--no-tools` 可以完全禁用 agentic 模式，以单次调用方式运行。

---

## CLI 参数参考

```
ds-rescue [flags] [<input-file>]

Flags:
  --mode {plan|scheme|execution}   review 模式（默认：自动检测）
  --plan                           等同于 --mode plan
  --scheme                         等同于 --mode scheme
  --execution                      等同于 --mode execution
  --skill <path>                   SKILL.md 路径（覆盖环境变量/插件/主目录查找）
  --model {pro|flash|reasoner}     DeepSeek 模型别名（默认：pro = deepseek-chat）
  --api-timeout <seconds>          API 调用超时（默认：90）
  --exec-timeout <seconds>         每次工具调用超时（默认：30）
  --no-tools                       禁用 agentic 工具调用循环
  --verbose                        把每次工具调用记录到 stderr
  --version                        输出版本和 git commit
  --check                          自检：key 找到了？二进制正常？SKILL.md 可加载？模型响应？
```

---

## 许可证

MIT — 参见 [LICENSE](LICENSE)。

Copyright (c) 2026 Daniel

---

## 贡献

参见 [CONTRIBUTING.md](CONTRIBUTING.md) 了解贡献规范、提交消息格式和安全规则。特别欢迎针对 SKILL.md prompt 定义的 PR——你不需要写任何 Go 代码就能对 ds-rescue 做出有意义的改进。

---

## 致谢

[codex-plugin-cc](https://github.com/anthropics/codex-plugin-cc) 建立了把非 Claude reviewer 引入 Claude Code 工作流的先例。ds-rescue-cc 在那个基础上，以不同的成本结构和不同的模型家族继续前行。没有那项开创性工作，这个思路不会存在。
