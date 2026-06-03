package model

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleBuyer  UserRole = "buyer"
	RoleSeller UserRole = "seller"
)

type User struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	OpenID    string         `gorm:"uniqueIndex;size:64;not null" json:"openId"`
	Nickname  string         `gorm:"size:64" json:"nickname"`
	AvatarURL string         `gorm:"size:512" json:"avatarUrl"`
	Role      UserRole       `gorm:"size:16;not null;default:buyer" json:"role"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type BuyerProfile struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	UserID           uint64    `gorm:"uniqueIndex;not null" json:"userId"`
	ContactName      string    `gorm:"size:64;not null" json:"contactName"`
	ContactPhone     string    `gorm:"size:20;not null" json:"contactPhone"`
	Address          string    `gorm:"size:512;not null" json:"address"`
	AddressLat       *float64  `json:"addressLat,omitempty"`
	AddressLng       *float64  `json:"addressLng,omitempty"`
	ProfileCompleted bool      `gorm:"not null;default:false" json:"profileCompleted"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type SellerProfile struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	UserID     uint64    `gorm:"uniqueIndex;not null" json:"userId"`
	ShopName   string    `gorm:"size:128;not null" json:"shopName"`
	Address    string    `gorm:"size:512;not null" json:"address"`
	AddressLat float64   `gorm:"not null" json:"addressLat"`
	AddressLng float64   `gorm:"not null" json:"addressLng"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// SellerPhone binds seller role to a phone number (whitelist).
type SellerPhone struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Phone     string    `gorm:"uniqueIndex;size:20;not null" json:"phone"`
	ShopName  string    `gorm:"size:128;not null;default:我的店铺" json:"shopName"`
	CreatedAt time.Time `json:"createdAt"`
}

type MenuStatus string

const (
	MenuDraft     MenuStatus = "draft"
	MenuPublished MenuStatus = "published"
	MenuExpired   MenuStatus = "expired"
)

type Menu struct {
	ID           uint64         `gorm:"primaryKey" json:"id"`
	SellerID     uint64         `gorm:"index;not null" json:"sellerId"`
	DeliveryDate time.Time      `gorm:"type:date;index;not null" json:"deliveryDate"`
	Status       MenuStatus     `gorm:"size:16;not null;default:draft" json:"status"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Items        []MenuItem     `gorm:"foreignKey:MenuID" json:"items,omitempty"`
}

type MenuItem struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	MenuID      uint64         `gorm:"index;not null" json:"menuId"`
	Name        string         `gorm:"size:128;not null" json:"name"`
	ImageURL    string         `gorm:"size:512" json:"imageUrl"`
	Price       int64          `gorm:"not null" json:"price"`
	Description string         `gorm:"size:512" json:"description"`
	SortOrder   int            `gorm:"not null;default:0" json:"sortOrder"`
	IsAvailable bool           `gorm:"not null;default:true" json:"isAvailable"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type OrderStatus string

const (
	OrderPendingPayment OrderStatus = "pending_payment"
	OrderPending        OrderStatus = "pending"
	OrderConfirmed      OrderStatus = "confirmed"
	OrderDelivering     OrderStatus = "delivering"
	OrderCompleted      OrderStatus = "completed"
	OrderRefunded       OrderStatus = "refunded"
	OrderCancelled      OrderStatus = "cancelled"
)

type Order struct {
	ID               uint64         `gorm:"primaryKey" json:"id"`
	OrderNo          string         `gorm:"uniqueIndex;size:32;not null" json:"orderNo"`
	BuyerID          uint64         `gorm:"index;not null" json:"buyerId"`
	MenuID           uint64         `gorm:"index;not null" json:"menuId"`
	DeliveryDate     time.Time      `gorm:"type:date;index;not null" json:"deliveryDate"`
	TotalAmount      int64          `gorm:"not null" json:"totalAmount"`
	Status           OrderStatus    `gorm:"size:32;not null;default:pending_payment" json:"status"`
	ContactName      string         `gorm:"size:64;not null" json:"contactName"`
	ContactPhone     string         `gorm:"size:20;not null" json:"contactPhone"`
	Address          string         `gorm:"size:512;not null" json:"address"`
	AddressLat       *float64       `json:"addressLat,omitempty"`
	AddressLng       *float64       `json:"addressLng,omitempty"`
	DeliveryTimePref string         `gorm:"size:64" json:"deliveryTimePref"`
	Remark           string         `gorm:"size:512" json:"remark"`
	RefundReason     string         `gorm:"size:256" json:"refundReason,omitempty"`
	RefundRemark     string         `gorm:"size:512" json:"refundRemark,omitempty"`
	RefundedAt       *time.Time     `json:"refundedAt,omitempty"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	Items            []OrderItem    `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

type OrderItem struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	OrderID       uint64    `gorm:"index;not null" json:"orderId"`
	MenuItemID    uint64    `gorm:"not null" json:"menuItemId"`
	NameSnapshot  string    `gorm:"size:128;not null" json:"nameSnapshot"`
	PriceSnapshot int64     `gorm:"not null" json:"priceSnapshot"`
	Quantity      int       `gorm:"not null" json:"quantity"`
	CreatedAt     time.Time `json:"createdAt"`
}

type PaymentStatus string

const (
	PaymentPending PaymentStatus = "pending"
	PaymentPaid    PaymentStatus = "paid"
	PaymentFailed  PaymentStatus = "failed"
	PaymentRefunded PaymentStatus = "refunded"
)

type Payment struct {
	ID            uint64        `gorm:"primaryKey" json:"id"`
	OrderID       uint64        `gorm:"uniqueIndex;not null" json:"orderId"`
	TransactionID string        `gorm:"size:64" json:"transactionId"`
	Amount        int64         `gorm:"not null" json:"amount"`
	Status        PaymentStatus `gorm:"size:16;not null;default:pending" json:"status"`
	PaidAt        *time.Time    `json:"paidAt,omitempty"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

type CutoffSetting struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	OrderDate  time.Time `gorm:"type:date;uniqueIndex;not null" json:"orderDate"`
	CutoffTime string    `gorm:"size:8;not null" json:"cutoffTime"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type MessageType string

const (
	MessageText  MessageType = "text"
	MessageImage MessageType = "image"
)

type OrderMessage struct {
	ID         uint64      `gorm:"primaryKey" json:"id"`
	OrderID    uint64      `gorm:"index;not null" json:"orderId"`
	SenderID   uint64      `gorm:"index;not null" json:"senderId"`
	SenderRole UserRole    `gorm:"size:16;not null" json:"senderRole"`
	Type       MessageType `gorm:"size:16;not null" json:"type"`
	Content    string      `gorm:"type:text;not null" json:"content"`
	CreatedAt  time.Time   `gorm:"index" json:"createdAt"`
}

type DeliveryRoute struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	SellerID     uint64    `gorm:"index;not null" json:"sellerId"`
	DeliveryDate time.Time `gorm:"type:date;uniqueIndex;not null" json:"deliveryDate"`
	StopsJSON    string    `gorm:"type:text;not null" json:"stopsJson"`
	TotalDistance int      `gorm:"not null;default:0" json:"totalDistance"`
	TotalDuration int      `gorm:"not null;default:0" json:"totalDuration"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&BuyerProfile{},
		&SellerProfile{},
		&SellerPhone{},
		&Menu{},
		&MenuItem{},
		&Order{},
		&OrderItem{},
		&Payment{},
		&CutoffSetting{},
		&OrderMessage{},
		&DeliveryRoute{},
	)
}
