module github.com/nugget/phoebot

go 1.16

// Temporarily override for now, until Tnze merges my PR
//replace github.com/Tnze/go-mc => /Users/nugget/src/go-mc
//replace github.com/Tnze/go-mc => github.com/nugget/go-mc v1.14.5-0.20201118172317-e9cb621f23ef

require (
	github.com/Tnze/go-mc v1.17.2-0.20211211093242-38cc60ea6593
	github.com/blang/semver v3.5.1+incompatible
	github.com/bwmarrin/discordgo v0.23.2
	github.com/gemnasium/logrus-graylog-hook/v3 v3.0.3
	github.com/go-sql-driver/mysql v1.6.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/iancoleman/strcase v0.2.0 // indirect
	github.com/lib/pq v1.10.4
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/seeruk/minecraft-rcon v0.0.0-20190221212056-6ab996d90449
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.9.0
	github.com/tidwall/gjson v1.12.1
	golang.org/x/crypto v0.0.0-20211209193657-4570a0811e8b // indirect
	golang.org/x/sys v0.0.0-20211210111614-af8b64212486 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
)
