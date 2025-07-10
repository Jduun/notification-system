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
	_ "notification_system/docs"
	"notification_system/internal/messaging"
	"notification_system/migrations"
	"notification_system/pkg/database"
	"notification_system/pkg/logger"
	"notification_system/pkg/server"
)

func main() {
	cfg := config.MustLoad()
	slogger.SetLogger(cfg.AppEnv)

	db := database.New(cfg.GetDBURL())
	migrations.Migrate(cfg.GetDBURL())

	srv := server.NewGinServer(cfg, db)
	go func() {
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Gin server error", slog.Any("error", err))
		}
	}()

	sender := messaging.NewNotificationSender(cfg, db)
	ctxSender, cancelSender := context.WithCancel(context.Background())
	sender.StartProcessNotifications(ctxSender, time.Duration(cfg.SenderHandlePeriodMs)*time.Millisecond)

	receiver := messaging.NewNotificationReceiver(cfg, db)
	ctxReceiver, cancelReceiver := context.WithCancel(context.Background())
	receiver.StartProcessNotifications(ctxReceiver)

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancelSender()
	cancelReceiver()
	ctxShutdown, cancelShutdown := context.WithCancel(context.Background())
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		slog.Error("Error during server shutdown", slog.Any("error", err))
	}
	sender.Close()
	if err := receiver.Close(); err != nil {
		slog.Error("Error during receiver shutdown", slog.Any("error", err))
	}
	db.Pool.Close()
	slog.Info("App gracefully stopped")
}
