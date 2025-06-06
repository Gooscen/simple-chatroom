package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"simple-chatroom/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
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
	Model     string      `json:"model"`
	Messages  []AIMessage `json:"messages"`
	MaxTokens int         `json:"max_tokens"`
}

type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAI API响应结构
type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message AIMessage `json:"message"`
}

// AI对话记录结构
type AIChatRecord struct {
	UserMessage string    `json:"user_message"`
	AIReply     string    `json:"ai_reply"`
	Timestamp   time.Time `json:"timestamp"`
}

// GetAIResponse 对外提供的AI响应函数
func GetAIResponse(message string) string {
	// 首先尝试调用真实的AI API
	reply, err := getAIResponse(message)
	if err != nil {
		// 如果AI服务失败，使用本地智能回复
		reply = getLocalResponse(message)
	}
	return reply
}

// GetAIResponseAndStore 获取AI回复并存储到Redis
func GetAIResponseAndStore(message string, userID int) string {
	// 获取AI回复
	reply := GetAIResponse(message)

	// 存储对话到Redis
	storeAIChatToRedis(userID, message, reply)

	return reply
}

// GetAIChatHistory 获取用户的AI对话历史（返回结构化数据）
func GetAIChatHistory(userID int, start, end int64) []AIChatRecord {
	ctx := context.Background()
	chatKey := "ai_chat_" + strconv.Itoa(userID)

	// 从Redis获取对话记录（按时间倒序）
	records, err := utils.Red.ZRevRange(ctx, chatKey, start, end).Result()
	if err != nil {
		fmt.Println("获取AI对话历史失败:", err)
		return []AIChatRecord{}
	}

	// 解析JSON记录
	var chatHistory []AIChatRecord
	for _, recordStr := range records {
		var record AIChatRecord
		if err := json.Unmarshal([]byte(recordStr), &record); err == nil {
			chatHistory = append(chatHistory, record)
		}
	}

	fmt.Printf("获取用户%d的AI对话历史成功，记录数量: %d\n", userID, len(chatHistory))
	return chatHistory
}

// RedisAIMsg 获取AI对话缓存消息（参考message.go的RedisMsg函数）
func RedisAIMsg(userID int64, start int64, end int64, isRev bool) []string {
	ctx := context.Background()
	chatKey := "ai_chat_" + strconv.Itoa(int(userID))

	var rels []string
	var err error
	if isRev {
		rels, err = utils.Red.ZRange(ctx, chatKey, start, end).Result()
	} else {
		rels, err = utils.Red.ZRevRange(ctx, chatKey, start, end).Result()
	}
	if err != nil {
		fmt.Println("获取AI对话历史失败:", err)
	} else {
		fmt.Printf("获取用户%d的AI对话历史成功，消息数量: %d\n", userID, len(rels))
	}
	return rels
}

// storeAIChatToRedis 将AI对话存储到Redis中
func storeAIChatToRedis(userID int, userMessage, aiReply string) {
	ctx := context.Background()
	chatKey := "ai_chat_" + strconv.Itoa(userID)

	// 创建对话记录
	record := AIChatRecord{
		UserMessage: userMessage,
		AIReply:     aiReply,
		Timestamp:   time.Now(),
	}

	// 序列化为JSON
	recordJSON, err := json.Marshal(record)
	if err != nil {
		fmt.Println("AI对话记录序列化失败:", err)
		return
	}

	// 获取当前对话列表长度，用作score
	res, err := utils.Red.ZRevRange(ctx, chatKey, 0, -1).Result()
	if err != nil {
		fmt.Println("Redis ZRevRange error:", err)
	}
	score := float64(len(res)) + 1

	// 存储到Redis有序集合
	_, err = utils.Red.ZAdd(ctx, chatKey, &redis.Z{Score: score, Member: recordJSON}).Result()
	if err != nil {
		fmt.Println("AI对话存储到Redis失败:", err)
	} else {
		fmt.Printf("AI对话已存储到Redis: 用户%d\n", userID)
		// 设置3小时过期时间
		utils.Red.Expire(ctx, chatKey, 3*time.Hour)
	}
}

// getAIResponse 调用真实的AI API（如OpenAI）
func getAIResponse(message string) (string, error) {
	// 从viper读取AI配置
	apiKey := viper.GetString("ai.api_key")
	provider := viper.GetString("ai.provider")
	baseURL := viper.GetString("ai.base_url")
	model := viper.GetString("ai.model")
	maxTokens := viper.GetInt("ai.max_tokens")
	timeout := viper.GetInt("ai.timeout")

	// 设置默认值
	if provider == "" {
		provider = "deepseek"
	}
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	if model == "" {
		model = "deepseek-chat"
	}
	if maxTokens == 0 {
		maxTokens = 150
	}
	if timeout == 0 {
		timeout = 30
	}

	if apiKey == "" || apiKey == "your-api-key-here" {
		// 如果没有配置API Key，返回错误，使用本地回复
		return "", fmt.Errorf("AI API key not configured in config.yml")
	}

	// 构建AI API请求
	requestBody := OpenAIRequest{
		Model: model,
		Messages: []AIMessage{
			{
				Role:    "system",
				Content: "你是一个友好的聊天室AI助手，请用中文回答用户关于聊天室功能的问题。保持回答简洁有用。",
			},
			{
				Role:    "user",
				Content: message,
			},
		},
		MaxTokens: maxTokens,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// 构建请求URL
	requestURL := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(baseURL, "/"))

	// 发送请求到AI API
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
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
		return "", fmt.Errorf("AI API error: %s", string(body))
	}

	var aiResp OpenAIResponse
	err = json.Unmarshal(body, &aiResp)
	if err != nil {
		return "", err
	}

	if len(aiResp.Choices) > 0 {
		return aiResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from AI API")
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
