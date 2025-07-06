package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification_system/config"
	"notification_system/internal/messaging"
	"notification_system/migrations"
	"notification_system/pkg/database"
	"notification_system/pkg/logger"
	"notification_system/pkg/server"
)

func main() {
	cfg := config.MustLoad()
	slogger.SetLogger(cfg.AppEnv)
	db := database.New(cfg.GetDatabaseURL())
	migrations.Migrate(cfg.GetDatabaseURL())
	srv := server.NewGinServer(cfg, db)
	go func() {
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Gin server error", slog.Any("error", err))
		}
	}()
	sender := messaging.NewSender(cfg.NotificationTopicName, cfg, db)
	ctxSender, cancelSender := context.WithCancel(context.Background())
	sender.StartProcessNotifications(ctxSender, time.Duration(cfg.SenderHandlePeriodSeconds)*time.Second)

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancelSender()
	ctxShutdown, cancelShutdown := context.WithCancel(context.Background())
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		slog.Error("Error during server shutdown", slog.Any("error", err))
	}
	sender.Close()
	db.Pool.Close()
	slog.Info("App gracefully stopped")
}
