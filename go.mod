module github.com/nugget/phoebot

go 1.15

// Temporarily override for now, until Tnze merges my PR
replace github.com/Tnze/go-mc => github.com/nugget/go-mc v1.14.5-0.20200831164738-48dca250ed7e

// replace github.com/Tnze/go-mc => /Users/nugget/src/go-mc

require (
	github.com/Tnze/go-mc v0.0.0-00010101000000-000000000000
	github.com/blang/semver v3.5.1+incompatible
	github.com/bwmarrin/discordgo v0.22.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/gemnasium/logrus-graylog-hook/v3 v3.0.2
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/uuid v1.1.1
	github.com/lib/pq v1.8.0
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/seeruk/minecraft-rcon v0.0.0-20190221212056-6ab996d90449
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/afero v1.3.4 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.1
	github.com/tidwall/gjson v1.6.0
	github.com/tidwall/pretty v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de // indirect
	golang.org/x/sys v0.0.0-20200810151505-1b9f1253b3ed // indirect
	gopkg.in/ini.v1 v1.57.0 // indirect
)
