package notify

import (
	"context"
	"fmt"
	"log/slog"
)

// Service sends push notifications to devices via APNs.
// APNs integration is stubbed; actual sending will be wired in a future slice.
type Service struct {
	apnsKeyID   string
	apnsKeyPath string
	appleTeamID string
	bundleID    string
}

// NewService creates a new notify Service.
func NewService(apnsKeyID, apnsKeyPath, appleTeamID, bundleID string) *Service {
	return &Service{
		apnsKeyID:   apnsKeyID,
		apnsKeyPath: apnsKeyPath,
		appleTeamID: appleTeamID,
		bundleID:    bundleID,
	}
}

// Send dispatches a push notification to the specified device.
// Currently stubbed — logs the notification and returns nil.
func (s *Service) Send(ctx context.Context, n Notification) error {
	slog.InfoContext(ctx, "notify: send (stubbed)",
		"deviceToken", n.DeviceToken,
		"title", n.Title,
		"body", n.Body,
	)

	if n.DeviceToken == "" {
		return fmt.Errorf("notify: device token must not be empty")
	}

	// TODO: implement APNs HTTP/2 push via apns2 or sideshow/apns2
	return nil
}
