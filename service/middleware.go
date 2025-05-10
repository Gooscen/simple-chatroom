package service

import (
	"net/http"
	"simple-chatroom/models"
	"simple-chatroom/utils"
	"time"

	"github.com/gin-gonic/gin"
)

var jwtMiddleware *utils.GinJWTMiddleware

// InitJWT 初始化 JWT 中间件
func InitJWT() {
	jwtMiddleware = &utils.GinJWTMiddleware{
		Realm:          "gin jwt",
		Key:            []byte("your-secret-key"), // 请更改为你的密钥
		Timeout:        24 * time.Hour,            // token 有效期 24 小时
		MaxRefresh:     24 * time.Hour,            // 最大刷新时间 24 小时
		IdentityKey:    "user_id",
		TokenLookup:    "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:  "Bearer",
		TimeFunc:       time.Now,
		Authenticator:  authenticate,
		Authorizator:   authorize,
		Unauthorized:   unauthorized,
		LoginResponse:  loginResponse,
		LogoutResponse: logoutResponse,
	}

	// 初始化中间件
	_, err := utils.New(jwtMiddleware)
	if err != nil {
		panic(err)
	}
}

// authenticate 认证用户
func authenticate(c *gin.Context) (interface{}, error) {
	var loginReq models.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		return nil, err
	}

	// 这里应该调用你的用户验证逻辑
	// 示例中简单判断用户名和密码
	if loginReq.Username == "admin" && loginReq.Password == "password" {
		return &models.UserClaims{
			UserID:   1,
			Username: loginReq.Username,
		}, nil
	}

	return nil, utils.ErrFailedAuthentication
}

// authorize 授权用户
func authorize(data interface{}, c *gin.Context) bool {
	if _, ok := data.(*models.UserClaims); ok {
		return true
	}
	return false
}

// unauthorized 未授权处理
func unauthorized(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
	})
}

// loginResponse 登录响应
func loginResponse(c *gin.Context, code int, token string, expire time.Time) {
	c.JSON(http.StatusOK, models.LoginResponse{
		Token:  token,
		Expire: expire,
	})
}

// logoutResponse 登出响应
func logoutResponse(c *gin.Context, code int) {
	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": "logout success",
	})
}

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	return jwtMiddleware.MiddlewareFunc()
}

// LoginHandler 登录处理
func LoginHandler(c *gin.Context) {
	jwtMiddleware.LoginHandler(c)
}

// LogoutHandler 登出处理
func LogoutHandler(c *gin.Context) {
	jwtMiddleware.LogoutHandler(c)
}

// RefreshHandler 刷新 token 处理
func RefreshHandler(c *gin.Context) {
	jwtMiddleware.RefreshHandler(c)
}
