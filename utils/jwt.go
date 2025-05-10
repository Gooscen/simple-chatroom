package utils

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/youmark/pkcs8"
)

type MapClaims jwt.MapClaims

type GinJWTMiddleware struct {
	// Realm 名称，用于显示给用户。必填项。
	Realm string

	// 签名算法 - 可选值有 HS256, HS384, HS512, RS256, RS384 或 RS512
	// 可选，默认为 HS256。
	SigningAlgorithm string

	// 用于签名的密钥。必填项。
	Key []byte

	// 用于获取签名密钥的回调函数。设置 KeyFunc 将绕过
	// 所有其他密钥设置
	KeyFunc func(token *jwt.Token) (interface{}, error)

	// JWT token 的有效期。可选，默认为一小时。
	Timeout time.Duration
	// 用于覆盖默认超时时长的回调函数。
	TimeoutFunc func(data interface{}) time.Duration

	// 此字段允许客户端在 MaxRefresh 时间过去之前刷新他们的 token。
	// 注意，客户端可以在 MaxRefresh 的最后时刻刷新他们的 token。
	// 这意味着 token 的最大有效时间跨度是 TokenTime + MaxRefresh。
	// 可选，默认为 0，表示不可刷新。
	MaxRefresh time.Duration

	// 用于基于登录信息执行用户认证的回调函数。
	// 必须返回用户数据作为用户标识符，它将被存储在 Claim 数组中。必填项。
	// 检查错误 (e) 以确定适当的错误消息。
	Authenticator func(c *gin.Context) (interface{}, error)

	// 用于执行已认证用户授权的回调函数。仅在
	// 认证成功后调用。成功时返回 true，失败时返回 false。
	// 可选，默认为成功。
	Authorizator func(data interface{}, c *gin.Context) bool

	// 在登录期间调用的回调函数。
	// 使用此函数可以向 web token 添加额外的负载数据。
	// 然后可以通过 c.Get("JWT_PAYLOAD") 在请求期间使用这些数据。
	// 注意，负载数据不会被加密。
	// jwt.io 上提到的属性不能用作 map 的键。
	// 可选，默认不会设置额外的数据。
	PayloadFunc func(data interface{}) MapClaims

	// 用户可以定义自己的未授权处理函数。
	Unauthorized func(c *gin.Context, code int, message string)

	// 用户可以定义自己的登录响应函数。
	LoginResponse func(c *gin.Context, code int, message string, time time.Time)

	// 用户可以定义自己的登出响应函数。
	LogoutResponse func(c *gin.Context, code int)

	// 用户可以定义自己的刷新响应函数。
	RefreshResponse func(c *gin.Context, code int, message string, time time.Time)

	// 设置身份处理函数
	IdentityHandler func(*gin.Context) interface{}

	// 设置身份键
	IdentityKey string

	// TokenLookup 是一个字符串，格式为 "<source>:<name>"，用于
	// 从请求中提取 token。
	// 可选。默认值为 "header:Authorization"。
	// 可能的值：
	// - "header:<name>"
	// - "query:<name>"
	// - "cookie:<name>"
	TokenLookup string

	// TokenHeadName 是 header 中的字符串。默认值为 "Bearer"
	TokenHeadName string

	// TimeFunc 提供当前时间。你可以覆盖它以使用其他时间值。这对于测试或如果你的服务器使用与你的 token 不同的时区很有用。
	TimeFunc func() time.Time

	// JWT 中间件失败时的 HTTP 状态消息。
	// 检查错误 (e) 以确定适当的错误消息。
	HTTPStatusMessageFunc func(e error, c *gin.Context) string

	// 非对称算法的私钥文件
	PrivKeyFile string

	// 非对称算法的私钥字节
	//
	// 注意：如果同时设置了 PrivKeyFile 和 PrivKeyBytes，PrivKeyFile 优先
	PrivKeyBytes []byte

	// 非对称算法的公钥文件
	PubKeyFile string

	// 私钥密码
	PrivateKeyPassphrase string

	// 非对称算法的公钥字节。
	//
	// 注意：如果同时设置了 PubKeyFile 和 PubKeyBytes，PubKeyFile 优先
	PubKeyBytes []byte

	// 私钥
	privKey *rsa.PrivateKey

	// 公钥
	pubKey *rsa.PublicKey

	// 可选择将 token 作为 cookie 返回
	SendCookie bool

	// cookie 的有效期。可选，默认等于 Timeout 值。
	CookieMaxAge time.Duration

	// 允许在开发环境中通过 http 使用不安全的 cookie
	SecureCookie bool

	// 允许在开发环境中从客户端访问 cookie
	CookieHTTPOnly bool

	// 允许在开发环境中更改 cookie 域
	CookieDomain string

	// SendAuthorization 允许为每个请求返回授权头
	SendAuthorization bool

	// 禁用 context 的 abort()
	DisabledAbort bool

	// CookieName 允许在开发环境中更改 cookie 名称
	CookieName string

	// CookieSameSite 允许使用 http.SameSite cookie 参数
	CookieSameSite http.SameSite

	// ParseOptions 允许修改 jwt 的解析方法
	ParseOptions []jwt.ParserOption

	// 默认值为 "exp"
	ExpField string
}

var (
	// ErrMissingSecretKey 表示需要密钥
	ErrMissingSecretKey = errors.New("secret key is required")

	// ErrForbidden 当 HTTP 状态码为 403 时
	ErrForbidden = errors.New("you don't have permission to access this resource")

	// ErrMissingAuthenticatorFunc 表示需要认证器函数
	ErrMissingAuthenticatorFunc = errors.New("ginJWTMiddleware.Authenticator func is undefined")

	// ErrMissingLoginValues 表示用户尝试在没有用户名或密码的情况下进行认证
	ErrMissingLoginValues = errors.New("missing Username or Password")

	// ErrFailedAuthentication 表示认证失败，可能是用户名或密码错误
	ErrFailedAuthentication = errors.New("incorrect Username or Password")

	// ErrFailedTokenCreation 表示 JWT Token 创建失败，原因未知
	ErrFailedTokenCreation = errors.New("failed to create JWT Token")

	// ErrExpiredToken 表示 JWT token 已过期。无法刷新。
	ErrExpiredToken = errors.New("token is expired")

	// ErrEmptyAuthHeader 如果使用 HTTP header 进行认证，但 Auth header 为空时抛出
	ErrEmptyAuthHeader = errors.New("auth header is empty")

	// ErrMissingExpField token 中缺少 exp 字段
	ErrMissingExpField = errors.New("missing exp field")

	// ErrWrongFormatOfExp exp 字段必须是 float64 格式
	ErrWrongFormatOfExp = errors.New("exp must be float64 format")

	// ErrInvalidAuthHeader 表示 auth header 无效，例如可能使用了错误的 Realm 名称
	ErrInvalidAuthHeader = errors.New("auth header is invalid")

	// ErrEmptyQueryToken 如果使用 URL Query 进行认证，但查询 token 变量为空时抛出
	ErrEmptyQueryToken = errors.New("query token is empty")

	// ErrEmptyCookieToken 如果使用 cookie 进行认证，但 token cookie 为空时抛出
	ErrEmptyCookieToken = errors.New("cookie token is empty")

	// ErrEmptyParamToken 如果使用路径参数进行认证，但路径中的参数为空时抛出
	ErrEmptyParamToken = errors.New("parameter token is empty")

	// ErrInvalidSigningAlgorithm 表示签名算法无效，需要是 HS256, HS384, HS512, RS256, RS384 或 RS512
	ErrInvalidSigningAlgorithm = errors.New("invalid signing algorithm")

	// ErrNoPrivKeyFile 表示给定的私钥文件无法读取
	ErrNoPrivKeyFile = errors.New("private key file unreadable")

	// ErrNoPubKeyFile 表示给定的公钥文件无法读取
	ErrNoPubKeyFile = errors.New("public key file unreadable")

	// ErrInvalidPrivKey 表示给定的私钥无效
	ErrInvalidPrivKey = errors.New("private key invalid")

	// ErrInvalidPubKey 表示给定的公钥无效
	ErrInvalidPubKey = errors.New("public key invalid")

	// IdentityKey 默认身份键
	IdentityKey = "identity"
)

// New 用于检查 GinJWTMiddleware 的错误
func New(m *GinJWTMiddleware) (*GinJWTMiddleware, error) {
	if err := m.MiddlewareInit(); err != nil {
		return nil, err
	}
	return m, nil
}

// readKeys 读取密钥
func (mw *GinJWTMiddleware) readKeys() error {
	err := mw.privateKey()
	if err != nil {
		return err
	}
	err = mw.publicKey()
	if err != nil {
		return err
	}
	return nil
}

// privateKey 处理私钥
func (mw *GinJWTMiddleware) privateKey() error {
	var keyData []byte
	var err error
	if mw.PrivKeyFile == "" {
		keyData = mw.PrivKeyBytes
	} else {
		var filecontent []byte
		filecontent, err = os.ReadFile(mw.PrivKeyFile)
		if err != nil {
			return ErrNoPrivKeyFile
		}
		keyData = filecontent
	}
	if mw.PrivateKeyPassphrase != "" {
		var key interface{}
		key, err = pkcs8.ParsePKCS8PrivateKey(keyData, []byte(mw.PrivateKeyPassphrase))
		if err != nil {
			return ErrInvalidPrivKey
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return ErrInvalidPrivKey
		}
		mw.privKey = rsaKey
		return nil
	}
	var key *rsa.PrivateKey
	key, err = jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return ErrInvalidPrivKey
	}
	mw.privKey = key
	return nil
}

// publicKey 处理公钥
func (mw *GinJWTMiddleware) publicKey() error {
	var keyData []byte
	if mw.PubKeyFile == "" {
		keyData = mw.PubKeyBytes
	} else {
		filecontent, err := os.ReadFile(mw.PubKeyFile)
		if err != nil {
			return ErrNoPubKeyFile
		}
		keyData = filecontent
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return ErrInvalidPubKey
	}
	mw.pubKey = key
	return nil
}

// usingPublicKeyAlgo 检查是否使用公钥算法
func (mw *GinJWTMiddleware) usingPublicKeyAlgo() bool {
	switch mw.SigningAlgorithm {
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

// MiddlewareInit 初始化 jwt 配置
func (mw *GinJWTMiddleware) MiddlewareInit() error {
	if mw.TokenLookup == "" {
		mw.TokenLookup = "header:Authorization"
	}
	if mw.SigningAlgorithm == "" {
		mw.SigningAlgorithm = "HS256"
	}
	if mw.Timeout == 0 {
		mw.Timeout = time.Hour
	}
	if mw.TimeoutFunc == nil {
		mw.TimeoutFunc = func(data interface{}) time.Duration {
			return mw.Timeout
		}
	}
	if mw.TimeFunc == nil {
		mw.TimeFunc = time.Now
	}
	mw.TokenHeadName = strings.TrimSpace(mw.TokenHeadName)
	if len(mw.TokenHeadName) == 0 {
		mw.TokenHeadName = "Bearer"
	}
	if mw.Authorizator == nil {
		mw.Authorizator = func(data interface{}, c *gin.Context) bool {
			return true
		}
	}
	if mw.Unauthorized == nil {
		mw.Unauthorized = func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		}
	}
	if mw.LoginResponse == nil {
		mw.LoginResponse = func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(http.StatusOK, gin.H{
				"code":   http.StatusOK,
				"token":  token,
				"expire": expire.Format(time.RFC3339),
			})
		}
	}
	if mw.LogoutResponse == nil {
		mw.LogoutResponse = func(c *gin.Context, code int) {
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
			})
		}
	}
	if mw.RefreshResponse == nil {
		mw.RefreshResponse = func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(http.StatusOK, gin.H{
				"code":   http.StatusOK,
				"token":  token,
				"expire": expire.Format(time.RFC3339),
			})
		}
	}
	if mw.IdentityKey == "" {
		mw.IdentityKey = IdentityKey
	}
	if mw.IdentityHandler == nil {
		mw.IdentityHandler = func(c *gin.Context) interface{} {
			claims := ExtractClaims(c)
			return claims[mw.IdentityKey]
		}
	}
	if mw.HTTPStatusMessageFunc == nil {
		mw.HTTPStatusMessageFunc = func(e error, c *gin.Context) string {
			return e.Error()
		}
	}
	if mw.Realm == "" {
		mw.Realm = "gin jwt"
	}
	if mw.CookieMaxAge == 0 {
		mw.CookieMaxAge = mw.Timeout
	}
	if mw.CookieName == "" {
		mw.CookieName = "jwt"
	}
	if mw.ExpField == "" {
		mw.ExpField = "exp"
	}
	if mw.KeyFunc != nil {
		return nil
	}
	if mw.usingPublicKeyAlgo() {
		return mw.readKeys()
	}
	if mw.Key == nil {
		return ErrMissingSecretKey
	}
	return nil
}

// MiddlewareFunc 使 GinJWTMiddleware 实现 Middleware 接口
func (mw *GinJWTMiddleware) MiddlewareFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		mw.middlewareImpl(c)
	}
}

// middlewareImpl 中间件实现
func (mw *GinJWTMiddleware) middlewareImpl(c *gin.Context) {
	claims, err := mw.GetClaimsFromJWT(c)
	if err != nil {
		mw.unauthorized(c, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(err, c))
		return
	}
	switch v := claims[mw.ExpField].(type) {
	case nil:
		mw.unauthorized(c, http.StatusBadRequest, mw.HTTPStatusMessageFunc(ErrMissingExpField, c))
		return
	case float64:
		if int64(v) < mw.TimeFunc().Unix() {
			mw.unauthorized(c, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(ErrExpiredToken, c))
			return
		}
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			mw.unauthorized(c, http.StatusBadRequest, mw.HTTPStatusMessageFunc(ErrWrongFormatOfExp, c))
			return
		}
		if n < mw.TimeFunc().Unix() {
			mw.unauthorized(c, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(ErrExpiredToken, c))
			return
		}
	default:
		mw.unauthorized(c, http.StatusBadRequest, mw.HTTPStatusMessageFunc(ErrWrongFormatOfExp, c))
		return
	}
	c.Set("JWT_PAYLOAD", claims)
	identity := mw.IdentityHandler(c)
	if identity != nil {
		c.Set(mw.IdentityKey, identity)
	}
	if !mw.Authorizator(identity, c) {
		mw.unauthorized(c, http.StatusForbidden, mw.HTTPStatusMessageFunc(ErrForbidden, c))
		return
	}
	c.Next()
}

// GetClaimsFromJWT 从 JWT token 中获取 claims
func (mw *GinJWTMiddleware) GetClaimsFromJWT(c *gin.Context) (MapClaims, error) {
	token, err := mw.ParseToken(c)
	if err != nil {
		return nil, err
	}
	if mw.SendAuthorization {
		if v, ok := c.Get("JWT_TOKEN"); ok {
			c.Header("Authorization", mw.TokenHeadName+" "+v.(string))
		}
	}
	claims := MapClaims{}
	for key, value := range token.Claims.(jwt.MapClaims) {
		claims[key] = value
	}
	return claims, nil
}

// LoginHandler 可以被客户端用来获取 jwt token
func (mw *GinJWTMiddleware) LoginHandler(c *gin.Context) {
	if mw.Authenticator == nil {
		mw.unauthorized(c, http.StatusInternalServerError, mw.HTTPStatusMessageFunc(ErrMissingAuthenticatorFunc, c))
		return
	}
	data, err := mw.Authenticator(c)
	if err != nil {
		mw.unauthorized(c, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(err, c))
		return
	}
	token := jwt.New(jwt.GetSigningMethod(mw.SigningAlgorithm))
	claims := token.Claims.(jwt.MapClaims)
	if mw.PayloadFunc != nil {
		for key, value := range mw.PayloadFunc(data) {
			claims[key] = value
		}
	}
	expire := mw.TimeFunc().Add(mw.TimeoutFunc(claims))
	claims[mw.ExpField] = expire.Unix()
	claims["orig_iat"] = mw.TimeFunc().Unix()
	tokenString, err := mw.signedString(token)
	if err != nil {
		mw.unauthorized(c, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(ErrFailedTokenCreation, c))
		return
	}
	mw.SetCookie(c, tokenString)
	mw.LoginResponse(c, http.StatusOK, tokenString, expire)
}

// LogoutHandler 可以被客户端用来移除 jwt cookie（如果设置了的话）
func (mw *GinJWTMiddleware) LogoutHandler(c *gin.Context) {
	if mw.SendCookie {
		if mw.CookieSameSite != 0 {
			c.SetSameSite(mw.CookieSameSite)
		}
		c.SetCookie(
			mw.CookieName,
			"",
			-1,
			"/",
			mw.CookieDomain,
			mw.SecureCookie,
			mw.CookieHTTPOnly,
		)
	}
	mw.LogoutResponse(c, http.StatusOK)
}

// signedString 签名 token
func (mw *GinJWTMiddleware) signedString(token *jwt.Token) (string, error) {
	var tokenString string
	var err error
	if mw.usingPublicKeyAlgo() {
		tokenString, err = token.SignedString(mw.privKey)
	} else {
		tokenString, err = token.SignedString(mw.Key)
	}
	return tokenString, err
}

// RefreshHandler 可以用来刷新 token。token 在刷新时仍然需要有效。
func (mw *GinJWTMiddleware) RefreshHandler(c *gin.Context) {
	tokenString, expire, err := mw.RefreshToken(c)
	if err != nil {
		mw.unauthorized(c, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(err, c))
		return
	}
	mw.RefreshResponse(c, http.StatusOK, tokenString, expire)
}

// RefreshToken 刷新 token 并检查 token 是否过期
func (mw *GinJWTMiddleware) RefreshToken(c *gin.Context) (string, time.Time, error) {
	claims, err := mw.CheckIfTokenExpire(c)
	if err != nil {
		return "", time.Now(), err
	}
	newToken := jwt.New(jwt.GetSigningMethod(mw.SigningAlgorithm))
	newClaims := newToken.Claims.(jwt.MapClaims)
	for key := range claims {
		newClaims[key] = claims[key]
	}
	expire := mw.TimeFunc().Add(mw.TimeoutFunc(claims))
	newClaims[mw.ExpField] = expire.Unix()
	newClaims["orig_iat"] = mw.TimeFunc().Unix()
	tokenString, err := mw.signedString(newToken)
	if err != nil {
		return "", time.Now(), err
	}
	mw.SetCookie(c, tokenString)
	return tokenString, expire, nil
}

// CheckIfTokenExpire 检查 token 是否过期
func (mw *GinJWTMiddleware) CheckIfTokenExpire(c *gin.Context) (jwt.MapClaims, error) {
	token, err := mw.ParseToken(c)
	if err != nil {
		validationErr, ok := err.(*jwt.ValidationError)
		if !ok || validationErr.Errors != jwt.ValidationErrorExpired {
			return nil, err
		}
	}
	claims := token.Claims.(jwt.MapClaims)
	origIat := int64(claims["orig_iat"].(float64))
	if origIat < mw.TimeFunc().Add(-mw.MaxRefresh).Unix() {
		return nil, ErrExpiredToken
	}
	return claims, nil
}

// TokenGenerator 客户端可以用来获取 jwt token 的方法
func (mw *GinJWTMiddleware) TokenGenerator(data interface{}) (string, time.Time, error) {
	token := jwt.New(jwt.GetSigningMethod(mw.SigningAlgorithm))
	claims := token.Claims.(jwt.MapClaims)
	if mw.PayloadFunc != nil {
		for key, value := range mw.PayloadFunc(data) {
			claims[key] = value
		}
	}
	expire := mw.TimeFunc().Add(mw.TimeoutFunc(claims))
	claims[mw.ExpField] = expire.Unix()
	claims["orig_iat"] = mw.TimeFunc().Unix()
	tokenString, err := mw.signedString(token)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expire, nil
}

// jwtFromHeader 从 header 中获取 jwt token
func (mw *GinJWTMiddleware) jwtFromHeader(c *gin.Context, key string) (string, error) {
	authHeader := c.Request.Header.Get(key)
	if authHeader == "" {
		return "", ErrEmptyAuthHeader
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == mw.TokenHeadName) {
		return "", ErrInvalidAuthHeader
	}
	return parts[1], nil
}

// jwtFromQuery 从查询参数中获取 jwt token
func (mw *GinJWTMiddleware) jwtFromQuery(c *gin.Context, key string) (string, error) {
	token := c.Query(key)
	if token == "" {
		return "", ErrEmptyQueryToken
	}
	return token, nil
}

// jwtFromCookie 从 cookie 中获取 jwt token
func (mw *GinJWTMiddleware) jwtFromCookie(c *gin.Context, key string) (string, error) {
	cookie, _ := c.Cookie(key)
	if cookie == "" {
		return "", ErrEmptyCookieToken
	}
	return cookie, nil
}

// jwtFromParam 从路径参数中获取 jwt token
func (mw *GinJWTMiddleware) jwtFromParam(c *gin.Context, key string) (string, error) {
	token := c.Param(key)
	if token == "" {
		return "", ErrEmptyParamToken
	}
	return token, nil
}

// jwtFromForm 从表单中获取 jwt token
func (mw *GinJWTMiddleware) jwtFromForm(c *gin.Context, key string) (string, error) {
	token := c.PostForm(key)
	if token == "" {
		return "", ErrEmptyParamToken
	}
	return token, nil
}

// ParseToken 从 gin context 中解析 jwt token
func (mw *GinJWTMiddleware) ParseToken(c *gin.Context) (*jwt.Token, error) {
	var token string
	var err error
	methods := strings.Split(mw.TokenLookup, ",")
	for _, method := range methods {
		if len(token) > 0 {
			break
		}
		parts := strings.Split(strings.TrimSpace(method), ":")
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		switch k {
		case "header":
			token, err = mw.jwtFromHeader(c, v)
		case "query":
			token, err = mw.jwtFromQuery(c, v)
		case "cookie":
			token, err = mw.jwtFromCookie(c, v)
		case "param":
			token, err = mw.jwtFromParam(c, v)
		case "form":
			token, err = mw.jwtFromForm(c, v)
		}
	}
	if err != nil {
		return nil, err
	}
	if mw.KeyFunc != nil {
		return jwt.Parse(token, mw.KeyFunc, mw.ParseOptions...)
	}
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(mw.SigningAlgorithm) != t.Method {
			return nil, ErrInvalidSigningAlgorithm
		}
		if mw.usingPublicKeyAlgo() {
			return mw.pubKey, nil
		}
		c.Set("JWT_TOKEN", token)
		return mw.Key, nil
	}, mw.ParseOptions...)
}

// ParseTokenString 解析 jwt token 字符串
func (mw *GinJWTMiddleware) ParseTokenString(token string) (*jwt.Token, error) {
	if mw.KeyFunc != nil {
		return jwt.Parse(token, mw.KeyFunc, mw.ParseOptions...)
	}
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(mw.SigningAlgorithm) != t.Method {
			return nil, ErrInvalidSigningAlgorithm
		}
		if mw.usingPublicKeyAlgo() {
			return mw.pubKey, nil
		}
		return mw.Key, nil
	}, mw.ParseOptions...)
}

// unauthorized 处理未授权的情况
func (mw *GinJWTMiddleware) unauthorized(c *gin.Context, code int, message string) {
	c.Header("WWW-Authenticate", "JWT realm="+mw.Realm)
	if !mw.DisabledAbort {
		c.Abort()
	}
	mw.Unauthorized(c, code, message)
}

// ExtractClaims 帮助提取 JWT claims
func ExtractClaims(c *gin.Context) MapClaims {
	claims, exists := c.Get("JWT_PAYLOAD")
	if !exists {
		return make(MapClaims)
	}
	return claims.(MapClaims)
}

// ExtractClaimsFromToken 帮助从 token 中提取 JWT claims
func ExtractClaimsFromToken(token *jwt.Token) MapClaims {
	if token == nil {
		return make(MapClaims)
	}
	claims := MapClaims{}
	for key, value := range token.Claims.(jwt.MapClaims) {
		claims[key] = value
	}
	return claims
}

// GetToken 帮助获取 JWT token 字符串
func GetToken(c *gin.Context) string {
	token, exists := c.Get("JWT_TOKEN")
	if !exists {
		return ""
	}
	return token.(string)
}

// SetCookie 帮助在 cookie 中设置 token
func (mw *GinJWTMiddleware) SetCookie(c *gin.Context, token string) {
	if mw.SendCookie {
		expireCookie := mw.TimeFunc().Add(mw.CookieMaxAge)
		maxage := int(expireCookie.Unix() - mw.TimeFunc().Unix())
		if mw.CookieSameSite != 0 {
			c.SetSameSite(mw.CookieSameSite)
		}
		c.SetCookie(
			mw.CookieName,
			token,
			maxage,
			"/",
			mw.CookieDomain,
			mw.SecureCookie,
			mw.CookieHTTPOnly,
		)
	}
}
