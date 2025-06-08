package models

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID               int    `json:"user_id"`
	Username             string `json:"username"`
	jwt.RegisteredClaims        // v5版本新加的方法
}

func GenerateJWT(userID int, username, secretKey string) (string, error) {
	claims := UserClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 过期时间24小时
			IssuedAt:  jwt.NewNumericDate(time.Now()),                     // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                     // 生效时间
		},
	}
	// 使用HS256签名算法
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString([]byte(secretKey))

	return s, err
}

// 解析JWT
func ParseJwt(tokenstring, secretKey string) (*UserClaims, error) {
	t, err := jwt.ParseWithClaims(tokenstring, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if claims, ok := t.Claims.(*UserClaims); ok && t.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
}

func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 添加调试信息
		fmt.Printf("请求路径: %s\n", c.Request.URL.Path)
		fmt.Printf("请求方法: %s\n", c.Request.Method)

		//获取到请求头中的token
		authHeader := c.Request.Header.Get("Authorization")
		fmt.Printf("Authorization头: %s\n", authHeader)

		if authHeader == "" {
			fmt.Printf("错误: Authorization头为空\n")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "访问失败,请登录!",
				"data": nil,
			})
			c.Abort()
			return
		}

		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			fmt.Printf("错误: token格式不正确, parts: %v\n", parts)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "访问失败,无效的token格式,请登录!",
				"data": nil,
			})
			c.Abort()
			return
		}

		// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		fmt.Printf("尝试解析token: %s\n", parts[1])
		mc, err := ParseJwt(parts[1], "secretKey")
		if err != nil {
			fmt.Printf("错误: token解析失败: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "访问失败,无效的token,请登录!",
				"data": nil,
			})
			c.Abort()
			return
		}

		fmt.Printf("token解析成功, 用户: %s\n", mc.Username)
		// 将当前请求的userID信息保存到请求的上下文c上
		c.Set("userID", mc.UserID)
		c.Set("username", mc.Username)
		c.Next() // 后续的处理函数可以用过c.Get("username")来获取当前请求的用户信息
	}
}
