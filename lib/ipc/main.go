package ipc

import (
	"github.com/nugget/phoebot/models"
)

type SubscriptionChannel struct {
	Operation string
	UserID    string
	Sub       models.Subscription
}

var (
	SubStream chan SubscriptionChannel
)

func InitSubStream() error {
	SubStream = make(chan SubscriptionChannel)
	return nil
}
