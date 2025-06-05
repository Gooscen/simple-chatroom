# AI API 集成指南

## 当前状态说明

**目前的 AI 回答是预设的模拟回复，不是真正的 AI 服务。**

要使用真正的 AI API，请按照以下步骤配置：

## 快速配置指南

### 方式 1: 直接修改代码（简单方式）

在 `service/ai_service.go` 文件的第 68 行左右，找到：

```go
// TODO: 在这里设置您的OpenAI API Key
apiKey = "your-openai-api-key-here"
```

将 `your-openai-api-key-here` 替换为您的真实 API Key：

```go
apiKey = "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

### 方式 2: 使用环境变量（推荐方式）

1. **获取 API Key**

   - 访问 [OpenAI 官网](https://platform.openai.com/api-keys)
   - 注册并创建 API Key
   - 复制您的 API Key（格式如：sk-xxxxxxxx...）

2. **设置环境变量**

   **Mac/Linux:**

   ```bash
   export OPENAI_API_KEY="sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
   export AI_API_KEY="sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
   ```

   **Windows:**

   ```cmd
   set OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   set AI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   ```

3. **重启应用**
   ```bash
   go run main.go
   ```

## 支持的 AI 服务商

### 1. OpenAI (默认)

- 模型：gpt-3.5-turbo, gpt-4
- 官网：https://openai.com
- 文档：https://platform.openai.com/docs

### 2. 其他兼容 OpenAI 格式的服务

- Claude (Anthropic)
- Azure OpenAI
- 本地部署的模型

## 配置不同 AI 服务的方法

### OpenAI 配置示例

```bash
export AI_PROVIDER=openai
export AI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
export AI_BASE_URL=https://api.openai.com/v1
export AI_MODEL=gpt-3.5-turbo
```

### Azure OpenAI 配置示例

```bash
export AI_PROVIDER=azure
export AI_API_KEY=your-azure-api-key
export AI_BASE_URL=https://your-resource.openai.azure.com/
export AI_MODEL=gpt-35-turbo
```

## 测试 AI 集成

1. **启动应用**

   ```bash
   go run main.go
   ```

2. **访问聊天室**

   - 打开浏览器访问聊天室
   - 点击"AI 助手"选项卡
   - 发送消息测试

3. **检查是否使用真实 AI**
   - 问一些复杂问题
   - 查看回复是否更智能和多样化
   - 检查控制台日志

## 故障排除

### 问题 1: API Key 无效

**错误信息**: `OpenAI API error: 401 Unauthorized`

**解决方案**:

- 检查 API Key 是否正确
- 确认 API Key 是否有效且未过期
- 检查账户是否有余额

### 问题 2: 网络连接问题

**错误信息**: `dial tcp: connect: connection refused`

**解决方案**:

- 检查网络连接
- 确认防火墙设置
- 考虑使用代理（如在中国大陆）

### 问题 3: 配额限制

**错误信息**: `rate limit exceeded`

**解决方案**:

- 检查 API 使用限制
- 升级 API 计划
- 实现请求限流

## 成本考虑

### OpenAI 定价（2024 年参考价格）

- GPT-3.5-turbo: $0.0015/1K tokens (输入), $0.002/1K tokens (输出)
- GPT-4: $0.03/1K tokens (输入), $0.06/1K tokens (输出)

### 优化建议

- 限制每次对话的 token 数量
- 使用 GPT-3.5-turbo 而不是 GPT-4 以降低成本
- 实现对话历史限制
- 添加用户请求频率限制

## 高级配置

### 自定义 AI 行为

在 `service/ai_service.go` 中修改系统提示：

```go
{
    Role:    "system",
    Content: "你是一个专业的聊天室助手，专门帮助用户解决聊天室相关问题。请保持回答简洁、准确、友好。",
},
```

### 添加对话历史

可以扩展代码以支持多轮对话：

```go
type ConversationHistory struct {
    UserID   int
    Messages []Message
}
```

## 代码结构说明

```
service/
├── ai_service.go      # AI服务主要逻辑
└── index.go          # 路由注册

config/
└── ai_config.go      # AI配置管理（新增）

views/chat/
├── ai.html           # AI聊天界面
└── foot.html         # 前端JavaScript逻辑
```

## 验证集成成功

当您正确配置 API Key 后，AI 助手将：

1. **回复更智能**: 不再是预设的固定回复
2. **回复多样化**: 同样的问题会有不同的回答方式
3. **理解上下文**: 能够根据对话历史回答问题
4. **回答专业**: 对复杂问题给出更详细的解答

## 下一步计划

- [ ] 添加对话历史保存
- [ ] 实现多用户会话管理
- [ ] 添加文件上传和图片分析
- [ ] 集成更多 AI 服务商
- [ ] 添加 AI 能力扩展（如天气查询、网络搜索等）

---

如需技术支持，请参考项目文档或联系开发团队。
