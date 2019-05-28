package ipc

import (
	"github.com/nugget/phoebot/models"
	"github.com/sirupsen/logrus"
)

type SubscriptionChannel struct {
	Operation string
	UserID    string
	Sub       models.Subscription
}

var (
	SubStream      chan SubscriptionChannel
	AnnounceStream chan models.Product
	MojangStream   chan models.Article
	MsgStream      chan models.DiscordMessage
)

func InitStreams() error {
	logrus.Debug("Initializing ipc streams")

	SubStream = make(chan SubscriptionChannel)
	MsgStream = make(chan models.DiscordMessage)
	AnnounceStream = make(chan models.Product)
	MojangStream = make(chan models.Article)

	return nil
}
