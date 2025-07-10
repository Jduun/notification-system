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

	urlSendNotifications := fmt.Sprintf("%s/notifications", host)
	resp, err := http.Post(urlSendNotifications, "application/json", bytes.NewBuffer([]byte(payload)))
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
	var sendNotificationsResponse struct {
		IDs []string `json:"ids"`
	}
	if err = json.Unmarshal(bodyBytes, &sendNotificationsResponse); err != nil {
		t.Errorf("failed to parse response JSON: %v", err)
		return
	}

	deadline := time.Now().Add(10 * time.Second)
	allDelivered := true
	for time.Now().Before(deadline) {
		urlGetNotificationsByIDs := fmt.Sprintf("%s/notifications/batch?ids=", host)
		for i, id := range sendNotificationsResponse.IDs {
			urlGetNotificationsByIDs += id
			if i != notificationCount-1 {
				urlGetNotificationsByIDs += ","
			}
		}
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
		var getNotificationsResponse struct {
			Notifications []dto.NotificationResponse `json:"notifications"`
		}
		if err = json.Unmarshal(bodyBytes, &getNotificationsResponse); err != nil {
			t.Errorf("failed to parse response JSON: %v", err)
			return
		}
		allDelivered = true
		for _, notification := range getNotificationsResponse.Notifications {
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
