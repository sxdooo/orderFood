package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/orderfood/server/internal/config"
	"github.com/orderfood/server/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Bind seller role to a phone number.
// Usage: go run ./scripts/init_seller_phone.go <phone> [shop_name]
func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) < 2 {
		log.Fatal("usage: init_seller_phone <phone> [shop_name]")
	}
	phone := os.Args[1]
	shopName := "我的店铺"
	if len(os.Args) > 2 {
		shopName = os.Args[2]
	}

	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := model.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}

	sp := model.SellerPhone{Phone: phone, ShopName: shopName}
	if err := db.Where("phone = ?", phone).Assign(sp).FirstOrCreate(&sp).Error; err != nil {
		log.Fatal(err)
	}
	log.Printf("seller phone registered: %s (%s). User with this phone becomes seller after profile save.", phone, shopName)
}
