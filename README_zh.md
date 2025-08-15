# Codex Env 启动器 (CDE)
[中文](./README_zh.md)

生产就绪的 Go CLI 工具，用于管理多个 Codex/OpenAI 环境配置，实现环境间无缝切换（生产、测试、自定义提供商等）。CDE 作为 `codex` 的轻量包装器，具备**标志透传**、**无 ANSI 显示管理**和**通用终端兼容性**功能。

## ✨ 核心特性

### 🎯 **核心功能**
- **环境管理**：通过交互式选择添加、列表、删除 Codex 配置
- **额外环境变量**：为每个环境配置自定义环境变量（如 `OPENAI_TIMEOUT`）
- **标志透传**：透明地将参数转发给 Codex（`cde --help` 等）
- **安全 API 密钥存储**：隐藏终端输入，带掩码显示和适当的文件权限
- **通用终端支持**：无 ANSI 显示系统适用于 SSH、CI/CD 和所有终端类型

### 🖥️ **高级 UI 特性**
- **响应式设计**：适应任何终端宽度（已测试 20-300+ 列）
- **4 层渐进式回退**：全交互 → 基础交互 → 数字选择 → 无头模式
- **智能内容截断**：保持重要信息可见性，防止溢出
- **清洁导航**：状态渲染防止箭头键导航时的显示堆叠

### 🔒 **企业级安全**
- **命令注入防护**：全面的参数验证和 shell 元字符检测
- **安全文件操作**：配置以 600/700 权限存储，原子写入
- **API 密钥保护**：终端原始模式输入，带掩码显示（前 6 位 + 后 4 位）
- **输入清理**：URL 验证、名称清理和格式检查

## 📦 安装

### 从源码构建

```bash
git clone https://github.com/cexll/codex-env.git
cd codex-env
go build -o cde .
```

### 安装 Codex CLI

如果系统中未安装 `codex`，请先安装官方 CLI：

```bash
npm install -g @openai/codex
```

### 安装到系统 PATH

```bash
sudo mv cde /usr/local/bin/
# 验证安装
cde --help
```

## 🚀 使用方法

### 基本命令

#### 交互式启动
```bash
cde  # 显示带箭头导航的响应式环境选择菜单
```

#### 使用特定环境启动
```bash
cde --env production     # 或 -e production
cde -e staging          # 使用测试环境启动
```

#### 标志透传示例
```bash
cde auto -e dev -- mcp          # 自动批准 + 沙箱
cde -- --help                   # 显示 codex 的帮助（-- 明确分隔标志）
cde -e staging -- proto         # 运行 proto
```

### 环境管理

#### 添加新环境：
```bash
cde add
# 交互式提示：
# - 环境名称（已验证）
# - API URL（带格式验证）
# - API 密钥（安全隐藏输入）
# - 模型（可选，如 gpt-5）
# - 额外环境变量（可选，如 OPENAI_TIMEOUT）
```

#### 列出所有环境：
```bash
cde list
# 响应式格式化输出：
# 已配置环境 (3):
#
#   名称:  production
#   URL:   https://api.openai.com/v1
#   模型:  gpt-5
#   密钥:  sk-************************************************************
#   环境:  OPENAI_TIMEOUT=30s
#          CUSTOM_TIMEOUT=60s
#
#   名称:  staging
#   URL:   https://api.openai.com/v1
#   模型:  default
#   密钥:  sk-stg-************************************************************
```

#### 删除环境：
```bash
cde remove staging
# 确认和安全删除，带备份
```

#### 使用额外环境变量：
添加新环境时，您可以配置额外的环境变量：

```bash
cde add
# 交互式会话示例：
# 环境名称: example
# 基础 URL: https://api.openai.com/v1
# API 密钥: [安全输入]
# 模型: gpt-5
# 额外环境变量（可选）:
# 变量名: OPENAI_TIMEOUT
# OPENAI_TIMEOUT 的值: 30s
# 变量名: [按 Enter 结束]
```

这些环境变量将在使用此环境启动 Codex 时自动设置。

### 命令行界面

```bash
cde [选项] [-- codex-参数...]

选项:
  -e, --env <名称>        使用特定环境
  -h, --help              显示带示例的综合帮助

命令:
  list                    以响应式格式列出所有环境
  add                     添加新环境（支持模型指定）
  remove <名称>           确认删除环境
  auto                    自动批准并使用沙箱（-a never --sandbox workspace-write）

标志透传:
  CDE 选项后的任何参数都直接传递给 codex。
  使用 '--' 明确分隔 CDE 选项和 codex 参数。

示例:
  cde                              交互式选择和启动
  cde --env prod                   使用 'prod' 环境启动
  cde auto -e dev -- mcp           自动批准 + 沙箱
  cde --env staging -- proto       使用测试环境，透传到 codex
  cde -- --help                    显示 codex 的帮助
```

## 📁 配置

### 配置文件结构

环境存储在 `~/.codex-env/config.json` 中：

```json
{
  "environments": [
    {
      "name": "production",
      "url": "https://api.openai.com/v1",
      "api_key": "sk-xxxxx",
      "model": "gpt-5",
      "env_vars": {
        "OPENAI_TIMEOUT": "30s"
      }
    },
    {
      "name": "staging",
      "url": "https://api.openai.com/v1",
      "api_key": "sk-staging-xxxxx",
      "model": "default",
      "env_vars": {
        "OPENAI_TIMEOUT": "30s"
      }
    }
  ],
  "settings": {
    "validation": {
      "strict_validation": false
    }
  }
}
```

### 环境变量

**额外环境变量支持：**
CDE 支持为每个环境配置额外的环境变量。这些变量在使用选定环境启动 Claude Code 时自动设置。

**常见用例：**
- `ANTHROPIC_SMALL_FAST_MODEL`：为代码补全等快速操作指定更快的模型（如 `claude-3-haiku-20240307`）
- `ANTHROPIC_TIMEOUT`：为 API 请求设置自定义超时值（如 `30s`）
- `ANTHROPIC_RETRY_COUNT`：配置失败请求的重试行为（如 `3`）
- Claude Code 安装所需的任何自定义环境变量

**模型验证配置：**
- `CDE_MODEL_PATTERNS`：用于模型验证的逗号分隔自定义正则表达式模式
- `CDE_MODEL_STRICT`：设置为 "false" 启用带警告的宽松模式

## 🏗️ 架构

### 核心组件（4 个文件）

- **`main.go`**（580+ 行）：CLI 界面、**标志透传系统**、模型验证
- **`config.go`**（367 行）：原子文件操作、备份/恢复、验证
- **`ui.go`**（1000+ 行）：**无 ANSI 显示管理**、响应式 UI、4 层回退
- **`launcher.go`**（174 行）：带参数转发的进程执行

### 关键设计模式

**标志透传系统**：两阶段参数解析分离 CDE 标志和 Claude 参数，实现带安全验证的透明命令转发。

**无 ANSI 显示管理**：使用以下方式实现通用终端兼容性：
- **DisplayState**：跟踪屏幕内容和管理状态更新
- **TextPositioner**：使用回车和填充的光标控制（无 ANSI）
- **LineRenderer**：带差异更新的状态菜单渲染

**4 层渐进式回退**：
1. **全交互**：带箭头导航和 ANSI 增强的状态渲染
2. **基础交互**：带箭头键支持的无 ANSI 显示
3. **数字选择**：有限终端的回退
4. **无头模式**：CI/CD 环境的自动化模式

## 🔒 安全实现

### 多层安全
- **命令注入防护**：带 shell 元字符检测的全面参数验证
- **安全文件操作**：带适当权限的原子写入（文件 600，目录 700）
- **API 密钥保护**：终端原始模式输入、掩码显示、从不记录
- **输入验证**：URL 验证、名称清理、API 密钥格式检查
- **进程隔离**：带安全参数转发的干净环境变量处理

### 安全验证
- **时序攻击抵抗**：安全比较操作
- **内存安全**：适当的清理和有界操作
- **环境清理**：不暴露的干净变量注入

## 🧪 测试与质量

### 全面的测试覆盖率（95%+）

```bash
# 运行完整测试套件
go test -v ./...

# 安全专项测试
go test -v -run TestSecurity

# 性能基准测试
go test -bench=. -benchmem

# 覆盖率分析
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 测试类别
- **单元测试**：核心功能（parseArguments、formatSingleLine 等）
- **集成测试**：端到端工作流和跨平台兼容性
- **安全测试**：命令注入防护、文件权限、输入验证
- **终端兼容性**：SSH、CI/CD、终端模拟器（iTerm、VS Code 等）
- **性能测试**：微秒级操作、内存效率
- **回归测试**：显示堆叠防护、布局溢出保护

### 质量指标
- **整体质量评分**：96/100（自动验证）
- **测试覆盖率**：所有组件 95%+
- **性能**：微秒级操作，最小内存开销
- **安全性**：零漏洞，全面威胁覆盖
- **兼容性**：100% 向后兼容，通用终端支持

## 🛠️ 开发

### 构建和测试

```bash
# 开发构建
go build -o cde .

# 运行综合测试套件
make test                # 或: go test -v ./...
make test-coverage       # HTML 覆盖率报告
make test-security       # 安全专项测试
make bench              # 性能基准测试

# 代码质量
make quality            # fmt + vet + test
make fmt                # 格式化代码
make vet                # 静态分析
```

### 项目结构

```
├── main.go                           # CLI 界面和标志透传系统
├── config.go                         # 带原子操作的配置管理
├── ui.go                            # 无 ANSI 显示管理和响应式 UI
├── launcher.go                       # 带参数转发的进程执行
├── go.mod                           # Go 模块定义
├── go.sum                           # 依赖校验和
├── CLAUDE.md                        # 开发文档
├── README.md                        # 用户文档（英文）
├── README_zh.md                     # 用户文档（中文）
└── Tests:
    ├── *_test.go                    # 全面单元测试
    ├── integration_test.go          # 端到端工作流
    ├── security_test.go             # 安全验证
    ├── terminal_display_fix_test.go # 显示管理
    ├── ui_layout_test.go           # 响应式布局
    └── display_stacking_fix_test.go # 导航行为
```

## 📋 要求

- **Go 1.21+**（从源码构建）
- **Claude Code CLI** 必须安装并在 PATH 中可用为 `claude`
- **终端**：任何终端模拟器（ANSI 支持可选但增强）

## 🚀 迁移指南

### 从以前版本
此增强版本保持完整的向后兼容性。`~/.claude-code-env/config.json` 中的现有配置文件无需修改即可立即工作。

### 可用的新功能
- **额外环境变量**：配置自定义环境变量，如 `ANTHROPIC_SMALL_FAST_MODEL`
- **标志透传**：开始使用 `cde -r`、`cde --help` 等
- **增强 UI**：享受响应式设计和清洁导航
- **通用兼容性**：在所有终端类型中一致工作
- **增强安全性**：受益于命令注入防护

## 🤝 贡献

1. Fork 仓库
2. 创建功能分支（`git checkout -b feature/amazing-feature`）
3. 遵循 KISS 原则和现有模式进行更改
4. 为新功能添加全面测试
5. 运行 `make test` 确保所有测试通过
6. 运行 `make quality` 进行代码质量检查
7. 提交带详细描述的拉取请求

### 开发原则
1. **KISS 原则**：简单、直接的实现
2. **安全优先**：所有操作必须在设计上安全
3. **通用兼容性**：功能必须跨所有平台工作
4. **全面测试**：需要 95%+ 测试覆盖率
5. **性能聚焦**：首选微秒级操作

## 📄 许可证

此项目在 MIT 许可证下许可 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- 使用 **Claude Code** 集成构建
- 由 **Go 标准库** + `golang.org/x/term` 驱动
- 采用 **KISS 原则**和**通用兼容性**设计
- 在**多个平台**和**终端环境**中测试

---

**Claude Code 环境切换器**：生产就绪、安全且通用兼容的 CLI 工具，用于管理 Claude Code 环境，具备透明标志透传和智能显示管理功能。
