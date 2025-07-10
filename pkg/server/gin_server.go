package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"notification_system/config"
	"notification_system/internal/handlers/http/v1"
	"notification_system/internal/repositories"
	"notification_system/internal/services"
	"notification_system/pkg/database"

	"github.com/gin-gonic/gin"
)

type GinServer struct {
	router     *gin.Engine
	db         *database.PostgresDatabase
	cfg        *config.Config
	httpServer *http.Server
}

// @title           Notification System
// @version         1.0
// @description     Fast and reliable notification system.

// @host      localhost:8080
// @BasePath  /api/v1

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func NewGinServer(cfg *config.Config, db *database.PostgresDatabase) *GinServer {
	switch cfg.AppEnv {
	case config.Local, config.Dev:
		gin.SetMode(gin.DebugMode)
	case config.Prod:
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	apiV1 := router.Group("/api/v1")

	notificationRepo := repositories.NewNotificationPostgresRepository(db)
	notificationService := services.NewNotificationServiceImpl(notificationRepo)
	notificationHandlers := v1.NewNotificationHTTPHandlers(notificationService)

	notificationRoutes := apiV1.Group(
		"/notifications",
		v1.RequestIDMiddleware(),
		v1.SetLoggerMiddleware(),
	)
	notificationRoutes.GET("/new", notificationHandlers.GetNewNotifications)
	notificationRoutes.GET("/batch", notificationHandlers.GetNotificationsByIDs)
	notificationRoutes.GET("/:id", notificationHandlers.GetNotificationByID)
	notificationRoutes.POST("/", notificationHandlers.CreateNotifications)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.AppPort),
		Handler: router,
	}

	return &GinServer{
		router:     router,
		db:         db,
		cfg:        cfg,
		httpServer: httpServer,
	}
}

func (s *GinServer) Run() error {
	slog.Info("Starting Gin server")
	return s.httpServer.ListenAndServe()
}

func (s *GinServer) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down Gin server...")
	return s.httpServer.Shutdown(ctx)
}
