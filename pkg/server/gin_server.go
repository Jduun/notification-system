package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"notification_system/config"
	"notification_system/internal/handlers/http/v1"
	"notification_system/internal/repositories"
	"notification_system/internal/services"
	"notification_system/pkg/database"

	"github.com/gin-gonic/gin"
)

type ginServer struct {
	router     *gin.Engine
	db         *database.PostgresDatabase
	cfg        *config.Config
	httpServer *http.Server
}

func NewGinServer(cfg *config.Config, db *database.PostgresDatabase) *ginServer {
	switch cfg.AppEnv {
	case config.Local, config.Dev:
		gin.SetMode(gin.DebugMode)
	case config.Prod:
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	notificationRepo := repositories.NewNotificationPostgresRepository(db)
	notificationService := services.NewNotificationServiceImpl(notificationRepo)
	notificationHandlers := v1.NewNotificationHTTPHandlers(notificationService)

	notificationRoutes := router.Group(
		"/notifications",
		v1.RequestIDMiddleware(),
		v1.SetLoggerMiddleware(),
	)
	notificationRoutes.GET("/new", notificationHandlers.GetNewNotifications)
	notificationRoutes.GET("/batch", notificationHandlers.GetNotificationsByIDs)
	notificationRoutes.GET("/:id", notificationHandlers.GetNotificationByID)
	notificationRoutes.POST("/", notificationHandlers.CreateNotifications)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.AppPort),
		Handler: router,
	}

	return &ginServer{
		router:     router,
		db:         db,
		cfg:        cfg,
		httpServer: httpServer,
	}
}
func (s *ginServer) Run() error {
	slog.Info("Starting Gin server")
	return s.httpServer.ListenAndServe()
}

func (s *ginServer) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down Gin server...")
	return s.httpServer.Shutdown(ctx)
}
