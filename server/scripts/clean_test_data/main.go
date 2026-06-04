package main

// Clean all mock/seed data created by scripts/seed_test_data.
//
// Removes, inside a single transaction:
//   - Test orders (order_no LIKE 'TEST%' or placed by fake buyers) and their
//     order_items / order_messages / order_read_markers.
//   - Fake buyer users (open_id LIKE 'fake_buyer_openid_%') and their profiles.
//   - Seeded menus (identified by the seeded dish names) and their items.
//   - All delivery_routes (regenerated on demand; safe to wipe before launch).
//
// Real data is preserved: the seller account, seller_phones, real orders
// (order_no prefix 'OF'), and any menus the seller created in the app.
//
// Usage: go run ./scripts/clean_test_data/main.go

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/orderfood/server/internal/config"
	"github.com/orderfood/server/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Dish names seeded by scripts/seed_test_data. Used to locate seeded menus.
var seededDishNames = []string{
	"红烧肉盖饭", "番茄鸡蛋面", "麻婆豆腐套餐", "蒸蛋羹",
}

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("load config:", err)
	}

	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("connect db:", err)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// 1. Collect fake buyer user ids.
		var fakeUserIDs []uint64
		tx.Model(&model.User{}).
			Where("open_id LIKE ?", "fake_buyer_openid_%").
			Pluck("id", &fakeUserIDs)

		// 2. Collect test order ids (TEST prefix OR placed by a fake buyer).
		var testOrderIDs []uint64
		q := tx.Model(&model.Order{}).Where("order_no LIKE ?", "TEST%")
		if len(fakeUserIDs) > 0 {
			q = q.Or("buyer_id IN ?", fakeUserIDs)
		}
		q.Pluck("id", &testOrderIDs)

		// 3. Delete everything hanging off those orders.
		if len(testOrderIDs) > 0 {
			d := tx.Where("order_id IN ?", testOrderIDs).Delete(&model.OrderItem{})
			log.Printf("deleted order_items: %d", d.RowsAffected)

			d = tx.Where("order_id IN ?", testOrderIDs).Delete(&model.OrderMessage{})
			log.Printf("deleted order_messages: %d", d.RowsAffected)

			d = tx.Where("order_id IN ?", testOrderIDs).Delete(&model.OrderReadMarker{})
			log.Printf("deleted order_read_markers: %d", d.RowsAffected)

			// Orders use soft delete; Unscoped() hard-deletes so they're truly gone.
			d = tx.Unscoped().Where("id IN ?", testOrderIDs).Delete(&model.Order{})
			log.Printf("deleted orders: %d", d.RowsAffected)
		} else {
			log.Println("no test orders found")
		}

		// 4. Delete fake buyers and their profiles.
		if len(fakeUserIDs) > 0 {
			d := tx.Where("user_id IN ?", fakeUserIDs).Delete(&model.BuyerProfile{})
			log.Printf("deleted buyer_profiles: %d", d.RowsAffected)

			d = tx.Unscoped().Where("id IN ?", fakeUserIDs).Delete(&model.User{})
			log.Printf("deleted fake buyer users: %d", d.RowsAffected)
		} else {
			log.Println("no fake buyer users found")
		}

		// 5. Delete seeded menus (matched by seeded dish names) and their items.
		var seededMenuIDs []uint64
		tx.Model(&model.MenuItem{}).
			Where("name IN ?", seededDishNames).
			Distinct().
			Pluck("menu_id", &seededMenuIDs)
		if len(seededMenuIDs) > 0 {
			d := tx.Where("menu_id IN ?", seededMenuIDs).Delete(&model.MenuItem{})
			log.Printf("deleted menu_items (seeded menus): %d", d.RowsAffected)

			d = tx.Unscoped().Where("id IN ?", seededMenuIDs).Delete(&model.Menu{})
			log.Printf("deleted seeded menus: %d", d.RowsAffected)
		} else {
			log.Println("no seeded menus found")
		}

		// 6. Wipe delivery routes (regenerated from live orders on demand).
		d := tx.Where("1 = 1").Delete(&model.DeliveryRoute{})
		log.Printf("deleted delivery_routes: %d", d.RowsAffected)

		return nil
	})
	if err != nil {
		log.Fatal("cleanup failed (rolled back):", err)
	}

	log.Println("\nmock data cleanup complete.")
}
