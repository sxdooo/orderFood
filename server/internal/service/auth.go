package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/orderfood/server/internal/config"
	"github.com/orderfood/server/internal/model"
	"gorm.io/gorm"
)

var ErrTokenBlacklisted = errors.New("token blacklisted")

type AuthService struct {
	db    *gorm.DB
	redis RedisClient
	cfg   *config.Config
}

type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

type jwtClaims struct {
	UserID uint64         `json:"userId"`
	Role   model.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type WeChatSession struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func NewAuthService(db *gorm.DB, redis RedisClient, cfg *config.Config) *AuthService {
	return &AuthService{db: db, redis: redis, cfg: cfg}
}

// WeChatLogin authenticates via code from wx.login(), and optionally upgrades the
// user to seller role when phoneCode (from getPhoneNumber button) is provided and
// the resolved phone number exists in the seller_phones whitelist.
// Returns (user, token, resolvedPhone, error); resolvedPhone is empty in dev mode.
func (s *AuthService) WeChatLogin(ctx context.Context, code, phoneCode string) (*model.User, string, string, error) {
	session, err := s.exchangeWeChatCode(code)
	if err != nil {
		return nil, "", "", err
	}
	if session.OpenID == "" {
		return nil, "", "", fmt.Errorf("wechat login failed: %s", session.ErrMsg)
	}

	var user model.User
	err = s.db.WithContext(ctx).Where("open_id = ?", session.OpenID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = model.User{OpenID: session.OpenID, Role: model.RoleBuyer}
		if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
			return nil, "", "", err
		}
	} else if err != nil {
		return nil, "", "", err
	}

	// Try to determine seller role from phone whitelist when phoneCode is provided.
	resolvedPhone := ""
	if phoneCode != "" {
		if phone, pErr := s.getPhoneByCode(ctx, phoneCode); pErr == nil && phone != "" {
			resolvedPhone = phone
			var sp model.SellerPhone
			if dbErr := s.db.WithContext(ctx).Where("phone = ?", phone).First(&sp).Error; dbErr == nil {
				// Phone is in whitelist — ensure the user carries seller role.
				if user.Role != model.RoleSeller {
					s.db.WithContext(ctx).Model(&user).Update("role", model.RoleSeller)
					user.Role = model.RoleSeller
				}
			} else if user.Role == model.RoleSeller {
				// Phone no longer in whitelist — downgrade.
				s.db.WithContext(ctx).Model(&user).Update("role", model.RoleBuyer)
				user.Role = model.RoleBuyer
			}
		}
	}

	token, err := s.issueToken(user)
	if err != nil {
		return nil, "", "", err
	}
	return &user, token, resolvedPhone, nil
}

// getPhoneByCode exchanges a getPhoneNumber button code for the user's phone number
// via WeChat's cloud API. Returns empty string in dev mode (no appid configured).
func (s *AuthService) getPhoneByCode(ctx context.Context, phoneCode string) (string, error) {
	if s.cfg.WeChatAppID == "" || s.cfg.WeChatAppSecret == "" {
		return "", nil
	}
	accessToken, err := s.getAccessToken(ctx)
	if err != nil {
		return "", err
	}
	endpoint := fmt.Sprintf(
		"https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s",
		url.QueryEscape(accessToken),
	)
	body, _ := json.Marshal(map[string]string{"code": phoneCode})
	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var result struct {
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
		PhoneInfo struct {
			PhoneNumber     string `json:"phoneNumber"`
			PurePhoneNumber string `json:"purePhoneNumber"`
		} `json:"phone_info"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("getPhoneNumber failed: %s", result.ErrMsg)
	}
	phone := strings.TrimSpace(result.PhoneInfo.PurePhoneNumber)
	return phone, nil
}

// getAccessToken fetches (and caches in Redis) the WeChat access token.
func (s *AuthService) getAccessToken(ctx context.Context) (string, error) {
	const cacheKey = "wechat:access_token"
	if cached, err := s.redis.Get(ctx, cacheKey); err == nil && cached != "" {
		return cached, nil
	}
	endpoint := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		url.QueryEscape(s.cfg.WeChatAppID),
		url.QueryEscape(s.cfg.WeChatAppSecret),
	)
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("get access token failed: %s", result.ErrMsg)
	}
	// Cache slightly shorter than actual expiry to avoid edge cases.
	ttl := time.Duration(result.ExpiresIn-60) * time.Second
	_ = s.redis.Set(ctx, cacheKey, result.AccessToken, ttl)
	return result.AccessToken, nil
}

func (s *AuthService) exchangeWeChatCode(code string) (*WeChatSession, error) {
	if s.cfg.WeChatAppID == "" || s.cfg.WeChatAppSecret == "" {
		return &WeChatSession{OpenID: "dev_" + code}, nil
	}
	endpoint := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(s.cfg.WeChatAppID),
		url.QueryEscape(s.cfg.WeChatAppSecret),
		url.QueryEscape(code),
	)
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var session WeChatSession
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *AuthService) issueToken(user model.User) (string, error) {
	claims := jwtClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        fmt.Sprintf("%d-%d", user.ID, time.Now().UnixNano()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) ParseToken(ctx context.Context, tokenStr string) (*model.User, error) {
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.ID != "" {
		blacklisted, _ := s.redis.Get(ctx, "jwt:blacklist:"+claims.ID)
		if blacklisted != "" {
			return nil, ErrTokenBlacklisted
		}
	}
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, claims.UserID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) Logout(ctx context.Context, tokenStr string) error {
	claims := &jwtClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return err
	}
	if claims.ID == "" || claims.ExpiresAt == nil {
		return nil
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}
	return s.redis.Set(ctx, "jwt:blacklist:"+claims.ID, "1", ttl)
}
