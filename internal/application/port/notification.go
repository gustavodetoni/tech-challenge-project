package port

import "context"

type StatusChangedNotification struct {
	ServiceOrderID string
	Code           string
	ClientName     string
	ClientEmail    string
	CurrentStatus  string
}

type ServiceOrderNotifier interface {
	NotifyStatusChanged(ctx context.Context, input StatusChangedNotification) error
}
