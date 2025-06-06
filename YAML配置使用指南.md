# YAML 配置文件使用指南

## 🎯 快速开始

### 1. 复制配置文件

```bash
cp config.example.yml config.yml
```

### 2. 编辑配置文件

打开 `config.yml` 文件，填入您的 API Key：

```yaml
ai:
  api_key: "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

### 3. 启动应用

```bash
go run main.go
```

## 📋 完整配置说明

### AI 配置

```yaml
ai:
  # 服务提供商 (openai, azure, claude等)
  provider: "openai"

  # API密钥 - 在这里填入您的真实API Key
  api_key: "your-openai-api-key-here"

  # API基础URL
  base_url: "https://api.openai.com/v1"

  # 使用的模型
  model: "gpt-3.5-turbo"

  # 最大token数量
  max_tokens: 150

  # 请求超时时间(秒)
  timeout: 30
```

### 数据库配置

```yaml
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: ""
  database: "chatroom"
```

### 服务器配置

```yaml
server:
  host: "localhost"
  port: 8081
  debug: true
```

## 🔧 配置方式优先级

系统会按以下优先级读取配置：

1. **YAML 配置文件** (`config.yml` 或 `config.yaml`)
2. **环境变量**
3. **代码中的默认值**

## 🌍 环境变量支持

即使使用 YAML 配置，您仍然可以通过环境变量覆盖设置：

```bash
export OPENAI_API_KEY="sk-xxxxxxxx..."
export AI_API_KEY="sk-xxxxxxxx..."
export AI_PROVIDER="openai"
export AI_MODEL="gpt-4"
```

## 🔒 安全最佳实践

### 1. 使用 .gitignore

确保 `config.yml` 已添加到 `.gitignore` 文件中：

```gitignore
# 配置文件（包含敏感信息）
config.yml
config.yaml
```

### 2. 权限设置

```bash
# 设置配置文件权限（仅所有者可读写）
chmod 600 config.yml
```

### 3. 环境变量

生产环境推荐使用环境变量而不是配置文件：

```bash
# 设置环境变量
export AI_API_KEY="sk-xxxxxxxx..."

# 启动应用（不需要config.yml）
go run main.go
```

## 📝 配置验证

应用启动时会显示配置加载状态：

```
配置文件不存在，使用默认配置。请复制 config.example.yml 为 config.yml 并配置您的API密钥。
```

或者：

```
配置加载成功，AI提供商: openai
```

## 🚀 不同 AI 服务商配置

### OpenAI

```yaml
ai:
  provider: "openai"
  api_key: "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  base_url: "https://api.openai.com/v1"
  model: "gpt-3.5-turbo"
```

### Azure OpenAI

```yaml
ai:
  provider: "azure"
  api_key: "your-azure-api-key"
  base_url: "https://your-resource.openai.azure.com/"
  model: "gpt-35-turbo"
```

### 自定义 API（兼容 OpenAI 格式）

```yaml
ai:
  provider: "custom"
  api_key: "your-custom-api-key"
  base_url: "https://your-custom-api-endpoint.com/v1"
  model: "custom-model"
```

## 🔧 故障排除

### 问题 1: 配置文件读取失败

**错误信息**: `读取配置文件失败: no such file or directory`

**解决方案**:

```bash
# 检查文件是否存在
ls -la config.yml

# 如果不存在，复制示例文件
cp config.example.yml config.yml
```

### 问题 2: YAML 格式错误

**错误信息**: `解析配置文件失败: yaml: line X: found character`

**解决方案**:

- 检查 YAML 格式（注意缩进，使用空格而不是制表符）
- 使用在线 YAML 验证器检查格式
- 确保字符串值用引号包围

### 问题 3: API Key 无效

**错误信息**: `OpenAI API key not configured in config.yml`

**解决方案**:

- 确认 API Key 格式正确（以 sk-开头）
- 检查 API Key 是否有效且未过期
- 确认账户余额充足

## 📊 配置监控

### 检查当前配置

启动应用时查看控制台输出，确认配置加载状态。

### API 调用统计

可以在后端添加日志来监控 API 调用：

```go
fmt.Printf("AI API调用: 模型=%s, Token=%d\n", aiConfig.Model, aiConfig.MaxTokens)
```

## 🎮 键盘快捷键功能

除了 YAML 配置，我们还增强了用户体验：

### 键盘快捷键

- **Enter**: 发送消息（在任何聊天输入框）
- **Ctrl+Enter**: 换行（如果需要多行输入）
- **Escape**: 返回主界面

### 滚动控制

- **鼠标滚轮**: 在所有页面都可以滚动
- **触摸滚动**: 支持移动设备滑动
- **自动滚动**: 新消息自动滚动到底部

---

现在您的聊天室支持：
✅ YAML 配置文件管理
✅ 安全的 API Key 存储
✅ 鼠标滚轮控制
✅ 回车键发送消息
✅ 多种 AI 服务商支持
