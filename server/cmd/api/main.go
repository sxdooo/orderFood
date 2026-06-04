package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orderfood/server/internal/config"
	"github.com/orderfood/server/internal/handler"
	"github.com/orderfood/server/internal/job"
	"github.com/orderfood/server/internal/middleware"
	"github.com/orderfood/server/internal/model"
	"github.com/orderfood/server/internal/pkg/amap"
	"github.com/orderfood/server/internal/repository"
	"github.com/orderfood/server/internal/service"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// In production, quiet down: release HTTP mode + warn-only SQL logging.
	gormLogLevel := logger.Info
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
		gormLogLevel = logger.Warn
	}

	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	if err := model.AutoMigrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	redisRepo := repository.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisRepo.Ping(ctx); err != nil {
		log.Printf("warn: redis ping failed: %v", err)
	}

	amapClient := amap.NewClient(cfg)
	sellerRoleSvc := service.NewSellerRoleService(db, amapClient)
	authSvc := service.NewAuthService(db, redisRepo, cfg)
	buyerProfileSvc := service.NewBuyerProfileService(db, amapClient, sellerRoleSvc)
	menuSvc := service.NewMenuService(db)
	cutoffSvc := service.NewCutoffService(db, redisRepo)
	orderSvc := service.NewOrderService(db, menuSvc, cutoffSvc, amapClient)
	routeSvc := service.NewRouteService(db, orderSvc)
	chatSvc := service.NewChatService(db, redisRepo)

	job.StartCron(cutoffSvc, sellerRoleSvc)

	authHandler := handler.NewAuthHandler(authSvc)
	buyerProfileHandler := handler.NewBuyerProfileHandler(buyerProfileSvc)
	menuHandler := handler.NewMenuHandler(menuSvc)
	cutoffHandler := handler.NewCutoffHandler(cutoffSvc)
	orderHandler := handler.NewOrderHandler(orderSvc)
	sellerOrderHandler := handler.NewSellerOrderHandler(orderSvc, sellerRoleSvc)
	routeHandler := handler.NewRouteHandler(routeSvc, amapClient)
	chatHandler := handler.NewChatHandler(chatSvc)
	sellerProfileHandler := handler.NewSellerProfileHandler(sellerRoleSvc)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.CORS())

	r.GET("/health", handler.Health)

	api := r.Group("/api/v1")
	{
		api.POST("/auth/wechat", authHandler.WeChatLogin)
		api.GET("/cutoff/status", cutoffHandler.Status)

		authRequired := api.Group("")
		authRequired.Use(middleware.AuthRequired(authSvc))
		{
			authRequired.POST("/auth/logout", authHandler.Logout)
			authRequired.GET("/auth/me", authHandler.Me)

			authRequired.GET("/buyer/profile", buyerProfileHandler.Get)
			authRequired.PUT("/buyer/profile", buyerProfileHandler.Update)

			authRequired.GET("/menus/tomorrow", menuHandler.Tomorrow)
			authRequired.POST("/orders", orderHandler.Create)
			authRequired.GET("/orders", orderHandler.ListBuyer)
			authRequired.GET("/orders/:id", orderHandler.GetBuyer)
			authRequired.POST("/orders/:id/cancel", orderHandler.Cancel)
			authRequired.GET("/orders/:id/messages", chatHandler.List)
			authRequired.POST("/orders/:id/messages", chatHandler.Send)
			authRequired.GET("/messages/conversations", chatHandler.Conversations)
			authRequired.GET("/messages/unread-count", chatHandler.UnreadCount)

			seller := authRequired.Group("/seller")
			seller.Use(middleware.RequireRole(model.RoleSeller))
			{
				seller.GET("/profile", sellerProfileHandler.Get)
				seller.PUT("/profile", sellerProfileHandler.Update)
				seller.PUT("/cutoff", cutoffHandler.Set)
				seller.GET("/menus", menuHandler.ListSeller)
				seller.POST("/menus", menuHandler.Create)
				seller.PUT("/menus/:id/items", menuHandler.UpdateItems)
				seller.POST("/menus/:id/publish", menuHandler.Publish)
				seller.GET("/orders", sellerOrderHandler.List)
				seller.GET("/orders/:id", sellerOrderHandler.Detail)
				seller.PUT("/orders/:id/status", sellerOrderHandler.UpdateStatus)
				seller.POST("/orders/:id/refund", sellerOrderHandler.Refund)
				seller.GET("/orders/stats", sellerOrderHandler.Stats)
				seller.POST("/routes", routeHandler.Generate)
				seller.GET("/routes", routeHandler.Get)
				seller.PUT("/routes/stops", routeHandler.UpdateStops)
				seller.POST("/routes/cluster", routeHandler.Cluster)
				seller.GET("/routes/cluster", routeHandler.Clusters)
				seller.POST("/routes/directions", routeHandler.Directions)
			}
		}
	}

	go func() {
		log.Printf("server listening on :%s", cfg.HTTPPort)
		if err := r.Run(":" + cfg.HTTPPort); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down")
}
