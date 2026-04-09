package handler

import (
	"github.com/dev-32/auth-service/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	authService domain.AuthService
}

func NewUserHandler(authService domain.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userId, err := getUserID(c)
	if err != nil {
		unauthorized(c, "invalid user id")
		return
	}

	user, err := h.authService.GetUserByID(c, userId)
	if err != nil {
		domainError(c, err)
		return
	}

	ok(c, "user fetched", user.ToResponse())
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	userId, err := getUserID(c)
	if err != nil {
		unauthorized(c, "invalid user id")
		return
	}

	var req domain.UpdateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}

	user, err := h.authService.UpdateProfile(c, userId, req)
	if err != nil {
		domainError(c, err)
		return
	}

	ok(c, "user updated successfully", user.ToResponse())
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	idStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, domain.ErrUnauthorized
	}
	return uuid.Parse(idStr.(string))
}
