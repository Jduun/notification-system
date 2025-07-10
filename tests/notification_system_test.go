package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"

	"notification_system/internal/dto"
	"notification_system/internal/entities"
)

var (
	host = fmt.Sprintf("http://localhost:%s", os.Getenv("APP_PORT"))
)

func TestNotificationSystem_SendNotifications(t *testing.T) {
	email := gofakeit.Email()
	message := gofakeit.Sentence(5)
	notification := fmt.Sprintf(`{
		"delivery_type": "test",
		"recipient": "%s",
		"content": "%s"
	}`, email, message)
	notificationCount := 10
	payload := "[" + strings.Repeat(notification+",", notificationCount-1) + notification + "]"

	urlCreateNotifications := fmt.Sprintf("%s/api/v1/notifications", host)
	resp, err := http.Post(urlCreateNotifications, "application/json", bytes.NewBuffer([]byte(payload)))
	if err != nil {
		t.Errorf("request failed: %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("request returned status: %s", resp.Status)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("failed to read response body: %v", err)
		return
	}
	var IDs []uuid.UUID
	if err = json.Unmarshal(bodyBytes, &IDs); err != nil {
		t.Errorf("failed to parse response JSON: %v", err)
		return
	}

	urlGetNotificationsByIDs := fmt.Sprintf("%s/api/v1/notifications/batch?ids=", host)
	for i, ID := range IDs {
		urlGetNotificationsByIDs += ID.String()
		if i != notificationCount-1 {
			urlGetNotificationsByIDs += ","
		}
	}
	deadline := time.Now().Add(10 * time.Second)
	allDelivered := true
	for time.Now().Before(deadline) {
		resp, err = http.Get(urlGetNotificationsByIDs)
		if err != nil {
			t.Errorf("request failed: %v", err)
			return
		}
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("failed to read response body: %v", err)
			return
		}
		var notifications []dto.Notification
		if err = json.Unmarshal(bodyBytes, &notifications); err != nil {
			t.Errorf("failed to parse response JSON: %v", err)
			return
		}
		allDelivered = true
		for _, notification := range notifications {
			if notification.Status != entities.StatusDelivered {
				allDelivered = false
				continue
			}
		}
		if allDelivered {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !allDelivered {
		t.Errorf("failed to deliver all notifications")
	}

	err = resp.Body.Close()
	if err != nil {
		t.Errorf("failed to close response body: %v", err)
	}
}
