package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env      string
	HTTPPort string

	MySQLDSN string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	JWTSecret     string
	JWTExpiration time.Duration

	WeChatAppID     string
	WeChatAppSecret string

	AmapAPIKey string

	OSSEndpoint        string
	OSSAccessKeyID     string
	OSSAccessKeySecret string
	OSSBucket          string

	WeChatMchID      string
	WeChatMchSerial  string
	WeChatAPIv3Key   string
	WeChatPrivateKey string
	WeChatNotifyURL  string

	Timezone string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtHours := getEnvInt("JWT_EXPIRATION_HOURS", 168)
	return &Config{
		Env:                getEnv("APP_ENV", "development"),
		HTTPPort:           getEnv("HTTP_PORT", "8080"),
		MySQLDSN:           getEnv("MYSQL_DSN", "root:root@tcp(localhost:3306)/orderfood?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            getEnvInt("REDIS_DB", 0),
		JWTSecret:          getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiration:      time.Duration(jwtHours) * time.Hour,
		WeChatAppID:        getEnv("WECHAT_APP_ID", ""),
		WeChatAppSecret:    getEnv("WECHAT_APP_SECRET", ""),
		AmapAPIKey:         getEnv("AMAP_API_KEY", ""),
		OSSEndpoint:        getEnv("OSS_ENDPOINT", ""),
		OSSAccessKeyID:     getEnv("OSS_ACCESS_KEY_ID", ""),
		OSSAccessKeySecret: getEnv("OSS_ACCESS_KEY_SECRET", ""),
		OSSBucket:          getEnv("OSS_BUCKET", ""),
		WeChatMchID:        getEnv("WECHAT_MCH_ID", ""),
		WeChatMchSerial:    getEnv("WECHAT_MCH_SERIAL", ""),
		WeChatAPIv3Key:     getEnv("WECHAT_API_V3_KEY", ""),
		WeChatPrivateKey:   getEnv("WECHAT_PRIVATE_KEY_PATH", ""),
		WeChatNotifyURL:    getEnv("WECHAT_NOTIFY_URL", ""),
		Timezone:           getEnv("TZ", "Asia/Shanghai"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
