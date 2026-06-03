package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/orderfood/server/internal/model"
	"github.com/orderfood/server/internal/pkg/amap"
	"github.com/orderfood/server/internal/pkg/timeutil"
	"gorm.io/gorm"
)

var ErrOrderingClosed = errors.New("ordering is closed")
var ErrOrderNotFound = errors.New("order not found")

type OrderService struct {
	db     *gorm.DB
	menu   *MenuService
	cutoff *CutoffService
	amap   *amap.Client
}

func NewOrderService(db *gorm.DB, menu *MenuService, cutoff *CutoffService, amapClient *amap.Client) *OrderService {
	return &OrderService{db: db, menu: menu, cutoff: cutoff, amap: amapClient}
}

type OrderItemReq struct {
	MenuItemID uint64 `json:"menuItemId"`
	Quantity   int    `json:"quantity"`
}

type CreateOrderInput struct {
	Items            []OrderItemReq `json:"items"`
	ContactName      string         `json:"contactName"`
	ContactPhone     string         `json:"contactPhone"`
	Address          string         `json:"address"`
	DeliveryTimePref string         `json:"deliveryTimePref"`
	Remark           string         `json:"remark"`
}

func (s *OrderService) Create(ctx context.Context, buyerID uint64, input CreateOrderInput) (*model.Order, error) {
	open, err := s.cutoff.IsOrderingOpen(ctx)
	if err != nil {
		return nil, err
	}
	if !open {
		return nil, ErrOrderingClosed
	}
	menu, err := s.menu.GetTomorrowPublished(ctx)
	if err != nil {
		return nil, fmt.Errorf("tomorrow menu not available")
	}
	if len(input.Items) == 0 {
		return nil, fmt.Errorf("empty order")
	}

	itemMap := map[uint64]model.MenuItem{}
	for _, it := range menu.Items {
		itemMap[it.ID] = it
	}

	var total int64
	var orderItems []model.OrderItem
	for _, req := range input.Items {
		mi, ok := itemMap[req.MenuItemID]
		if !ok || !mi.IsAvailable || req.Quantity <= 0 {
			return nil, fmt.Errorf("invalid menu item")
		}
		total += mi.Price * int64(req.Quantity)
		orderItems = append(orderItems, model.OrderItem{
			MenuItemID: mi.ID, NameSnapshot: mi.Name,
			PriceSnapshot: mi.Price, Quantity: req.Quantity,
		})
	}

	lat, lng := (*float64)(nil), (*float64)(nil)
	if geo, err := s.amap.Geocode(ctx, input.Address); err == nil {
		lat, lng = &geo.Lat, &geo.Lng
	}

	order := model.Order{
		OrderNo:          genOrderNo(),
		BuyerID:          buyerID,
		MenuID:           menu.ID,
		DeliveryDate:     timeutil.Tomorrow(),
		TotalAmount:      total,
		Status:           model.OrderPending,
		ContactName:      input.ContactName,
		ContactPhone:     input.ContactPhone,
		Address:          input.Address,
		AddressLat:       lat,
		AddressLng:       lng,
		DeliveryTimePref: input.DeliveryTimePref,
		Remark:           input.Remark,
	}
	if err := s.db.WithContext(ctx).Create(&order).Error; err != nil {
		return nil, err
	}
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := s.db.WithContext(ctx).Create(&orderItems[i]).Error; err != nil {
			return nil, err
		}
	}
	return s.GetByID(ctx, order.ID)
}

func (s *OrderService) GetByID(ctx context.Context, id uint64) (*model.Order, error) {
	var order model.Order
	err := s.db.WithContext(ctx).Preload("Items").First(&order, id).Error
	return &order, err
}

func (s *OrderService) ListBuyer(ctx context.Context, buyerID uint64) ([]model.Order, error) {
	var orders []model.Order
	err := s.db.WithContext(ctx).Where("buyer_id = ?", buyerID).Order("delivery_date desc, id desc").Preload("Items").Find(&orders).Error
	return orders, err
}

func (s *OrderService) CancelBuyer(ctx context.Context, buyerID, orderID uint64) error {
	open, _ := s.cutoff.IsOrderingOpen(ctx)
	if !open {
		return fmt.Errorf("cannot cancel after cutoff")
	}
	var order model.Order
	if err := s.db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		return err
	}
	if order.BuyerID != buyerID {
		return ErrOrderNotFound
	}
	if order.Status != model.OrderPending {
		return fmt.Errorf("order cannot be cancelled")
	}
	return s.db.WithContext(ctx).Model(&order).Update("status", model.OrderCancelled).Error
}

func (s *OrderService) ListSellerByDate(ctx context.Context, deliveryDate string) ([]model.Order, error) {
	d, err := timeutil.ParseDate(deliveryDate)
	if err != nil {
		return nil, err
	}
	var orders []model.Order
	err = s.db.WithContext(ctx).Where("delivery_date = ? AND status NOT IN ?", d, []model.OrderStatus{model.OrderCancelled}).
		Order("id asc").Preload("Items").Find(&orders).Error
	return orders, err
}

type OrderDetailSeller struct {
	Order           model.Order `json:"order"`
	DistanceKm      *float64    `json:"distanceKm,omitempty"`
	SellerLat       float64     `json:"sellerLat"`
	SellerLng       float64     `json:"sellerLng"`
	BuyerLat        *float64    `json:"buyerLat,omitempty"`
	BuyerLng        *float64    `json:"buyerLng,omitempty"`
}

func (s *OrderService) SellerDetail(ctx context.Context, orderID uint64, sellerProfile *model.SellerProfile) (*OrderDetailSeller, error) {
	order, err := s.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	detail := &OrderDetailSeller{
		Order:     *order,
		SellerLat: sellerProfile.AddressLat,
		SellerLng: sellerProfile.AddressLng,
		BuyerLat:  order.AddressLat,
		BuyerLng:  order.AddressLng,
	}
	if order.AddressLat != nil && order.AddressLng != nil {
		d := amap.HaversineDistanceKm(sellerProfile.AddressLat, sellerProfile.AddressLng, *order.AddressLat, *order.AddressLng)
		detail.DistanceKm = &d
	}
	return detail, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, orderID uint64, status model.OrderStatus) error {
	return s.db.WithContext(ctx).Model(&model.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

func (s *OrderService) Refund(ctx context.Context, orderID uint64, reason, remark string) error {
	if reason == "" {
		return fmt.Errorf("refund reason required")
	}
	now := time.Now()
	return s.db.WithContext(ctx).Model(&model.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"status":        model.OrderRefunded,
		"refund_reason": reason,
		"refund_remark": remark,
		"refunded_at":   &now,
	}).Error
}

func (s *OrderService) DailyStats(ctx context.Context, deliveryDate string) (map[string]interface{}, error) {
	d, err := timeutil.ParseDate(deliveryDate)
	if err != nil {
		return nil, err
	}
	var orders []model.Order
	s.db.WithContext(ctx).Where("delivery_date = ?", d).Find(&orders)
	stats := map[string]interface{}{
		"totalCount": 0, "totalRevenue": int64(0),
		"pending": 0, "confirmed": 0, "delivering": 0, "completed": 0, "refunded": 0,
	}
	for _, o := range orders {
		stats["totalCount"] = stats["totalCount"].(int) + 1
		if o.Status != model.OrderRefunded && o.Status != model.OrderCancelled {
			stats["totalRevenue"] = stats["totalRevenue"].(int64) + o.TotalAmount
		}
		if c, ok := stats[string(o.Status)].(int); ok {
			stats[string(o.Status)] = c + 1
		}
	}
	return stats, nil
}

func genOrderNo() string {
	return fmt.Sprintf("OF%s", time.Now().Format("20060102150405")) + uuid.New().String()[:6]
}

type RouteStop struct {
	OrderID      uint64  `json:"orderId"`
	ContactName  string  `json:"contactName"`
	Address      string  `json:"address"`
	Phone        string  `json:"phone"`
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
}

type RouteService struct {
	db    *gorm.DB
	order *OrderService
}

func NewRouteService(db *gorm.DB, order *OrderService) *RouteService {
	return &RouteService{db: db, order: order}
}

func (s *RouteService) Generate(ctx context.Context, sellerID uint64, deliveryDate string) (*model.DeliveryRoute, error) {
	orders, err := s.order.ListSellerByDate(ctx, deliveryDate)
	if err != nil {
		return nil, err
	}
	var stops []RouteStop
	for _, o := range orders {
		if o.Status == model.OrderRefunded || o.Status == model.OrderCancelled {
			continue
		}
		if o.AddressLat == nil || o.AddressLng == nil {
			continue
		}
		stops = append(stops, RouteStop{
			OrderID: o.ID, ContactName: o.ContactName, Address: o.Address,
			Phone: o.ContactPhone, Lat: *o.AddressLat, Lng: *o.AddressLng,
		})
	}
	// nearest-neighbor sort from first stop (simple heuristic)
	sorted := nearestNeighbor(stops)
	raw, _ := json.Marshal(sorted)
	d, _ := timeutil.ParseDate(deliveryDate)
	route := model.DeliveryRoute{
		SellerID: sellerID, DeliveryDate: d, StopsJSON: string(raw),
		TotalDistance: estimateDistance(sorted), UpdatedAt: time.Now(),
	}
	err = s.db.WithContext(ctx).Where("seller_id = ? AND delivery_date = ?", sellerID, d).Assign(route).FirstOrCreate(&route).Error
	return &route, err
}

func (s *RouteService) Get(ctx context.Context, sellerID uint64, deliveryDate string) (*model.DeliveryRoute, error) {
	d, err := timeutil.ParseDate(deliveryDate)
	if err != nil {
		return nil, err
	}
	var route model.DeliveryRoute
	err = s.db.WithContext(ctx).Where("seller_id = ? AND delivery_date = ?", sellerID, d).First(&route).Error
	return &route, err
}

func (s *RouteService) UpdateStops(ctx context.Context, sellerID uint64, deliveryDate string, stops []RouteStop) (*model.DeliveryRoute, error) {
	d, err := timeutil.ParseDate(deliveryDate)
	if err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(stops)
	route := model.DeliveryRoute{
		SellerID: sellerID, DeliveryDate: d, StopsJSON: string(raw),
		TotalDistance: estimateDistance(stops), UpdatedAt: time.Now(),
	}
	err = s.db.WithContext(ctx).Where("seller_id = ? AND delivery_date = ?", sellerID, d).Assign(route).FirstOrCreate(&route).Error
	return &route, err
}

func nearestNeighbor(stops []RouteStop) []RouteStop {
	if len(stops) <= 1 {
		return stops
	}
	remaining := append([]RouteStop{}, stops...)
	var result []RouteStop
	current := remaining[0]
	result = append(result, current)
	remaining = remaining[1:]
	for len(remaining) > 0 {
		best := 0
		bestD := 1e18
		for i, s := range remaining {
			d := amap.HaversineDistanceKm(current.Lat, current.Lng, s.Lat, s.Lng)
			if d < bestD {
				bestD = d
				best = i
			}
		}
		current = remaining[best]
		result = append(result, current)
		remaining = append(remaining[:best], remaining[best+1:]...)
	}
	return result
}

func estimateDistance(stops []RouteStop) int {
	if len(stops) < 2 {
		return 0
	}
	var km float64
	for i := 1; i < len(stops); i++ {
		km += amap.HaversineDistanceKm(stops[i-1].Lat, stops[i-1].Lng, stops[i].Lat, stops[i].Lng)
	}
	return int(km * 1000)
}

type ChatService struct {
	db    *gorm.DB
	redis RedisClient
}

func NewChatService(db *gorm.DB, redis RedisClient) *ChatService {
	return &ChatService{db: db, redis: redis}
}

func (s *ChatService) Send(ctx context.Context, orderID, senderID uint64, role model.UserRole, msgType model.MessageType, content string) (*model.OrderMessage, error) {
	msg := model.OrderMessage{
		OrderID: orderID, SenderID: senderID, SenderRole: role,
		Type: msgType, Content: content, CreatedAt: time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(&msg).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

func (s *ChatService) List(ctx context.Context, orderID uint64, since int64) ([]model.OrderMessage, error) {
	q := s.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at asc")
	if since > 0 {
		q = q.Where("created_at > ?", time.UnixMilli(since))
	}
	var msgs []model.OrderMessage
	err := q.Find(&msgs).Error
	return msgs, err
}

func (s *ChatService) CanAccess(ctx context.Context, orderID, userID uint64, role model.UserRole) (bool, error) {
	var order model.Order
	if err := s.db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		return false, err
	}
	if role == model.RoleSeller {
		return true, nil
	}
	return order.BuyerID == userID, nil
}
