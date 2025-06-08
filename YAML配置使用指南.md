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
  # 如果选择使用deepSeek，具体参考https://api-docs.deepseek.com/
  provider: "deepSeek"

  # API密钥 - 请填入您的真实API Key
  api_key: "your-api-key-here"

  # API基础URL
  base_url: "https://api.deepseek.com"

  # 使用的模型
  model: "deepSeek-chat"

  # 最大token数量
  max_tokens: 100

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
  port: 8082
  debug: true
```

## 🔧 配置方式优先级

系统会按以下优先级读取配置：

1. **YAML 配置文件** (`config.yml` 或 `config.yaml`)
2. **环境变量**
3. **代码中的默认值**

## 🌍 环境变量支持

即使使用 YAML 配置文件值为空时，可以通过环境变量设置：

```bash
export AI_API_KEY="sk-xxxxxxxx..."
export AI_PROVIDER="xxxxx"
export AI_MODEL="xxxxx"
```

## 🔒 安全最佳实践

### 1. 使用 .gitignore

确保 `config.yml` 已添加到 `.gitignore` 文件中：

```gitignore
# 配置文件（包含敏感信息）
config.yml
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
配置加载成功，AI提供商: xxxxx
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

**错误信息**: `AI API key not configured in config.yml`

**解决方案**:

- 确认 API Key 格式正确（以 sk-开头）
- 检查 API Key 是否有效且未过期
- 确认账户余额充足
