module github.com/nugget/phoebot

go 1.16

// Temporarily override for now, until Tnze merges my PR
//replace github.com/Tnze/go-mc => /Users/nugget/src/go-mc
//replace github.com/Tnze/go-mc => github.com/nugget/go-mc v1.14.5-0.20201118172317-e9cb621f23ef

require (
	github.com/Tnze/go-mc v1.17.1-0.20210806203433-99081e1b9cfb
	github.com/blang/semver v3.5.1+incompatible
	github.com/bwmarrin/discordgo v0.23.2
	github.com/gemnasium/logrus-graylog-hook/v3 v3.0.3
	github.com/go-sql-driver/mysql v1.6.0
	github.com/google/uuid v1.2.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/lib/pq v1.10.2
	github.com/pkg/errors v0.9.1 // indirect
	github.com/seeruk/minecraft-rcon v0.0.0-20190221212056-6ab996d90449
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.8.1
	github.com/tidwall/gjson v1.8.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
)
