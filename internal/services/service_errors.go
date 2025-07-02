package services

import "errors"

var (
	ErrNotificationNotFound          = errors.New("notification not found")
	ErrCannotGetNotificationByID     = errors.New("cannot get notification by ID")
	ErrCannotGetNotifications        = errors.New("cannot get notifications")
	ErrCannotGetNotificationsByIDs   = errors.New("cannot get notifications by IDs")
	ErrCannotSendNotification        = errors.New("cannot send notification")
	ErrCannotSendNotifications       = errors.New("cannot send notifications")
	ErrTooManyRequestedNotifications = errors.New("too many requested notifications")
	ErrTooManyNotificationsToSend    = errors.New("too many notifications to send")
)
