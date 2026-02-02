package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse sends a JSON error response
func ErrorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": message,
	})
}

// SuccessResponse sends a JSON success response
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

// CreatedResponse sends a 201 created response
func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{
		"data": data,
	})
}
