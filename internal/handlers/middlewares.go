package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	slogger "notification_system/pkg/logger"
)

const RequestIDKey = "RequestID"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set(RequestIDKey, requestID)
		c.Next()
	}
}

func SetLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := slog.Default().With("request_id", c.GetString(RequestIDKey))
		c.Set(slogger.LoggerKey, logger)
		c.Next()
	}
}
