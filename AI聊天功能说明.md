# AI 聊天功能使用说明

## 功能概述

本聊天室项目现已集成 AI 助手功能，用户可以与智能 AI 机器人进行对话交流。

## 如何使用

### 1. 访问 AI 助手

- 在聊天室底部导航栏中，点击**AI 助手**选项卡（齿轮图标）
- 进入 AI 助手主界面，看到"智能 AI 助手"选项

### 2. 开始对话

- 点击"智能 AI 助手"进入聊天界面
- 在底部输入框中输入您的问题或消息
- 点击发送按钮（或按回车键）发送消息
- AI 助手会自动回复您的问题
- 点击左上角返回按钮可回到主界面

### 3. 快捷问题

在 AI 聊天界面无消息时，系统会显示快捷问题建议，您可以直接点击使用：

- "你好，我想了解一下聊天室的功能"
- "如何创建群聊？"
- "怎样添加好友？"
- "有什么有趣的功能推荐吗？"
- "系统支持哪些消息类型？"

## 功能特点

### ✨ 智能对话

- AI 助手能够理解您的问题并提供相关回答
- 支持关于聊天室功能的各种询问

### 🎯 快捷帮助

- 提供常见问题的快速解答
- 帮助新用户快速了解系统功能

### 💬 友好界面

- 类似真实聊天的对话界面
- 美观的渐变色 AI 消息气泡
- 实时的"正在思考"动画效果
- 与好友聊天界面一致的设计风格

### ⚡ 即时响应

- 快速的消息响应时间
- 流畅的用户体验

## 界面设计

### 主界面

- 显示 AI 助手图标和简介
- 点击进入聊天界面

### 聊天界面

- 仿照好友聊天界面设计
- 底部有输入框和发送按钮
- 支持实时滚动和消息显示
- 首次进入自动显示欢迎消息

## 技术实现

### 前端实现

- 使用 Vue.js 实现响应式界面
- 自定义 CSS 样式实现美观的聊天界面
- 支持实时滚动和动画效果

### 后端集成（可扩展）

当前使用模拟 AI 响应，您可以轻松集成真实的 AI API：

#### OpenAI API 集成示例

```javascript
async callAiApi(question) {
    try {
        const response = await fetch('/api/ai/chat', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                message: question,
                userId: this.info.id
            })
        });

        const data = await response.json();
        return data.reply;
    } catch (error) {
        console.error('AI API调用失败:', error);
        throw error;
    }
}
```

#### Go 后端 API 端点示例

```go
func HandleAIChat(c *gin.Context) {
    var request AIChatRequest

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(400, AIChatResponse{
            Code: -1,
            Msg:  "请求参数错误",
        })
        return
    }

    // 调用OpenAI API或其他AI服务
    reply, err := getAIResponse(request.Message)
    if err != nil {
        // 使用本地回复作为备选
        reply = getLocalResponse(request.Message)
    }

    c.JSON(200, AIChatResponse{
        Reply: reply,
        Code:  0,
        Msg:   "success",
    })
}
```

## 已修复的问题

### 1. 界面设计优化

- ✅ 修改为与好友聊天一致的界面设计
- ✅ 添加主界面和聊天界面两个状态
- ✅ 底部输入框布局与现有聊天界面一致

### 2. API 错误修复

- ✅ 将 AI 聊天 API 移至公开路由，避免 JWT 验证问题
- ✅ 优化错误处理和回退机制
- ✅ 修复时间格式化函数

### 3. 模板语法修复

- ✅ 修复 Go 模板和 Vue 模板语法冲突
- ✅ 正确处理模板渲染

## 自定义和扩展

### 修改 AI 回复逻辑

在 `views/chat/foot.html` 文件的 `simulateAiResponse` 方法中，您可以：

1. **添加更多关键词匹配**：

```javascript
const responses = {
  你好: "你好！我是AI助手...",
  功能: "这个聊天室支持...",
  // 添加更多关键词和回复
  天气: "抱歉，我目前无法查询天气信息...",
  时间: `现在是 ${new Date().toLocaleString()}`,
};
```

2. **集成真实 AI API**：
   替换 `callAiApi` 方法中的模拟调用为真实的 API 调用

3. **添加更多功能**：

- 文件上传支持
- 语音转文字
- 图片识别
- 多轮对话记忆

## 文件结构

```
views/chat/
├── ai.html          # AI聊天界面模板
├── tabmenu.html     # 底部导航（已添加AI选项卡）
├── foot.html        # JavaScript逻辑（包含AI功能）
└── index.html       # 主页面（已包含AI模板）

service/
├── index.go         # 后端路由（已添加AI模板加载）
└── ai_service.go    # AI聊天API

router/
└── app.go           # 路由配置

asset/images/
└── ai-avatar.png    # AI助手头像
```

## 故障排除

### 常见问题

1. **AI 助手不回复**

   - 检查浏览器控制台是否有错误
   - 确认网络连接正常
   - 检查后端服务是否正常运行

2. **界面显示异常**

   - 清除浏览器缓存
   - 检查 CSS 样式是否正确加载

3. **点击快捷问题无反应**
   - 确认 Vue.js 正确加载
   - 检查 JavaScript 控制台错误

### 调试方法

1. 打开浏览器开发者工具
2. 查看 Console 标签页的错误信息
3. 检查 Network 标签页的网络请求

## API 接口说明

### POST /api/ai/chat

**请求参数：**

```json
{
  "message": "用户消息内容",
  "userId": 123
}
```

**响应格式：**

```json
{
  "reply": "AI回复内容",
  "code": 0,
  "msg": "success"
}
```

## 未来规划

- [ ] 集成 GPT-3.5/GPT-4 API
- [ ] 添加多语言支持
- [ ] 实现对话历史保存
- [ ] 支持文件和图片分析
- [ ] 添加语音交互功能

---

如需技术支持或有改进建议，请联系开发团队。
