package service

import (
	"simple-chatroom/models"
	"strconv"

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

// AI对话历史响应结构
type AIChatHistoryResponse struct {
	History []models.AIChatRecord `json:"history"`
	Code    int                   `json:"code"`
	Msg     string                `json:"msg"`
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

	// 调用models包中的AI服务，传递用户ID用于存储对话
	reply := models.GetAIResponseAndStore(request.Message, request.UserID)

	c.JSON(200, AIChatResponse{
		Reply: reply,
		Code:  0,
		Msg:   "success",
	})
}

// HandleAIChatHistory 获取AI对话历史
func HandleAIChatHistory(c *gin.Context) {
	userIDStr := c.Query("userId")
	if userIDStr == "" {
		c.JSON(400, AIChatHistoryResponse{
			Code: -1,
			Msg:  "缺少用户ID参数",
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(400, AIChatHistoryResponse{
			Code: -1,
			Msg:  "用户ID格式错误",
		})
		return
	}

	// 获取最近20条对话记录
	history := models.GetAIChatHistory(userID, 0, 19)

	c.JSON(200, AIChatHistoryResponse{
		History: history,
		Code:    0,
		Msg:     "success",
	})
}
