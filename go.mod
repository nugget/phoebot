module github.com/nugget/phoebot

go 1.16

// Temporarily override for now, until Tnze merges my PR
//replace github.com/Tnze/go-mc => /Users/nugget/src/go-mc
//replace github.com/Tnze/go-mc => github.com/nugget/go-mc v1.14.5-0.20201118172317-e9cb621f23ef

require (
	github.com/Tnze/go-mc v1.17.1-0.20210806203433-99081e1b9cfb
	github.com/beefsack/go-astar v0.0.0-20200827232313-4ecf9e304482 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/bwmarrin/discordgo v0.23.2
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/gemnasium/logrus-graylog-hook/v3 v3.0.3
	github.com/go-sql-driver/mysql v1.6.0
	github.com/google/uuid v1.2.0
	github.com/lib/pq v1.10.2
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mitchellh/mapstructure v1.4.0 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/seeruk/minecraft-rcon v0.0.0-20190221212056-6ab996d90449
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.1
	github.com/tidwall/gjson v1.8.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
	golang.org/x/sys v0.0.0-20210104204734-6f8348627aad // indirect
	golang.org/x/text v0.3.4 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
