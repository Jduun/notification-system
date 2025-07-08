package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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

func (h *NotificationHTTPHandlers) GetNewNotifications(c *gin.Context) {
	limitStr := c.Query("limit")
	const defaultLimit = 50
	limit := defaultLimit
	var err error

	if limitStr != "" {
		limit, err = strconv.Atoi(c.Query("limit"))
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value"})
			return
		}
	}

	notifications, err := h.notificationService.GetNewNotifications(c, limit)
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

func (h *NotificationHTTPHandlers) CreateNotifications(c *gin.Context) {
	const op = "handlers.CreateNotifications"
	logger := slogger.GetLoggerFromContext(c).With(slog.String("op", op))

	logger.Info("start to create notifications")
	var notificationsCreate []dto.NotificationCreate
	if err := c.ShouldBindJSON(&notificationsCreate); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	notifications := make([]*dto.NotificationCreate, len(notificationsCreate))
	for i := range notificationsCreate {
		notifications[i] = &notificationsCreate[i]
	}
	ids, err := h.notificationService.CreateNotifications(c, notifications)
	if err != nil {
		if errors.Is(err, services.ErrTooManyNotificationsToCreate) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		logger.Error("failed to create notifications", slog.Any("error", err))
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("notifications created successfully")
	c.IndentedJSON(http.StatusOK, gin.H{"ids": ids})
}
