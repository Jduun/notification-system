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

// GetNotificationByID godoc
// @Summary Get a notification by its ID
// @Description Get a notification by its ID
// @Tags notifications
// @Param id path string true "Notification UUID"
// @Produce json
// @Success 200 {object} dto.Notification
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/notifications/{id} [get]
func (h *NotificationHTTPHandlers) GetNotificationByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID"})
		return
	}

	notification, err := h.notificationService.GetNotificationByID(c, id)
	if err != nil {
		if errors.Is(err, services.ErrNotificationNotFound) {
			c.IndentedJSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, *notification)
}

// GetNewNotifications godoc
// @Summary Get new notifications
// @Description Get a limited number of the notifications with pending status
// @Tags notifications
// @Param limit query int false "Limit of notifications to return" default(50)
// @Produce json
// @Success 200 {array} dto.Notification
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/notifications/new [get]
func (h *NotificationHTTPHandlers) GetNewNotifications(c *gin.Context) {
	limitStr := c.Query("limit")
	const defaultLimit = 50
	limit := defaultLimit
	var err error

	if limitStr != "" {
		limit, err = strconv.Atoi(c.Query("limit"))
		if err != nil || limit < 0 {
			c.IndentedJSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid limit value"})
			return
		}
	}

	notifications, err := h.notificationService.GetNewNotifications(c, uint(limit))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, notifications)
}

// GetNotificationsByIDs godoc
// @Summary Get multiple notifications by their IDs
// @Description Get notifications using a comma-separated list of UUIDs
// @Tags notifications
// @Param ids query string true "Comma-separated list of notification UUIDs"
// @Produce json
// @Success 200 {array} dto.Notification
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/notifications/batch [get]
func (h *NotificationHTTPHandlers) GetNotificationsByIDs(c *gin.Context) {
	idsString := c.Query("ids")
	idsSeparated := strings.Split(idsString, ",")
	ids := make([]uuid.UUID, len(idsSeparated))
	for i, idStr := range idsSeparated {
		var err error
		ids[i], err = uuid.Parse(idStr)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID"})
			return
		}
	}
	notifications, err := h.notificationService.GetNotificationsByIDs(c, ids)
	if err != nil {
		if errors.Is(err, services.ErrTooManyRequestedNotifications) {
			c.IndentedJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		} else if errors.Is(err, services.ErrNotificationNotFound) {
			c.IndentedJSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, notifications)
}

// CreateNotifications godoc
// @Summary Create multiple notifications
// @Description Accepts a list of notifications to create
// @Tags notifications
// @Accept json
// @Produce json
// @Param notifications body []dto.NotificationCreate true "Data to create notifications"
// @Success 200 {array} string
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/notifications [post]
func (h *NotificationHTTPHandlers) CreateNotifications(c *gin.Context) {
	const op = "handlers.CreateNotifications"
	logger := slogger.GetLoggerFromContext(c).With(slog.String("op", op))

	logger.Info("start to create notifications")
	var notificationsCreate []dto.NotificationCreate
	if err := c.ShouldBindJSON(&notificationsCreate); err != nil {
		c.IndentedJSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	notifications := make([]*dto.NotificationCreate, len(notificationsCreate))
	for i := range notificationsCreate {
		notifications[i] = &notificationsCreate[i]
	}
	IDs, err := h.notificationService.CreateNotifications(c, notifications)
	if err != nil {
		if errors.Is(err, services.ErrTooManyNotificationsToCreate) {
			c.IndentedJSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		logger.Error("failed to create notifications", slog.Any("error", err))
		c.IndentedJSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	logger.Info("notifications created successfully")

	c.IndentedJSON(http.StatusOK, IDs)
}
