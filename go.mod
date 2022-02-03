module github.com/nugget/phoebot

go 1.17

// Temporarily override for now, until Tnze merges my PR
//replace github.com/Tnze/go-mc => /Users/nugget/src/go-mc
//replace github.com/Tnze/go-mc => github.com/nugget/go-mc v1.14.5-0.20201118172317-e9cb621f23ef

require (
	github.com/Tnze/go-mc v1.17.2-0.20211211093242-38cc60ea6593
	github.com/blang/semver v3.5.1+incompatible
	github.com/bwmarrin/discordgo v0.23.2
	github.com/gemnasium/logrus-graylog-hook/v3 v3.1.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/google/uuid v1.3.0
	github.com/lib/pq v1.10.4
	github.com/seeruk/minecraft-rcon v0.0.0-20190221212056-6ab996d90449
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.10.1
	github.com/tidwall/gjson v1.14.0
)

require (
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/afero v1.7.1 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3 // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
