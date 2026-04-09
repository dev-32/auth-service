package handler

import (
	"strconv"

	"github.com/dev-32/auth-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, err := h.adminService.ListUsers(c, limit, offset)
	if err != nil {
		serverError(c, "failed to fetch users")
		return
	}

	responses := make([]any, len(users))

	for i := range users {
		responses[i] = users[i].ToResponse()
	}

	ok(c, "users fetched", gin.H{
		"users":  responses,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		badRequest(c, "invalid user id")
		return
	}

	if err := h.adminService.DeleteUser(c, id); err != nil {
		domainError(c, err)
		return
	}

	ok(c, "user deleted", nil)

}
