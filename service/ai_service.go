package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AI聊天请求结构
type AIChatRequest struct {
	Message string `json:"message" binding:"required"`
	UserID  int    `json:"userId"`
}

// AI聊天响应结构
type AIChatResponse struct {
	Reply string `json:"reply"`
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
}

// OpenAI API请求结构
type OpenAIRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAI API响应结构
type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

// HandleAIChat 处理AI聊天请求
func HandleAIChat(c *gin.Context) {
	var request AIChatRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, AIChatResponse{
			Code: -1,
			Msg:  "请求参数错误",
		})
		return
	}

	// 调用AI服务获取回复
	reply, err := getAIResponse(request.Message)
	if err != nil {
		// 如果AI服务失败，使用本地智能回复
		reply = getLocalResponse(request.Message)
	}

	c.JSON(200, AIChatResponse{
		Reply: reply,
		Code:  0,
		Msg:   "success",
	})
}

// getAIResponse 调用真实的AI API（如OpenAI）
func getAIResponse(message string) (string, error) {
	// 从环境变量或配置文件读取API Key
	apiKey := os.Getenv("OPENAI_API_KEY")

	// 如果没有设置环境变量，尝试从这里读取（不推荐在生产环境中硬编码）
	if apiKey == "" {
		// TODO: 在这里设置您的OpenAI API Key
		apiKey = "your-openai-api-key-here"
	}

	if apiKey == "" || apiKey == "your-openai-api-key-here" {
		// 如果没有配置API Key，返回错误，使用本地回复
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	// 构建OpenAI API请求
	requestBody := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "system",
				Content: "你是一个友好的聊天室AI助手，请用中文回答用户关于聊天室功能的问题。保持回答简洁有用。",
			},
			{
				Role:    "user",
				Content: message,
			},
		},
		MaxTokens: 150,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// 发送请求到OpenAI API
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var openAIResp OpenAIResponse
	err = json.Unmarshal(body, &openAIResp)
	if err != nil {
		return "", err
	}

	if len(openAIResp.Choices) > 0 {
		return openAIResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from OpenAI")
}

// getLocalResponse 本地智能回复（作为AI服务的备用方案）
func getLocalResponse(message string) string {
	message = strings.ToLower(message)

	// 基于关键词的智能回复
	responses := map[string]string{
		"你好":    "你好！我是AI助手，很高兴为您服务。有什么我可以帮助您的吗？",
		"hello": "Hello! I'm your AI assistant. How can I help you today?",
		"功能":    "这个聊天室支持私聊、群聊、发送图片、语音消息、表情包等功能。您还可以创建群组、添加好友等。",
		"群聊":    "创建群聊很简单：1. 点击底部的\"群聊\"选项卡 2. 点击创建群按钮 3. 设置群名称和描述 4. 邀请好友加入",
		"好友":    "添加好友的方法：1. 点击联系人列表 2. 点击添加好友按钮 3. 输入对方的用户名或ID 4. 发送好友请求",
		"消息":    "系统支持多种消息类型：文字消息、图片、语音、表情包、文件传输等。",
		"帮助":    "我可以帮助您了解聊天室的各种功能。请告诉我您想了解什么，比如如何添加好友、创建群聊等。",
		"时间":    fmt.Sprintf("现在的时间是 %s", getCurrentTime()),
	}

	// 寻找匹配的关键词
	for keyword, response := range responses {
		if strings.Contains(message, keyword) {
			return response
		}
	}

	// 默认回复
	defaultResponses := []string{
		"这是一个很有趣的问题！让我想想...",
		"关于这个问题，我建议您可以尝试探索一下聊天室的各项功能。",
		"感谢您的提问！如果您需要帮助，可以随时问我关于聊天室功能的问题。",
		"我理解您的问题。您可以尝试使用聊天室的不同功能来获得更好的体验。",
		"很抱歉，我可能没有完全理解您的问题。您能详细说明一下吗？",
	}

	// 基于消息长度选择回复
	index := len(message) % len(defaultResponses)
	return defaultResponses[index]
}

// getCurrentTime 获取当前时间的格式化字符串
func getCurrentTime() string {
	return time.Now().Format("2006年01月02日 15:04")
}
