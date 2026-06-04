package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/orderfood/server/internal/middleware"
	"github.com/orderfood/server/internal/model"
	"github.com/orderfood/server/internal/pkg/amap"
	"github.com/orderfood/server/internal/pkg/response"
	"github.com/orderfood/server/internal/service"
)

type MenuHandler struct {
	menu *service.MenuService
}

func NewMenuHandler(menu *service.MenuService) *MenuHandler {
	return &MenuHandler{menu: menu}
}

func (h *MenuHandler) Tomorrow(c *gin.Context) {
	menu, err := h.menu.GetTomorrowPublished(c.Request.Context())
	if err != nil {
		response.OK(c, nil)
		return
	}
	response.OK(c, menu)
}

func (h *MenuHandler) ListSeller(c *gin.Context) {
	user := middleware.GetUser(c)
	from := c.Query("from")
	to := c.Query("to")
	menus, err := h.menu.ListBySeller(c.Request.Context(), user.ID, from, to)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, menus)
}

type createMenuReq struct {
	DeliveryDate string                  `json:"deliveryDate" binding:"required"`
	Items        []service.MenuItemInput `json:"items"`
}

func (h *MenuHandler) Create(c *gin.Context) {
	user := middleware.GetUser(c)
	var req createMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	menu, err := h.menu.Create(c.Request.Context(), user.ID, req.DeliveryDate, req.Items)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, menu)
}

func (h *MenuHandler) UpdateItems(c *gin.Context) {
	user := middleware.GetUser(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Items []service.MenuItemInput `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	menu, err := h.menu.UpdateItems(c.Request.Context(), user.ID, id, req.Items)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, menu)
}

func (h *MenuHandler) Publish(c *gin.Context) {
	user := middleware.GetUser(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	menu, err := h.menu.Publish(c.Request.Context(), user.ID, id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, menu)
}

type CutoffHandler struct {
	cutoff *service.CutoffService
}

func NewCutoffHandler(cutoff *service.CutoffService) *CutoffHandler {
	return &CutoffHandler{cutoff: cutoff}
}

func (h *CutoffHandler) Status(c *gin.Context) {
	st, err := h.cutoff.GetStatus(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, st)
}

func (h *CutoffHandler) Set(c *gin.Context) {
	var req struct {
		CutoffTime string `json:"cutoffTime" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	st, err := h.cutoff.SetCutoff(c.Request.Context(), req.CutoffTime)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, st)
}

type OrderHandler struct {
	order *service.OrderService
}

func NewOrderHandler(order *service.OrderService) *OrderHandler {
	return &OrderHandler{order: order}
}

func (h *OrderHandler) Create(c *gin.Context) {
	user := middleware.GetUser(c)
	var req service.CreateOrderInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	order, err := h.order.Create(c.Request.Context(), user.ID, req)
	if err != nil {
		if err == service.ErrOrderingClosed {
			response.Fail(c, http.StatusBadRequest, 4001, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, order)
}

func (h *OrderHandler) ListBuyer(c *gin.Context) {
	user := middleware.GetUser(c)
	orders, err := h.order.ListBuyer(c.Request.Context(), user.ID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, orders)
}

func (h *OrderHandler) GetBuyer(c *gin.Context) {
	user := middleware.GetUser(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	order, err := h.order.GetByID(c.Request.Context(), id)
	if err != nil || order.BuyerID != user.ID {
		response.NotFound(c, "order not found")
		return
	}
	response.OK(c, order)
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	user := middleware.GetUser(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.order.CancelBuyer(c.Request.Context(), user.ID, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

type SellerOrderHandler struct {
	order      *service.OrderService
	sellerRole *service.SellerRoleService
}

func NewSellerOrderHandler(order *service.OrderService, sellerRole *service.SellerRoleService) *SellerOrderHandler {
	return &SellerOrderHandler{order: order, sellerRole: sellerRole}
}

func (h *SellerOrderHandler) List(c *gin.Context) {
	date := c.Query("deliveryDate")
	if date == "" {
		response.BadRequest(c, "deliveryDate required")
		return
	}
	orders, err := h.order.ListSellerByDate(c.Request.Context(), date)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, orders)
}

func (h *SellerOrderHandler) Detail(c *gin.Context) {
	user := middleware.GetUser(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	profile, err := h.sellerRole.GetSellerProfile(c.Request.Context(), user.ID)
	if err != nil {
		response.BadRequest(c, "seller profile not found")
		return
	}
	detail, err := h.order.SellerDetail(c.Request.Context(), id, profile)
	if err != nil {
		response.NotFound(c, "order not found")
		return
	}
	response.OK(c, detail)
}

func (h *SellerOrderHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Status model.OrderStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	if err := h.order.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *SellerOrderHandler) Refund(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Reason string `json:"reason" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	if err := h.order.Refund(c.Request.Context(), id, req.Reason, req.Remark); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *SellerOrderHandler) Stats(c *gin.Context) {
	date := c.Query("deliveryDate")
	if date == "" {
		response.BadRequest(c, "deliveryDate required")
		return
	}
	stats, err := h.order.DailyStats(c.Request.Context(), date)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, stats)
}

type RouteHandler struct {
	route *service.RouteService
	amap  *amap.Client
}

func NewRouteHandler(route *service.RouteService, amapClient *amap.Client) *RouteHandler {
	return &RouteHandler{route: route, amap: amapClient}
}

func (h *RouteHandler) Generate(c *gin.Context) {
	user := middleware.GetUser(c)
	var req struct {
		DeliveryDate string `json:"deliveryDate" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	route, err := h.route.Generate(c.Request.Context(), user.ID, req.DeliveryDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, jsonRoute(route))
}

func (h *RouteHandler) Get(c *gin.Context) {
	user := middleware.GetUser(c)
	date := c.Query("deliveryDate")
	route, err := h.route.Get(c.Request.Context(), user.ID, date)
	if err != nil {
		response.NotFound(c, "route not found")
		return
	}
	response.OK(c, jsonRoute(route))
}

func (h *RouteHandler) UpdateStops(c *gin.Context) {
	user := middleware.GetUser(c)
	var req struct {
		DeliveryDate string              `json:"deliveryDate" binding:"required"`
		Stops        []service.RouteStop `json:"stops" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	route, err := h.route.UpdateStops(c.Request.Context(), user.ID, req.DeliveryDate, req.Stops)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, jsonRoute(route))
}

// Cluster groups the date's orders into driverCount balanced delivery routes.
func (h *RouteHandler) Cluster(c *gin.Context) {
	user := middleware.GetUser(c)
	var req struct {
		DeliveryDate string `json:"deliveryDate" binding:"required"`
		DriverCount  int    `json:"driverCount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	if req.DriverCount < 1 {
		response.BadRequest(c, "driverCount must be >= 1")
		return
	}
	result, err := h.route.Cluster(c.Request.Context(), user.ID, req.DeliveryDate, req.DriverCount)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, result)
}

// Clusters returns the last computed clustering for the date.
func (h *RouteHandler) Clusters(c *gin.Context) {
	user := middleware.GetUser(c)
	date := c.Query("deliveryDate")
	result, err := h.route.GetClusters(c.Request.Context(), user.ID, date)
	if err != nil {
		response.NotFound(c, "route not found")
		return
	}
	response.OK(c, result)
}

// Directions proxies Amap's driving (waypoint) API. waypoints is capped at 16.
func (h *RouteHandler) Directions(c *gin.Context) {
	var req struct {
		Origin      amap.LatLng   `json:"origin"`
		Destination amap.LatLng   `json:"destination"`
		Waypoints   []amap.LatLng `json:"waypoints"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	if len(req.Waypoints) > 16 {
		response.BadRequest(c, "too many waypoints (max 16)")
		return
	}
	result, err := h.amap.Driving(c.Request.Context(), req.Origin, req.Destination, req.Waypoints)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func jsonRoute(route *model.DeliveryRoute) gin.H {
	var stops []service.RouteStop
	_ = json.Unmarshal([]byte(route.StopsJSON), &stops)
	return gin.H{
		"deliveryDate":  route.DeliveryDate,
		"stops":         stops,
		"totalDistance": route.TotalDistance,
		"totalDuration": route.TotalDuration,
	}
}

type ChatHandler struct {
	chat *service.ChatService
}

func NewChatHandler(chat *service.ChatService) *ChatHandler {
	return &ChatHandler{chat: chat}
}

func (h *ChatHandler) List(c *gin.Context) {
	user := middleware.GetUser(c)
	orderID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ok, err := h.chat.CanAccess(c.Request.Context(), orderID, user.ID, user.Role)
	if err != nil || !ok {
		response.Forbidden(c, "forbidden")
		return
	}
	since, _ := strconv.ParseInt(c.Query("since"), 10, 64)
	msgs, err := h.chat.List(c.Request.Context(), orderID, since)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	// Opening / polling a conversation marks it as read for this user.
	_ = h.chat.MarkRead(c.Request.Context(), orderID, user.ID)
	response.OK(c, msgs)
}

// Conversations returns the user's message list (one entry per order with chat).
func (h *ChatHandler) Conversations(c *gin.Context) {
	user := middleware.GetUser(c)
	list, err := h.chat.Conversations(c.Request.Context(), user.ID, user.Role)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

// UnreadCount returns the total unread message count for the message tab badge.
func (h *ChatHandler) UnreadCount(c *gin.Context) {
	user := middleware.GetUser(c)
	cnt, err := h.chat.UnreadCount(c.Request.Context(), user.ID, user.Role)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, gin.H{"count": cnt})
}

func (h *ChatHandler) Send(c *gin.Context) {
	user := middleware.GetUser(c)
	orderID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ok, err := h.chat.CanAccess(c.Request.Context(), orderID, user.ID, user.Role)
	if err != nil || !ok {
		response.Forbidden(c, "forbidden")
		return
	}
	var req struct {
		Type    model.MessageType `json:"type" binding:"required"`
		Content string            `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	if req.Type != model.MessageText {
		response.BadRequest(c, "only text messages supported")
		return
	}
	msg, err := h.chat.Send(c.Request.Context(), orderID, user.ID, user.Role, req.Type, req.Content)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, msg)
}

type SellerProfileHandler struct {
	sellerRole *service.SellerRoleService
}

func NewSellerProfileHandler(sellerRole *service.SellerRoleService) *SellerProfileHandler {
	return &SellerProfileHandler{sellerRole: sellerRole}
}

func (h *SellerProfileHandler) Get(c *gin.Context) {
	user := middleware.GetUser(c)
	p, err := h.sellerRole.GetSellerProfile(c.Request.Context(), user.ID)
	if err != nil {
		response.NotFound(c, "not found")
		return
	}
	response.OK(c, p)
}

func (h *SellerProfileHandler) Update(c *gin.Context) {
	user := middleware.GetUser(c)
	var req struct {
		ShopName string  `json:"shopName" binding:"required"`
		Address  string  `json:"address" binding:"required"`
		Lat      float64 `json:"lat"`
		Lng      float64 `json:"lng"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	p, err := h.sellerRole.UpdateSellerProfile(c.Request.Context(), user.ID, req.ShopName, req.Address, req.Lat, req.Lng)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, p)
}
