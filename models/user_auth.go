package models

import "time"

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应结构体
type LoginResponse struct {
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}

// UserClaims JWT 用户信息结构体
type UserClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}
