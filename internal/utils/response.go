package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Success bool        `json:"success"`
	Timestamp int64     `json:"timestamp"`
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      http.StatusOK,
		Data:      data,
		Success:   true,
		Timestamp: time.Now().Unix(),
	})
}

// SuccessResponseWithMessage 带消息的成功响应
func SuccessResponseWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, Response{
		Code:      http.StatusOK,
		Data:      data,
		Message:   message,
		Success:   true,
		Timestamp: time.Now().Unix(),
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:      code,
		Error:     message,
		Success:   false,
		Timestamp: time.Now().Unix(),
	})
}

// ErrorResponseWithData 带数据的错误响应
func ErrorResponseWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, Response{
		Code:      code,
		Data:      data,
		Error:     message,
		Success:   false,
		Timestamp: time.Now().Unix(),
	})
}

// ValidationError 参数验证错误响应
func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:      http.StatusBadRequest,
		Error:     message,
		Message:   "Validation failed",
		Success:   false,
		Timestamp: time.Now().Unix(),
	})
}

// NotFoundResponse 资源未找到响应
func NotFoundResponse(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:      http.StatusNotFound,
		Error:     message,
		Success:   false,
		Timestamp: time.Now().Unix(),
	})
}

// UnauthorizedResponse 未授权响应
func UnauthorizedResponse(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:      http.StatusUnauthorized,
		Error:     message,
		Success:   false,
		Timestamp: time.Now().Unix(),
	})
}

// ForbiddenResponse 禁止访问响应
func ForbiddenResponse(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:      http.StatusForbidden,
		Error:     message,
		Success:   false,
		Timestamp: time.Now().Unix(),
	})
}

// BadRequestResponse 请求错误响应
func BadRequestResponse(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:      http.StatusBadRequest,
		Error:     message,
		Success:   false,
		Timestamp: time.Now().Unix(),
	})
}

// InternalServerErrorResponse 内部服务器错误响应
func InternalServerErrorResponse(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:      http.StatusInternalServerError,
		Error:     message,
		Success:   false,
		Timestamp: time.Now().Unix(),
	})
}