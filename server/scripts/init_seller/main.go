package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/orderfood/server/internal/config"
	"github.com/orderfood/server/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Initialize first seller user and shop profile.
// Usage: go run ./scripts/init_seller.go <open_id> <shop_name> <address> <lat> <lng>
func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) < 6 {
		log.Fatal("usage: init_seller <open_id> <shop_name> <address> <lat> <lng>")
	}
	openID := os.Args[1]
	shopName := os.Args[2]
	address := os.Args[3]
	var lat, lng float64
	if _, err := fmt.Sscanf(os.Args[4], "%f", &lat); err != nil {
		log.Fatal(err)
	}
	if _, err := fmt.Sscanf(os.Args[5], "%f", &lng); err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := model.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}

	var user model.User
	if err := db.Where("open_id = ?", openID).First(&user).Error; err != nil {
		user = model.User{OpenID: openID, Role: model.RoleSeller}
		if err := db.Create(&user).Error; err != nil {
			log.Fatal(err)
		}
	} else {
		user.Role = model.RoleSeller
		if err := db.Save(&user).Error; err != nil {
			log.Fatal(err)
		}
	}

	profile := model.SellerProfile{
		UserID:     user.ID,
		ShopName:   shopName,
		Address:    address,
		AddressLat: lat,
		AddressLng: lng,
		UpdatedAt:  time.Now(),
	}
	if err := db.Where("user_id = ?", user.ID).Assign(profile).FirstOrCreate(&profile).Error; err != nil {
		log.Fatal(err)
	}
	log.Printf("seller initialized: user_id=%d open_id=%s shop=%s", user.ID, openID, shopName)
}
