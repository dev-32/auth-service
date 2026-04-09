package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func ok(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func badRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error:   message,
	})
}

func unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error:   message,
	})
}

func forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error:   message,
	})
}

func notFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error:   message,
	})
}

func conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Success: false,
		Error:   message,
	})
}

func serverError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error:   message,
	})
}

// domainError maps domain errors to HTTP responses
func domainError(c *gin.Context, err error) {
	switch err.Error() {
	case "user not found":
		notFound(c, err.Error())
	case "email already exists":
		conflict(c, err.Error())
	case "invalid email or password":
		unauthorized(c, err.Error())
	case "invalid or expired token":
		unauthorized(c, err.Error())
	case "token has been revoked":
		unauthorized(c, err.Error())
	case "unauthorized":
		unauthorized(c, err.Error())
	case "forbidden":
		forbidden(c, err.Error())
	case "oauth authentication failed":
		badRequest(c, err.Error())
	default:
		serverError(c, "something went wrong")
	}
}
