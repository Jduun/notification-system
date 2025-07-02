package main

import (
	"notification_system/config"
	"notification_system/migrations"
	"notification_system/pkg/database"
	"notification_system/pkg/logger"
	"notification_system/pkg/server"
)

func main() {
	cfg := config.MustLoad()
	slogger.SetLogger(cfg.AppEnv)
	db := database.New(cfg.GetDbUrl())
	migrations.Migrate()
	server.NewGinServer(cfg, db).Start()
}
