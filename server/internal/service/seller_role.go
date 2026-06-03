package service

import (
	"context"
	"errors"
	"time"

	"github.com/orderfood/server/internal/model"
	"github.com/orderfood/server/internal/pkg/amap"
	"github.com/orderfood/server/internal/pkg/timeutil"
	"gorm.io/gorm"
)

type SellerRoleService struct {
	db   *gorm.DB
	amap *amap.Client
}

func NewSellerRoleService(db *gorm.DB, amapClient *amap.Client) *SellerRoleService {
	return &SellerRoleService{db: db, amap: amapClient}
}

func (s *SellerRoleService) SyncRoleByPhone(ctx context.Context, userID uint64, phone string) error {
	var allow model.SellerPhone
	err := s.db.WithContext(ctx).Where("phone = ?", phone).First(&allow).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("role", model.RoleBuyer).Error
	}
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("role", model.RoleSeller).Error; err != nil {
		return err
	}
	var profile model.SellerProfile
	err = s.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		profile = model.SellerProfile{
			UserID:     userID,
			ShopName:   allow.ShopName,
			Address:    "待设置店铺地址",
			AddressLat: 0,
			AddressLng: 0,
			UpdatedAt:  time.Now(),
		}
		return s.db.WithContext(ctx).Create(&profile).Error
	}
	return nil
}

func (s *SellerRoleService) IsSellerPhone(ctx context.Context, phone string) bool {
	var count int64
	s.db.WithContext(ctx).Model(&model.SellerPhone{}).Where("phone = ?", phone).Count(&count)
	return count > 0
}

func (s *SellerRoleService) GetSellerProfile(ctx context.Context, userID uint64) (*model.SellerProfile, error) {
	var p model.SellerProfile
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error
	return &p, err
}

func (s *SellerRoleService) UpdateSellerProfile(ctx context.Context, userID uint64, shopName, address string, lat, lng float64) (*model.SellerProfile, error) {
	p, err := s.GetSellerProfile(ctx, userID)
	if err != nil {
		// Create profile if not exists.
		p = &model.SellerProfile{UserID: userID}
	}
	p.ShopName = shopName
	p.Address = address

	// If no explicit coordinates provided, try geocoding the address.
	if lat == 0 && lng == 0 && address != "" {
		if geo, geoErr := s.amap.Geocode(ctx, address); geoErr == nil {
			lat = geo.Lat
			lng = geo.Lng
		}
	}
	p.AddressLat = lat
	p.AddressLng = lng
	p.UpdatedAt = time.Now()

	if p.ID == 0 {
		return p, s.db.WithContext(ctx).Create(p).Error
	}
	return p, s.db.WithContext(ctx).Save(p).Error
}

func (s *SellerRoleService) PrimarySellerID(ctx context.Context) (uint64, error) {
	var user model.User
	err := s.db.WithContext(ctx).Where("role = ?", model.RoleSeller).Order("id asc").First(&user).Error
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (s *SellerRoleService) EnsureMenuNotExpired(ctx context.Context) error {
	today := timeutil.Today()
	return s.db.WithContext(ctx).Model(&model.Menu{}).
		Where("delivery_date < ? AND status != ?", today, model.MenuExpired).
		Update("status", model.MenuExpired).Error
}
