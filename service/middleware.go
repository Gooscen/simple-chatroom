package service

import (
	"simple-chatroom/models"

	"github.com/gin-gonic/gin"
)

func JWTAuth() func(c *gin.Context) {
	return models.JWTAuthMiddleware()
}
