package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/orderfood/server/internal/model"
	"github.com/orderfood/server/internal/pkg/amap"
	"github.com/orderfood/server/internal/pkg/timeutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// DriverRoute is one delivery driver's ordered list of stops.
type DriverRoute struct {
	DriverIndex   int         `json:"driverIndex"`
	Color         string      `json:"color"`
	Stops         []RouteStop `json:"stops"`
	TotalDistance int         `json:"totalDistance"`
}

// ClusterResult is the full per-driver clustering for a delivery date.
type ClusterResult struct {
	DeliveryDate time.Time     `json:"deliveryDate"`
	DriverCount  int           `json:"driverCount"`
	SellerLat    float64       `json:"sellerLat"`
	SellerLng    float64       `json:"sellerLng"`
	Drivers      []DriverRoute `json:"drivers"`
}

// routeColors are cycled across drivers for map rendering.
var routeColors = []string{
	"#07c160", "#1989fa", "#ff976a", "#ee0a24",
	"#7232dd", "#ff5722", "#00b578", "#fa8c16",
}

func (s *RouteService) sellerLatLng(ctx context.Context, sellerID uint64) (float64, float64) {
	var p model.SellerProfile
	s.db.WithContext(ctx).Where("user_id = ?", sellerID).First(&p)
	return p.AddressLat, p.AddressLng
}

// Cluster groups the date's deliverable orders into driverCount balanced clusters
// (capacity-constrained k-means), nearest-neighbor sorts each cluster starting from
// the stop closest to the shop, persists the result, and returns it.
func (s *RouteService) Cluster(ctx context.Context, sellerID uint64, deliveryDate string, driverCount int) (*ClusterResult, error) {
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

	sellerLat, sellerLng := s.sellerLatLng(ctx, sellerID)

	k := driverCount
	if k < 1 {
		k = 1
	}
	if k > len(stops) {
		k = len(stops)
	}

	var drivers []DriverRoute
	if k > 0 {
		labels := balancedKMeans(stops, k)
		buckets := make([][]RouteStop, k)
		for i, lbl := range labels {
			buckets[lbl] = append(buckets[lbl], stops[i])
		}
		for ci, bucket := range buckets {
			sorted := nearestNeighborFrom(bucket, sellerLat, sellerLng)
			drivers = append(drivers, DriverRoute{
				DriverIndex:   ci,
				Color:         routeColors[ci%len(routeColors)],
				Stops:         sorted,
				TotalDistance: estimateDistance(sorted),
			})
		}
	}

	clustersRaw, _ := json.Marshal(drivers)
	// Keep StopsJSON as a flattened ordering for backward compatibility.
	var flat []RouteStop
	for _, d := range drivers {
		flat = append(flat, d.Stops...)
	}
	flatRaw, _ := json.Marshal(flat)
	d, err := timeutil.ParseDate(deliveryDate)
	if err != nil {
		return nil, err
	}
	route := model.DeliveryRoute{
		SellerID: sellerID, DeliveryDate: d,
		StopsJSON:     string(flatRaw),
		DriverCount:   driverCount,
		ClustersJSON:  string(clustersRaw),
		TotalDistance: estimateDistance(flat),
		UpdatedAt:     time.Now(),
	}
	if err := s.db.WithContext(ctx).Where("seller_id = ? AND delivery_date = ?", sellerID, d).Assign(route).FirstOrCreate(&route).Error; err != nil {
		return nil, err
	}
	return &ClusterResult{
		DeliveryDate: d, DriverCount: driverCount,
		SellerLat: sellerLat, SellerLng: sellerLng, Drivers: drivers,
	}, nil
}

// GetClusters reads back a previously computed clustering for the date.
func (s *RouteService) GetClusters(ctx context.Context, sellerID uint64, deliveryDate string) (*ClusterResult, error) {
	d, err := timeutil.ParseDate(deliveryDate)
	if err != nil {
		return nil, err
	}
	var route model.DeliveryRoute
	if err := s.db.WithContext(ctx).Where("seller_id = ? AND delivery_date = ?", sellerID, d).First(&route).Error; err != nil {
		return nil, err
	}
	var drivers []DriverRoute
	if route.ClustersJSON != "" {
		_ = json.Unmarshal([]byte(route.ClustersJSON), &drivers)
	}
	sellerLat, sellerLng := s.sellerLatLng(ctx, sellerID)
	return &ClusterResult{
		DeliveryDate: route.DeliveryDate, DriverCount: route.DriverCount,
		SellerLat: sellerLat, SellerLng: sellerLng, Drivers: drivers,
	}, nil
}

// nearestNeighborFrom orders stops greedily, starting from the stop nearest the origin.
func nearestNeighborFrom(stops []RouteStop, originLat, originLng float64) []RouteStop {
	if len(stops) <= 1 {
		return stops
	}
	remaining := append([]RouteStop{}, stops...)
	// Seed with the stop closest to the origin (shop).
	seed := 0
	seedD := 1e18
	for i, s := range remaining {
		d := amap.HaversineDistanceKm(originLat, originLng, s.Lat, s.Lng)
		if d < seedD {
			seedD = d
			seed = i
		}
	}
	current := remaining[seed]
	result := []RouteStop{current}
	remaining = append(remaining[:seed], remaining[seed+1:]...)
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

type centroid struct{ lat, lng float64 }

// balancedKMeans assigns each stop to one of k clusters of near-equal size
// (capacity = ceil(n/k)) using deterministic farthest-point seeding and a
// greedy capacity-constrained assignment. Returns the cluster label per stop.
func balancedKMeans(stops []RouteStop, k int) []int {
	n := len(stops)
	labels := make([]int, n)
	if k <= 1 || n == 0 {
		return labels
	}

	capacity := (n + k - 1) / k // ceil(n/k)

	// Farthest-point seeding for spread-out, deterministic initial centroids.
	centroids := make([]centroid, k)
	centroids[0] = centroid{stops[0].Lat, stops[0].Lng}
	for j := 1; j < k; j++ {
		far := 0
		farD := -1.0
		for i, s := range stops {
			minD := 1e18
			for c := 0; c < j; c++ {
				d := amap.HaversineDistanceKm(s.Lat, s.Lng, centroids[c].lat, centroids[c].lng)
				if d < minD {
					minD = d
				}
			}
			if minD > farD {
				farD = minD
				far = i
			}
		}
		centroids[j] = centroid{stops[far].Lat, stops[far].Lng}
	}

	for iter := 0; iter < 12; iter++ {
		// Greedy capacity-constrained assignment: assign closest pairs first.
		type pair struct {
			d          float64
			stop, clus int
		}
		pairs := make([]pair, 0, n*k)
		for i, s := range stops {
			for c := 0; c < k; c++ {
				pairs = append(pairs, pair{
					d:    amap.HaversineDistanceKm(s.Lat, s.Lng, centroids[c].lat, centroids[c].lng),
					stop: i, clus: c,
				})
			}
		}
		sort.Slice(pairs, func(a, b int) bool { return pairs[a].d < pairs[b].d })

		assigned := make([]bool, n)
		counts := make([]int, k)
		filled := 0
		newLabels := make([]int, n)
		for _, p := range pairs {
			if filled == n {
				break
			}
			if assigned[p.stop] || counts[p.clus] >= capacity {
				continue
			}
			newLabels[p.stop] = p.clus
			assigned[p.stop] = true
			counts[p.clus]++
			filled++
		}

		// Recompute centroids from the new assignment.
		sumLat := make([]float64, k)
		sumLng := make([]float64, k)
		cnt := make([]int, k)
		for i, lbl := range newLabels {
			sumLat[lbl] += stops[i].Lat
			sumLng[lbl] += stops[i].Lng
			cnt[lbl]++
		}
		changed := false
		for c := 0; c < k; c++ {
			if cnt[c] == 0 {
				continue
			}
			centroids[c] = centroid{sumLat[c] / float64(cnt[c]), sumLng[c] / float64(cnt[c])}
		}
		for i := range labels {
			if labels[i] != newLabels[i] {
				changed = true
			}
			labels[i] = newLabels[i]
		}
		if !changed && iter > 0 {
			break
		}
	}
	return labels
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

// ConversationSummary is one entry in a user's message list.
type ConversationSummary struct {
	OrderID     uint64    `json:"orderId"`
	OrderNo     string    `json:"orderNo"`
	OrderStatus string    `json:"orderStatus"`
	PeerName    string    `json:"peerName"`
	LastContent string    `json:"lastContent"`
	LastAt      time.Time `json:"lastAt"`
	Unread      int64     `json:"unread"`
}

// accessibleOrderIDs returns the order ids (with at least one message) the user
// is allowed to see. Seller sees every conversation; buyer sees only their own.
func (s *ChatService) accessibleOrderIDs(ctx context.Context, userID uint64, role model.UserRole) ([]uint64, error) {
	db := s.db.WithContext(ctx)
	q := db.Model(&model.OrderMessage{}).Distinct("order_id")
	if role != model.RoleSeller {
		q = q.Where("order_id IN (?)",
			db.Model(&model.Order{}).Select("id").Where("buyer_id = ?", userID))
	}
	var ids []uint64
	err := q.Pluck("order_id", &ids).Error
	return ids, err
}

// Conversations builds the message list for a user, newest activity first.
func (s *ChatService) Conversations(ctx context.Context, userID uint64, role model.UserRole) ([]ConversationSummary, error) {
	db := s.db.WithContext(ctx)
	ids, err := s.accessibleOrderIDs(ctx, userID, role)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []ConversationSummary{}, nil
	}

	var orders []model.Order
	if err := db.Where("id IN ?", ids).Find(&orders).Error; err != nil {
		return nil, err
	}
	orderMap := make(map[uint64]model.Order, len(orders))
	for _, o := range orders {
		orderMap[o.ID] = o
	}

	markerMap := make(map[uint64]time.Time)
	var markers []model.OrderReadMarker
	if err := db.Where("user_id = ? AND order_id IN ?", userID, ids).Find(&markers).Error; err == nil {
		for _, m := range markers {
			markerMap[m.OrderID] = m.LastReadAt
		}
	}

	result := make([]ConversationSummary, 0, len(ids))
	for _, oid := range ids {
		var last model.OrderMessage
		if err := db.Where("order_id = ?", oid).Order("created_at desc").First(&last).Error; err != nil {
			continue
		}
		unreadQ := db.Model(&model.OrderMessage{}).Where("order_id = ? AND sender_id <> ?", oid, userID)
		if lastRead, ok := markerMap[oid]; ok {
			unreadQ = unreadQ.Where("created_at > ?", lastRead)
		}
		var unread int64
		unreadQ.Count(&unread)

		o := orderMap[oid]
		peer := "店铺"
		if role == model.RoleSeller {
			peer = o.ContactName
		}
		result = append(result, ConversationSummary{
			OrderID:     oid,
			OrderNo:     o.OrderNo,
			OrderStatus: string(o.Status),
			PeerName:    peer,
			LastContent: last.Content,
			LastAt:      last.CreatedAt,
			Unread:      unread,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].LastAt.After(result[j].LastAt)
	})
	return result, nil
}

// UnreadCount returns the total number of unread messages across all of the
// user's conversations (drives the message tab badge).
func (s *ChatService) UnreadCount(ctx context.Context, userID uint64, role model.UserRole) (int64, error) {
	db := s.db.WithContext(ctx)
	q := db.Table("order_messages AS m").
		Joins("LEFT JOIN order_read_markers AS r ON r.order_id = m.order_id AND r.user_id = ?", userID).
		Where("m.sender_id <> ?", userID).
		Where("r.last_read_at IS NULL OR m.created_at > r.last_read_at")
	if role != model.RoleSeller {
		q = q.Where("m.order_id IN (?)",
			db.Model(&model.Order{}).Select("id").Where("buyer_id = ?", userID))
	}
	var cnt int64
	err := q.Count(&cnt).Error
	return cnt, err
}

// MarkRead upserts the user's read marker for an order to now.
func (s *ChatService) MarkRead(ctx context.Context, orderID, userID uint64) error {
	now := time.Now()
	marker := model.OrderReadMarker{OrderID: orderID, UserID: userID, LastReadAt: now, UpdatedAt: now}
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "order_id"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"last_read_at", "updated_at"}),
		}).Create(&marker).Error
}
