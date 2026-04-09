package handler

import (
	"strings"

	"github.com/dev-32/auth-service/internal/domain"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService domain.AuthService
}

func NewAuthHandler(authService domain.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}
	tokens, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		domainError(c, err)
		return
	}
	created(c, "registration successful", tokens)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}

	tokens, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		domainError(c, err)
		return
	}
	ok(c, "login successful", tokens)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		unauthorized(c, "token missing")
		return
	}

	if err := h.authService.Logout(c, token); err != nil {
		domainError(c, err)
		return
	}
	ok(c, "logged out successfully", nil)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req domain.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}

	tokens, err := h.authService.Refresh(c, req.RefreshToken)
	if err != nil {
		domainError(c, err)
		return
	}

	ok(c, "token refreshed", tokens)
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}
