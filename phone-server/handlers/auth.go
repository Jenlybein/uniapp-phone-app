package handlers

import (
	"net/http"
	"phone-server/models"
	"phone-server/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler 认证接口处理器
type AuthHandler struct {
	db        *gorm.DB // 数据库连接
	jwtSecret string   // JWT密钥
	jwtExpire int      // JWT过期时间（小时）
}

// NewAuthHandler 创建认证接口处理器实例
func NewAuthHandler(db *gorm.DB, jwtSecret string, jwtExpire int) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
		jwtExpire: jwtExpire,
	}
}

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest 刷新Token请求参数
type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// Register 处理用户注册请求
// @Summary 用户注册
// @Description 创建新用户
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 200 {object} gin.H{"code": 200, "message": "注册成功", "token": ""}
// @Failure 400 {object} gin.H{"code": 400, "message": "请求参数错误"}
// @Failure 500 {object} gin.H{"code": 500, "message": "服务器内部错误"}
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if result := h.db.Where("username = ?", req.Username).First(&existingUser); result.RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "用户名已存在"})
		return
	}

	// 检查邮箱是否已存在
	if result := h.db.Where("email = ?", req.Email).First(&existingUser); result.RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "邮箱已存在"})
		return
	}

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": "密码加密失败"})
		return
	}

	// 创建新用户
	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if result := h.db.Create(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": "用户创建失败: " + result.Error.Error()})
		return
	}

	// 生成JWT令牌
	token, err := utils.GenerateToken(user.ID, user.Username, h.jwtSecret, h.jwtExpire)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": "令牌生成失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "注册成功", "token": token})
}

// Login 处理用户登录请求
// @Summary 用户登录
// @Description 用户登录并获取JWT令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} gin.H{"code": 200, "message": "登录成功", "token": "", "user": {}}
// @Failure 400 {object} gin.H{"code": 400, "message": "请求参数错误"}
// @Failure 401 {object} gin.H{"code": 401, "message": "用户名或密码错误"}
// @Failure 500 {object} gin.H{"code": 500, "message": "服务器内部错误"}
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 查找用户
	var user models.User
	if result := h.db.Where("username = ?", req.Username).First(&user); result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "用户名或密码错误"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "用户名或密码错误"})
		return
	}

	// 生成JWT令牌
	token, err := utils.GenerateToken(user.ID, user.Username, h.jwtSecret, h.jwtExpire)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": "令牌生成失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "登录成功",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// RefreshToken 处理刷新Token请求
// @Summary 刷新Token
// @Description 使用旧Token获取新Token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新Token请求"
// @Success 200 {object} gin.H{"code": 200, "message": "刷新成功", "token": ""}
// @Failure 400 {object} gin.H{"code": 400, "message": "请求参数错误"}
// @Failure 401 {object} gin.H{"code": 401, "message": "无效的Token"}
// @Failure 500 {object} gin.H{"code": 500, "message": "服务器内部错误"}
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 解析旧Token
	claims, err := utils.ParseToken(req.Token, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "无效的Token: " + err.Error()})
		return
	}

	// 生成新Token
	newToken, err := utils.GenerateToken(claims.UserID, claims.Username, h.jwtSecret, h.jwtExpire)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": "令牌生成失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "刷新成功", "token": newToken})
}

// AuthMiddleware JWT认证中间件
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization头获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "缺少Authorization头"})
			c.Abort()
			return
		}

		// 解析Token
		claims, err := utils.ParseToken(authHeader, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "无效的Token: " + err.Error()})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
