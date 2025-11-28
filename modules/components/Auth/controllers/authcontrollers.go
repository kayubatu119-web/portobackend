package controllers

import (
	"fmt"
	"gintugas/modules/components/Auth/blacklist"
	models "gintugas/modules/components/Auth/model"
	utils "gintugas/modules/components/Auth/util"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler untuk menangani autentikasi
type AuthHandler struct {
	DB *gorm.DB
}

type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginInput untuk binding input login
type LoginInput struct {
	Identifier string `json:"cek" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

// Register godoc
// @Summary Register user baru
// @Description Mendaftarkan user baru ke sistem
// @Tags auth
// @Accept json
// @Produce json
// @Param input body map[string]interface{} true "Data registrasi { \"username\": \"john_doe\", \"email\": \"john@example.com\", \"password\": \"password123\" }"
// @Success 201 {object} map[string]interface{} "Register success"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "User registered successfully",
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// Login godoc
// @Summary Login user
// @Description Melakukan login dan mendapatkan token JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param input body map[string]interface{} true "Kredensial login { \"cek\": \"john@example.com\", \"password\": \"password123\" }"
// @Success 200 {object} map[string]interface{} "Login success"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("username = ? OR email = ?", input.Identifier, input.Identifier).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	tokenString, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   tokenString,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// Logout godoc
// @Summary Logout user
// @Description Melakukan logout dan invalidate token
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Logout success"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Ambil token dari header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	// Validasi dan ambil claims dari token
	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	// Ambil expiry time dari claims
	if exp, ok := claims["exp"].(float64); ok {
		expiryTime := time.Unix(int64(exp), 0)

		// Dapatkan blacklist instance dan tambahkan token
		bl := blacklist.GetInstance()
		bl.Add(tokenString, expiryTime)

		// Debug log (bisa dihapus di production)
		fmt.Printf("[LOGOUT] Token added to blacklist. Total tokens in blacklist: %d\n", bl.GetCount())
		fmt.Printf("[LOGOUT] Token: %s\n", tokenString[:20]+"...")
		fmt.Printf("[LOGOUT] Expiry: %v\n", expiryTime)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

func GetJwtSecret() []byte {
	return utils.JwtSecret
}

// Verify godoc
// @Summary Verify token
// @Description Memverifikasi validitas token JWT
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Token valid"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /api/auth/verify [get]
func (h *AuthHandler) Verify(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	username, _ := c.Get("username")
	role, _ := c.Get("user_role")

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"user": gin.H{
			"id":       userID,
			"username": username,
			"role":     role,
		},
	})
}
