package services

import "errors"

var (
	ErrNotificationNotFound          = errors.New("notification not found")
	ErrCannotGetNotificationByID     = errors.New("cannot get notification by ID")
	ErrCannotGetNotifications        = errors.New("cannot get notifications")
	ErrCannotGetNotificationsByIDs   = errors.New("cannot get notifications by IDs")
	ErrCannotCreateNotifications     = errors.New("cannot create notifications")
	ErrTooManyRequestedNotifications = errors.New("too many requested notifications")
	ErrTooManyNotificationsToCreate  = errors.New("too many notifications to create")
)
