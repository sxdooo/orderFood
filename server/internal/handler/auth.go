package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/orderfood/server/internal/middleware"
	"github.com/orderfood/server/internal/pkg/response"
	"github.com/orderfood/server/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type wechatLoginRequest struct {
	Code      string `json:"code"      binding:"required"`
	PhoneCode string `json:"phoneCode"` // optional: from getPhoneNumber button
}

func (h *AuthHandler) WeChatLogin(c *gin.Context) {
	var req wechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	user, token, phone, err := h.auth.WeChatLogin(c.Request.Context(), req.Code, req.PhoneCode)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, gin.H{
		"token": token,
		"phone": phone, // resolved phone number; empty string in dev mode
		"user": gin.H{
			"id":   user.ID,
			"role": user.Role,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		response.BadRequest(c, "missing token")
		return
	}
	token := strings.TrimPrefix(header, "Bearer ")
	if err := h.auth.Logout(c.Request.Context(), token); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *AuthHandler) Me(c *gin.Context) {
	user := middleware.GetUser(c)
	response.OK(c, gin.H{
		"id":   user.ID,
		"role": user.Role,
	})
}
