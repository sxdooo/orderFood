package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/orderfood/server/internal/middleware"
	"github.com/orderfood/server/internal/pkg/response"
	"github.com/orderfood/server/internal/service"
)

type BuyerProfileHandler struct {
	svc *service.BuyerProfileService
}

func NewBuyerProfileHandler(svc *service.BuyerProfileService) *BuyerProfileHandler {
	return &BuyerProfileHandler{svc: svc}
}

func (h *BuyerProfileHandler) Get(c *gin.Context) {
	user := middleware.GetUser(c)
	profile, err := h.svc.GetByUserID(c.Request.Context(), user.ID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, profile)
}

func (h *BuyerProfileHandler) Update(c *gin.Context) {
	user := middleware.GetUser(c)
	var req service.BuyerProfileInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	profile, err := h.svc.Upsert(c.Request.Context(), user.ID, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, profile)
}
