// Package notify handles push notifications via APNs.
package notify

// Notification is the payload sent to a device via APNs.
type Notification struct {
	DeviceToken string `json:"deviceToken"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	Data        any    `json:"data,omitempty"`
}

// NotificationTopic represents the category of a push notification.
type NotificationTopic string

const (
	TopicTradeReceived  NotificationTopic = "trade.received"
	TopicTradeAccepted  NotificationTopic = "trade.accepted"
	TopicTradeDeclined  NotificationTopic = "trade.declined"
	TopicTradeConfirmed NotificationTopic = "trade.confirmed"
	TopicTradeCompleted NotificationTopic = "trade.completed"
	TopicFriendRequest  NotificationTopic = "friend.request"
	TopicFriendAccepted NotificationTopic = "friend.accepted"
)
