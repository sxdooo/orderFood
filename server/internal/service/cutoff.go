package service

import (
	"context"
	"fmt"
	"time"

	"github.com/orderfood/server/internal/model"
	"github.com/orderfood/server/internal/pkg/timeutil"
	"gorm.io/gorm"
)

type CutoffService struct {
	db    *gorm.DB
	redis RedisClient
}

func NewCutoffService(db *gorm.DB, redis RedisClient) *CutoffService {
	return &CutoffService{db: db, redis: redis}
}

type CutoffStatus struct {
	OrderDate    string `json:"orderDate"`
	CutoffTime   string `json:"cutoffTime"`
	IsOpen       bool   `json:"isOpen"`
	SecondsLeft  int64  `json:"secondsLeft"`
}

func (s *CutoffService) SetCutoff(ctx context.Context, cutoffTime string) (CutoffStatus, error) {
	today := timeutil.Today()
	setting := model.CutoffSetting{OrderDate: today, CutoffTime: cutoffTime, UpdatedAt: time.Now()}
	err := s.db.WithContext(ctx).Where("order_date = ?", today).Assign(setting).FirstOrCreate(&setting).Error
	if err != nil {
		return CutoffStatus{}, err
	}
	_ = s.redis.Del(ctx, s.redisKey(today))
	return CutoffStatusFromSetting(setting), nil
}

func CutoffStatusFromSetting(s model.CutoffSetting) CutoffStatus {
	open, left := calcCutoffOpen(s.OrderDate, s.CutoffTime)
	return CutoffStatus{
		OrderDate:   timeutil.FormatDate(s.OrderDate),
		CutoffTime:  s.CutoffTime,
		IsOpen:      open,
		SecondsLeft: left,
	}
}

func (s *CutoffService) GetStatus(ctx context.Context) (CutoffStatus, error) {
	today := timeutil.Today()
	cached, _ := s.redis.Get(ctx, s.redisKey(today))
	if cached == "closed" {
		setting, _ := s.getSetting(ctx, today)
		st := CutoffStatus{OrderDate: timeutil.FormatDate(today), CutoffTime: "17:00", IsOpen: false, SecondsLeft: 0}
		if setting != nil {
			st.CutoffTime = setting.CutoffTime
		}
		return st, nil
	}
	setting, err := s.getSetting(ctx, today)
	if err != nil {
		// default open until 17:00
		open, left := calcCutoffOpen(today, "17:00")
		return CutoffStatus{OrderDate: timeutil.FormatDate(today), CutoffTime: "17:00", IsOpen: open, SecondsLeft: left}, nil
	}
	st := CutoffStatusFromSetting(*setting)
	if !st.IsOpen {
		_ = s.redis.Set(ctx, s.redisKey(today), "closed", 24*time.Hour)
	}
	return st, nil
}

func (s *CutoffService) IsOrderingOpen(ctx context.Context) (bool, error) {
	st, err := s.GetStatus(ctx)
	return st.IsOpen, err
}

func (s *CutoffService) Tick(ctx context.Context) error {
	today := timeutil.Today()
	setting, err := s.getSetting(ctx, today)
	if err != nil {
		return nil
	}
	open, _ := calcCutoffOpen(setting.OrderDate, setting.CutoffTime)
	if !open {
		return s.redis.Set(ctx, s.redisKey(today), "closed", 24*time.Hour)
	}
	return s.redis.Del(ctx, s.redisKey(today))
}

func (s *CutoffService) getSetting(ctx context.Context, d time.Time) (*model.CutoffSetting, error) {
	var setting model.CutoffSetting
	err := s.db.WithContext(ctx).Where("order_date = ?", d).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (s *CutoffService) redisKey(d time.Time) string {
	return fmt.Sprintf("cutoff:%s", timeutil.FormatDate(d))
}

func calcCutoffOpen(orderDate time.Time, cutoffTime string) (bool, int64) {
	now := timeutil.Now()
	var h, m int
	fmt.Sscanf(cutoffTime, "%d:%d", &h, &m)
	deadline := time.Date(orderDate.Year(), orderDate.Month(), orderDate.Day(), h, m, 0, 0, orderDate.Location())
	if now.After(deadline) {
		return false, 0
	}
	return true, int64(deadline.Sub(now).Seconds())
}
