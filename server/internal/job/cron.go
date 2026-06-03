package job

import (
	"context"
	"log"
	"time"

	"github.com/orderfood/server/internal/service"
)

func StartCron(cutoff *service.CutoffService, sellerRole *service.SellerRoleService) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			ctx := context.Background()
			if err := cutoff.Tick(ctx); err != nil {
				log.Printf("cutoff tick: %v", err)
			}
			if err := sellerRole.EnsureMenuNotExpired(ctx); err != nil {
				log.Printf("menu expire: %v", err)
			}
		}
	}()
}
