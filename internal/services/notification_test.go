package services

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"go.uber.org/mock/gomock"

	"notification_system/internal/dto"
	mocks_repositories "notification_system/internal/repositories/mocks"
)

func TestNotificationServiceImpl_CreateNotifications(t *testing.T) {
	type args struct {
		ctx           context.Context
		notifications []*dto.NotificationCreate
	}
	notificationCount := 5
	notifications := make([]*dto.NotificationCreate, notificationCount)
	for i := range notifications {
		notifications[i] = &dto.NotificationCreate{
			DeliveryType: "test",
			Recipient:    gofakeit.Email(),
			Content:      gofakeit.Sentence(5),
		}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"base test",
			args{context.Background(), notifications},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockRepo := mocks_repositories.NewMockNotificationRepository(ctrl)
			mockRepo.
				EXPECT().
				CreateNotifications(tt.args.ctx, gomock.Any()).
				Return(nil).
				MaxTimes(1)
			s := &NotificationServiceImpl{
				notificationRepo: mockRepo,
			}
			_, err := s.CreateNotifications(tt.args.ctx, tt.args.notifications)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateNotifications() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
