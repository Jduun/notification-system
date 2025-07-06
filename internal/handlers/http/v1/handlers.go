package http

import "github.com/gin-gonic/gin"

type NotificationHandlers interface {
	GetNotificationByID(c *gin.Context)
	GetNewNotifications(c *gin.Context)
	GetNotificationsByIDs(c *gin.Context)
	SendNotification(c *gin.Context)
	SendNotifications(c *gin.Context)
}
