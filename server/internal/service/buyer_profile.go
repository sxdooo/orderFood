package service

import (
	"context"
	"errors"

	"github.com/orderfood/server/internal/model"
	"github.com/orderfood/server/internal/pkg/amap"
	"gorm.io/gorm"
)

type BuyerProfileService struct {
	db         *gorm.DB
	amap       *amap.Client
	sellerRole *SellerRoleService
}

func NewBuyerProfileService(db *gorm.DB, amapClient *amap.Client, sellerRole *SellerRoleService) *BuyerProfileService {
	return &BuyerProfileService{db: db, amap: amapClient, sellerRole: sellerRole}
}

type BuyerProfileInput struct {
	ContactName  string `json:"contactName" binding:"required"`
	ContactPhone string `json:"contactPhone" binding:"required"`
	Address      string `json:"address" binding:"required"`
}

func (s *BuyerProfileService) GetByUserID(ctx context.Context, userID uint64) (*model.BuyerProfile, error) {
	var profile model.BuyerProfile
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &model.BuyerProfile{UserID: userID, ProfileCompleted: false}, nil
	}
	return &profile, err
}

func (s *BuyerProfileService) Upsert(ctx context.Context, userID uint64, input BuyerProfileInput) (*model.BuyerProfile, error) {
	profile, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile.ContactName = input.ContactName
	profile.ContactPhone = input.ContactPhone
	profile.Address = input.Address
	profile.ProfileCompleted = true

	if geo, err := s.amap.Geocode(ctx, input.Address); err == nil {
		profile.AddressLat = &geo.Lat
		profile.AddressLng = &geo.Lng
	}

	if profile.ID == 0 {
		profile.UserID = userID
		err = s.db.WithContext(ctx).Create(profile).Error
	} else {
		err = s.db.WithContext(ctx).Save(profile).Error
	}
	if err != nil {
		return nil, err
	}
	_ = s.sellerRole.SyncRoleByPhone(ctx, userID, input.ContactPhone)
	return profile, nil
}
