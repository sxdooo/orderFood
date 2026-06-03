package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/orderfood/server/internal/model"
	"github.com/orderfood/server/internal/pkg/timeutil"
	"gorm.io/gorm"
)

var ErrMenuNotFound = errors.New("menu not found")
var ErrMenuDuplicate = errors.New("menu already exists for date")

type MenuService struct {
	db *gorm.DB
}

func NewMenuService(db *gorm.DB) *MenuService {
	return &MenuService{db: db}
}

type MenuItemInput struct {
	Name        string `json:"name"`
	Price       int64  `json:"price"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
}

func (s *MenuService) ListBySeller(ctx context.Context, sellerID uint64, from, to string) ([]model.Menu, error) {
	q := s.db.WithContext(ctx).Preload("Items", "is_available = ?", true).Where("seller_id = ?", sellerID)
	if from != "" {
		if d, err := timeutil.ParseDate(from); err == nil {
			q = q.Where("delivery_date >= ?", d)
		}
	}
	if to != "" {
		if d, err := timeutil.ParseDate(to); err == nil {
			q = q.Where("delivery_date <= ?", d)
		}
	}
	var menus []model.Menu
	err := q.Order("delivery_date asc").Find(&menus).Error
	return menus, err
}

func (s *MenuService) GetTomorrowPublished(ctx context.Context) (*model.Menu, error) {
	tomorrow := timeutil.Tomorrow()
	var menu model.Menu
	err := s.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_available = ?", true).Order("sort_order asc")
		}).
		Where("delivery_date = ? AND status = ?", tomorrow, model.MenuPublished).
		First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (s *MenuService) Create(ctx context.Context, sellerID uint64, deliveryDateStr string, items []MenuItemInput) (*model.Menu, error) {
	d, err := timeutil.ParseDate(deliveryDateStr)
	if err != nil {
		return nil, err
	}
	var exists int64
	s.db.WithContext(ctx).Model(&model.Menu{}).Where("seller_id = ? AND delivery_date = ?", sellerID, d).Count(&exists)
	if exists > 0 {
		return nil, ErrMenuDuplicate
	}
	menu := model.Menu{SellerID: sellerID, DeliveryDate: d, Status: model.MenuDraft}
	if err := s.db.WithContext(ctx).Create(&menu).Error; err != nil {
		return nil, err
	}
	for i, it := range items {
		item := model.MenuItem{
			MenuID: menu.ID, Name: it.Name, Price: it.Price,
			Description: it.Description, SortOrder: i, IsAvailable: true,
		}
		if err := s.db.WithContext(ctx).Create(&item).Error; err != nil {
			return nil, err
		}
	}
	return s.GetByID(ctx, menu.ID)
}

func (s *MenuService) GetByID(ctx context.Context, id uint64) (*model.Menu, error) {
	var menu model.Menu
	err := s.db.WithContext(ctx).Preload("Items").First(&menu, id).Error
	return &menu, err
}

func (s *MenuService) UpdateItems(ctx context.Context, sellerID, menuID uint64, items []MenuItemInput) (*model.Menu, error) {
	menu, err := s.GetByID(ctx, menuID)
	if err != nil {
		return nil, err
	}
	if menu.SellerID != sellerID {
		return nil, ErrMenuNotFound
	}
	if menu.Status == model.MenuExpired {
		return nil, fmt.Errorf("menu expired")
	}
	// soft-disable all then upsert
	s.db.WithContext(ctx).Model(&model.MenuItem{}).Where("menu_id = ?", menuID).Update("is_available", false)
	for i, it := range items {
		var existing model.MenuItem
		err := s.db.WithContext(ctx).Where("menu_id = ? AND name = ?", menuID, it.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			existing = model.MenuItem{MenuID: menuID, Name: it.Name, IsAvailable: true}
		}
		existing.Price = it.Price
		existing.Description = it.Description
		existing.SortOrder = i
		existing.IsAvailable = true
		if existing.ID == 0 {
			err = s.db.WithContext(ctx).Create(&existing).Error
		} else {
			err = s.db.WithContext(ctx).Save(&existing).Error
		}
		if err != nil {
			return nil, err
		}
	}
	return s.GetByID(ctx, menuID)
}

func (s *MenuService) Publish(ctx context.Context, sellerID, menuID uint64) (*model.Menu, error) {
	menu, err := s.GetByID(ctx, menuID)
	if err != nil || menu.SellerID != sellerID {
		return nil, ErrMenuNotFound
	}
	var count int64
	s.db.WithContext(ctx).Model(&model.MenuItem{}).Where("menu_id = ? AND is_available = ?", menuID, true).Count(&count)
	if count == 0 {
		return nil, fmt.Errorf("no available dishes")
	}
	menu.Status = model.MenuPublished
	menu.UpdatedAt = time.Now()
	return menu, s.db.WithContext(ctx).Save(menu).Error
}

func (s *MenuService) DisableItem(ctx context.Context, sellerID, menuID, itemID uint64) error {
	menu, err := s.GetByID(ctx, menuID)
	if err != nil || menu.SellerID != sellerID {
		return ErrMenuNotFound
	}
	return s.db.WithContext(ctx).Model(&model.MenuItem{}).Where("id = ? AND menu_id = ?", itemID, menuID).Update("is_available", false).Error
}
