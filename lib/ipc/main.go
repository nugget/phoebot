package ipc

import (
	"github.com/nugget/phoebot/models"

	"github.com/Tnze/go-mc/chat"
	"github.com/sirupsen/logrus"
)

type SubscriptionChannel struct {
	Operation string
	UserID    string
	Sub       models.Subscription
}

var (
	SubStream           chan SubscriptionChannel
	AnnounceStream      chan models.Product
	MojangStream        chan models.Article
	MsgStream           chan models.DiscordMessage
	ServerChatStream    chan chat.Message
	ServerWhisperStream chan models.Whisper
	ServerSayStream     chan string
)

func InitStreams() error {
	logrus.Debug("Initializing ipc streams")

	SubStream = make(chan SubscriptionChannel)
	MsgStream = make(chan models.DiscordMessage)
	AnnounceStream = make(chan models.Product)
	MojangStream = make(chan models.Article)

	ServerChatStream = make(chan chat.Message)
	ServerSayStream = make(chan string)
	ServerWhisperStream = make(chan models.Whisper)

	return nil
}
