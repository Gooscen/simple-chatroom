package service

import (
	"simple-chatroom/models"

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

	// 调用models包中的AI服务
	reply := models.GetAIResponse(request.Message)

	c.JSON(200, AIChatResponse{
		Reply: reply,
		Code:  0,
		Msg:   "success",
	})
}
