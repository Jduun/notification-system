package server

import (
	"fmt"
	"log/slog"

	"notification_system/config"
	"notification_system/internal/handlers"
	"notification_system/internal/repositories"
	"notification_system/internal/services"
	"notification_system/pkg/database"

	"github.com/gin-gonic/gin"
)

type ginServer struct {
	router *gin.Engine
	db     *database.PostgresDatabase
	cfg    *config.Config
}

func NewGinServer(cfg *config.Config, db *database.PostgresDatabase) Server {
	switch config.Cfg.AppEnv {
	case config.Local, config.Dev:
		gin.SetMode(gin.DebugMode)
	case config.Prod:
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	return &ginServer{
		router: router,
		db:     db,
		cfg:    cfg,
	}
}

func (s *ginServer) Start() {
	notificationRepo := repositories.NewNotificationPostgresRepository(s.db)
	notificationService := services.NewNotificationServiceImpl(notificationRepo)
	notificationHandlers := handlers.NewNotificationHTTPHandlers(notificationService)

	notificationRoutes := s.router.Group(
		"/notifications",
		handlers.RequestIDMiddleware(),
		handlers.SetLoggerMiddleware(),
	)
	notificationRoutes.GET("/", notificationHandlers.GetNotifications)
	notificationRoutes.GET("/:id", notificationHandlers.GetNotificationByID)
	notificationRoutes.GET("/batch", notificationHandlers.GetNotificationsByIDs)
	notificationRoutes.POST("/", notificationHandlers.SendNotification)
	notificationRoutes.POST("/batch", notificationHandlers.SendNotifications)

	slog.Info("starting gin server")
	err := s.router.Run(fmt.Sprintf(":%s", s.cfg.AppPort))
	if err != nil {
		slog.Error("cannot run gin server: %s", err)
	}
}
