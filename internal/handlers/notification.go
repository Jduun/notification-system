package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"notification_system/internal/dto"
	"notification_system/internal/services"
	"notification_system/pkg/logger"
)

type NotificationHTTPHandlers struct {
	notificationService services.NotificationService
}

func NewNotificationHTTPHandlers(notificationService services.NotificationService) NotificationHandlers {
	return &NotificationHTTPHandlers{notificationService: notificationService}
}

func (h *NotificationHTTPHandlers) GetNotificationByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	notification, err := h.notificationService.GetNotificationByID(c, id)
	if err != nil {
		if errors.Is(err, services.ErrNotificationNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"notification": notification})
}

func (h *NotificationHTTPHandlers) GetNotifications(c *gin.Context) {
	cursorCreatedAtStr := c.Query("cursor_created_at")
	cursorIDStr := c.Query("cursor_id")
	limitStr := c.Query("limit")

	cursorCreatedAt := time.Time{}
	cursorID := uuid.Nil
	var err error

	const defaultLimit = 50
	limit := defaultLimit

	if cursorCreatedAtStr != "" {
		cursorCreatedAt, err = time.Parse(time.RFC3339, c.Query("cursor_created_at"))
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid cursor creation date"})
			return
		}
	}
	if cursorIDStr != "" {
		cursorID, err = uuid.Parse(c.Query("cursor_id"))
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid cursor ID"})
			return
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(c.Query("limit"))
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value"})
			return
		}
	}

	notifications, err := h.notificationService.GetNotifications(c, cursorCreatedAt, cursorID, limit)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"notifications": notifications})
}

func (h *NotificationHTTPHandlers) GetNotificationsByIDs(c *gin.Context) {
	idsString := c.Query("ids")
	idsSeparated := strings.Split(idsString, ",")
	ids := make([]uuid.UUID, len(idsSeparated))
	for i, idStr := range idsSeparated {
		var err error
		ids[i], err = uuid.Parse(idStr)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
	}
	notifications, err := h.notificationService.GetNotificationsByIDs(c, ids)
	if err != nil {
		if errors.Is(err, services.ErrTooManyRequestedNotifications) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"notifications": notifications})
}

func (h *NotificationHTTPHandlers) SendNotification(c *gin.Context) {
	logger := slogger.GetLoggerFromContext(c).With("handler", "SendNotification")

	var notification dto.NotificationCreate
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	logger.Info("sending notification")

	id, err := h.notificationService.SendNotification(c, &notification)
	if err != nil {
		if errors.Is(err, services.ErrTooManyNotificationsToSend) {
			logger.Warn("too many notifications", slog.Any("error", err))
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		logger.Error("failed to send notification", slog.Any("error", err))
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("notification sent successfully", slog.Any("id", id))
	c.IndentedJSON(http.StatusOK, gin.H{"id": id})
}

func (h *NotificationHTTPHandlers) SendNotifications(c *gin.Context) {
	var notificationsCreate []dto.NotificationCreate
	if err := c.ShouldBindJSON(&notificationsCreate); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	notifications := make([]*dto.NotificationCreate, len(notificationsCreate))
	for i := range notificationsCreate {
		notifications[i] = &notificationsCreate[i]
	}
	ids, err := h.notificationService.SendNotifications(c, notifications)
	if err != nil {
		if errors.Is(err, services.ErrTooManyNotificationsToSend) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"ids": ids})
}
