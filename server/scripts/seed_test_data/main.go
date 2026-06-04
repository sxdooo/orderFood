package main

// Seed test data for route planning testing.
// Creates: tomorrow's published menu + mock buyers with orders.
// Usage: go run ./scripts/seed_test_data/main.go
//
// All addresses are in Gu'an County, Langfang City, Hebei Province (GCJ-02 coords).

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/orderfood/server/internal/config"
	"github.com/orderfood/server/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type buyer struct {
	name    string
	phone   string
	address string
	lat     float64
	lng     float64
}

var buyers = []buyer{
	// ── 原有 6 个 ──────────────────────────────────────────────
	{"张伟", "13901001001", "固安县东方大道68号固安新城A区3号楼", 39.4420, 116.2960},
	{"李娜", "13901001002", "固安县兴业路88号金色家园小区2单元", 39.4378, 116.3045},
	{"王芳", "13901001003", "固安县永定路33号阳光花园4栋", 39.4352, 116.2915},
	{"刘洋", "13901001004", "固安县光明街15号碧水湾小区1号楼", 39.4455, 116.2992},
	{"赵磊", "13901001006", "固安县新华路12号翠园小区5单元101", 39.4468, 116.2938},

	// ── 新增 23 个 ─────────────────────────────────────────────
	{"吴军", "13901001007", "廊坊市固安县孔雀大道与永定北路交叉口西300米", 39.4380, 116.3180},
	{"孙梅", "13901001008", "廊坊市固安县东方街2号", 39.4390, 116.2985},
	{"周磊", "13901001009", "廊坊市固安县中央大道与环湖路交叉口西南50米孔雀广场", 39.4350, 116.3150},
	{"吴霞", "13901001010", "廊坊市固安县安康东街与家兴路交叉口南160米永定河孔雀城祥园", 39.4290, 116.3200},
	{"郑伟", "13901001011", "廊坊市固安县永定河孔雀城大卫城六期18号楼", 39.4270, 116.3230},
	{"冯丽", "13901001012", "廊坊市固安县北五里小区6号楼", 39.4510, 116.2975},
	{"陈明", "13901001013", "廊坊市固安县绿宸·凤栖华府一期10号楼", 39.4405, 116.2870},
	{"褚华", "13901001014", "廊坊市固安县方城佳苑13号楼", 39.4365, 116.2935},
	{"卫强", "13901001015", "固安国税小区6号楼", 39.4398, 116.2952},
	{"蒋英", "13901001016", "廊坊市固安县新中东街与永盛路交叉口西140米富丽雅园", 39.4375, 116.2910},
	{"沈勇", "13901001017", "京南绿洲2区14号楼", 39.4445, 116.3025},
	{"韩秀", "13901001018", "金海太阳公园二期北区13号楼", 39.4465, 116.3060},
	{"杨帆", "13901001019", "金海太阳公园二期北区12号楼", 39.4463, 116.3058},
	{"朱燕", "13901001020", "恒基·现代城5号楼", 39.4372, 116.3035},
	{"秦峰", "13901001021", "中鼎凤凰城三期11号楼", 39.4435, 116.2895},
	{"尤薇", "13901001022", "旭景花园8号楼", 39.4342, 116.2972},
	{"许刚", "13901001023", "固安县文教小区3号楼", 39.4418, 116.2942},
	{"何丽", "13901001024", "塔斯汀中国汉堡(廊坊市固安玉井路店)旁育才路3号", 39.4385, 116.3012},
	{"吕强", "13901001025", "育才小区1号楼", 39.4428, 116.2978},
	{"施雨", "13901001026", "金海太阳公园一期8号楼", 39.4472, 116.3042},
	{"张华", "13901001027", "天园小区一期7号楼", 39.4348, 116.2962},
	{"孔明", "13901001028", "天园小区A16号楼", 39.4350, 116.2960},
	{"赵云", "13901001029", "梧桐郡11号楼", 39.4482, 116.2922},
}

type dish struct {
	name  string
	price int64
	desc  string
}

var dishes = []dish{
	{"红烧肉盖饭", 2800, "五花肉慢炖，配米饭"},
	{"番茄鸡蛋面", 1500, "自制番茄底汤"},
	{"麻婆豆腐套餐", 2200, "麻辣口味，含米饭+汤"},
	{"蒸蛋羹", 800, "嫩滑蒸蛋，半份"},
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
	if err := model.AutoMigrate(db); err != nil {
		log.Fatal("migrate:", err)
	}

	// ── 1. Find seller ────────────────────────────────────────
	var seller model.User
	if err := db.Where("role = ?", model.RoleSeller).First(&seller).Error; err != nil {
		log.Fatal("no seller found — please run init_seller_phone first and login once")
	}
	log.Printf("seller: id=%d", seller.ID)

	// Ensure seller profile has coordinates
	var sp model.SellerProfile
	if err := db.Where("user_id = ?", seller.ID).First(&sp).Error; err != nil {
		sp = model.SellerProfile{
			UserID:     seller.ID,
			ShopName:   "固安美食驿站",
			Address:    "固安县政府广场东侧固安县行政服务中心",
			AddressLat: 39.4387,
			AddressLng: 116.2987,
			UpdatedAt:  time.Now(),
		}
		db.Create(&sp)
		log.Println("created seller profile")
	} else if sp.AddressLat == 0 {
		db.Model(&sp).Updates(map[string]interface{}{
			"address":     "固安县政府广场东侧固安县行政服务中心",
			"address_lat": 39.4387,
			"address_lng": 116.2987,
			"shop_name":   "固安美食驿站",
		})
		log.Println("updated seller profile coordinates")
	}

	// ── 2. Create tomorrow's menu ─────────────────────────────
	tomorrow := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)

	var menu model.Menu
	err = db.Where("seller_id = ? AND delivery_date = ?", seller.ID, tomorrow).First(&menu).Error
	if err != nil {
		menu = model.Menu{
			SellerID:     seller.ID,
			DeliveryDate: tomorrow,
			Status:       model.MenuPublished,
		}
		if err := db.Create(&menu).Error; err != nil {
			log.Fatal("create menu:", err)
		}
		for i, d := range dishes {
			db.Create(&model.MenuItem{
				MenuID:      menu.ID,
				Name:        d.name,
				Price:       d.price,
				Description: d.desc,
				SortOrder:   i,
				IsAvailable: true,
			})
		}
		log.Printf("created menu id=%d with %d items", menu.ID, len(dishes))
	} else {
		menu.Status = model.MenuPublished
		db.Save(&menu)
		log.Printf("menu already exists id=%d, set to published", menu.ID)
	}

	// Reload items
	db.Where("menu_id = ? AND is_available = true", menu.ID).Find(&menu.Items)
	if len(menu.Items) == 0 {
		log.Fatal("no menu items found")
	}

	// ── 3. Create buyers + orders ─────────────────────────────
	// Delete all existing test orders for tomorrow first to keep data clean.
	// Test users are identified by their fake_buyer_openid_ prefix.
	var testUsers []model.User
	db.Where("open_id LIKE ?", "fake_buyer_openid_%").Find(&testUsers)
	if len(testUsers) > 0 {
		ids := make([]uint64, len(testUsers))
		for i, u := range testUsers {
			ids[i] = u.ID
		}
		del := db.Where("buyer_id IN ? AND delivery_date = ?", ids, tomorrow).
			Delete(&model.Order{})
		if del.RowsAffected > 0 {
			log.Printf("deleted %d stale test orders for tomorrow", del.RowsAffected)
		}
	}

	created := 0
	for i, b := range buyers {
		openID := fmt.Sprintf("fake_buyer_openid_%d", i+1)

		var user model.User
		if err := db.Where("open_id = ?", openID).First(&user).Error; err != nil {
			user = model.User{OpenID: openID, Role: model.RoleBuyer}
			db.Create(&user)
		}

		lat, lng := b.lat, b.lng
		var bp model.BuyerProfile
		if err := db.Where("user_id = ?", user.ID).First(&bp).Error; err != nil {
			db.Create(&model.BuyerProfile{
				UserID:           user.ID,
				ContactName:      b.name,
				ContactPhone:     b.phone,
				Address:          b.address,
				AddressLat:       &lat,
				AddressLng:       &lng,
				ProfileCompleted: true,
				UpdatedAt:        time.Now(),
			})
		} else {
			// Always sync profile with current buyer data so renames take effect
			db.Model(&bp).Updates(map[string]interface{}{
				"contact_name":  b.name,
				"contact_phone": b.phone,
				"address":       b.address,
				"address_lat":   lat,
				"address_lng":   lng,
			})
		}

		// Distribute dishes: each buyer gets 1–2 items (cycling)
		items := []model.MenuItem{menu.Items[i%len(menu.Items)]}
		if i%3 != 0 && len(menu.Items) > 1 {
			items = append(items, menu.Items[(i+1)%len(menu.Items)])
		}

		total := int64(0)
		for _, it := range items {
			total += it.Price
		}

		orderNo := "TEST" + uuid.New().String()[:8]
		order := model.Order{
			OrderNo:      orderNo,
			BuyerID:      user.ID,
			MenuID:       menu.ID,
			DeliveryDate: tomorrow,
			TotalAmount:  total,
			Status:       model.OrderConfirmed,
			ContactName:  b.name,
			ContactPhone: b.phone,
			Address:      b.address,
			AddressLat:   &lat,
			AddressLng:   &lng,
		}
		if err := db.Create(&order).Error; err != nil {
			log.Printf("WARN create order for %s: %v", b.name, err)
			continue
		}
		for _, it := range items {
			db.Create(&model.OrderItem{
				OrderID:       order.ID,
				MenuItemID:    it.ID,
				NameSnapshot:  it.Name,
				PriceSnapshot: it.Price,
				Quantity:      1,
			})
		}
		created++
		log.Printf("[%d] order %s  %s  %s", i+1, orderNo, b.name, b.address)
	}

	log.Printf("\ndone: created=%d", created)
	log.Println("open seller app -> 订单管理 -> 生成路线 to test routing")
}
