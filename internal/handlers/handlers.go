package handlers

import "github.com/gin-gonic/gin"

type NotificationHandlers interface {
	GetNotificationByID(c *gin.Context)
	GetNotifications(c *gin.Context)
	GetNotificationsByIDs(c *gin.Context)
	SendNotification(c *gin.Context)
	SendNotifications(c *gin.Context)
}
