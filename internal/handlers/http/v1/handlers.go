package v1

import (
	"github.com/gin-gonic/gin"
)

type NotificationHandlers interface {
	GetNotificationByID(c *gin.Context)
	GetNewNotifications(c *gin.Context)
	GetNotificationsByIDs(c *gin.Context)
	CreateNotifications(c *gin.Context)
}
