package mcserver

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/nugget/phoebot/lib/config"
	"github.com/nugget/phoebot/lib/ipc"

	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"github.com/Tnze/go-mc/bot/screen"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/yggdrasil"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	_ "github.com/Tnze/go-mc/data/lang/en-au"

	pk "github.com/Tnze/go-mc/net/packet"
)

type Server struct {
	Client        *bot.Client
	Player        *basic.Player
	ScreenManager *screen.Manager
	Hostname      string
	Port          int
	Email         string
	password      string
	Connected     bool
	ConnectTime   time.Time
	MyName        string
	auth          *yggdrasil.Access
}

type PingStats struct {
	Delay         time.Duration
	PlayersOnline int64
	PlayersMax    int64
	MOTD          string
	Version       string
	Protocol      int64
}

func New() (Server, error) {
	s := Server{}
	s.Client = bot.NewClient()

	return s, nil
}

func (s *Server) Authenticate(hostname string, port int, email, password string) error {
	s.Hostname = hostname
	s.Port = port
	s.Email = email
	s.password = password

	return nil
}

func (s *Server) HandleGame() error {
	logrus.Debug("Minecraft HandleGame Started")
	err := s.Client.HandleGame()
	logrus.WithError(err).Error("Minecraft HandleGame Exited")

	s.Connected = false

	logrus.Error("Restarting")
	os.Exit(0)

	return err
}

func (s *Server) Connect() (err error) {
	if s.Connected {
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"connectTime": s.ConnectTime,
		"connected":   s.Connected,
	}).Debug("mcserver Connect()")

	if s.auth != nil {
		ok, err := s.auth.Validate()
		if err != nil {
			return err
		}

		if !ok {
			s.auth.Invalidate()
			logrus.Error("Mojang accessToken is not valid")
		}
	}

	s.auth, err = yggdrasil.Authenticate(s.Email, s.password)
	if err != nil {
		return err
	}

	s.Client.Auth.UUID, s.Client.Auth.Name = s.auth.SelectedProfile()
	s.Client.Auth.AsTk = s.auth.AccessToken()

	logrus.WithFields(logrus.Fields{
		"name":  s.Client.Name,
		"uuid":  s.Client.Auth.UUID,
		"token": s.Client.Auth.AsTk,
	}).Info("Authenticated with mojang")

	err = s.Client.JoinServer(s.Hostname)
	if err != nil {
		s.Connected = false

		if strings.Contains(err.Error(), "auth fail") {
			logrus.WithError(err).Error("Authentication Failure, clearing access token")
			s.auth.Invalidate()
		}

		return err
	}

	logrus.WithFields(logrus.Fields{
		"hostname": s.Hostname,
		"port":     s.Port,
		"uuid":     s.Client.Auth.UUID,
		"name":     s.Client.Auth.Name,
	}).Info("Connected to Minecraft server")

	s.Connected = true
	s.ConnectTime = time.Now()

	s.Player = basic.NewPlayer(s.Client, basic.DefaultSettings)

	basic.EventsListener{
		GameStart:    s.OnGameStart,
		ChatMsg:      s.OnChatMsg,
		Disconnect:   s.OnDisconnect,
		HealthChange: s.OnHealthChange,
		Death:        s.OnDeath,
	}.Attach(s.Client)

	go s.HandleGame()

	return nil
}

func (s *Server) TestConnection() error {
	if s.Connected == false {
		return fmt.Errorf("We are not connected")
	}

	if s.Client.Auth.Name != s.MyName {
		s.MyName = s.Client.Auth.Name
		config.WriteString("minecraftName", s.Client.Auth.Name)
	}

	return nil
}

func (s *Server) WaitForServer(secs int) {
	for {
		err := s.TestConnection()
		if err != nil {
			logrus.Trace("mc.WaitForServer Loop")
		} else {
			break
		}
		time.Sleep(time.Duration(secs) * time.Second)
	}
}

func (s *Server) LogFields(f logrus.Fields) logrus.Fields {
	if f == nil {
		f = logrus.Fields{}
	}

	conn := s.Client.Conn

	f["name"] = s.Client.Auth.Name
	f["server"] = conn.Socket.RemoteAddr()

	return f
}

func (s *Server) Handler() {
	backoffSeconds := 10
	backoffLimit := 300
	retries := 1

	for {
		logrus.Trace("Handler loop")

		err := s.Connect()
		if err == nil {
			// Reset the backoff after a successful connection
			backoffSeconds = 5
			retries = 1
		} else {
			retries++
			if backoffSeconds < backoffLimit {
				backoffSeconds = int(float64(backoffSeconds) * 1.5)
			}

			logrus.WithFields(logrus.Fields{
				"error":          err,
				"retries":        retries,
				"backoffSeconds": backoffSeconds,
				"backoffLimit":   backoffLimit,
			}).Error("Unable to connect to Minecraft server")

			if retries > 50 {
				logrus.Fatal("Retry count exceeded, restarting service")
			}

			if strings.Contains(err.Error(), "encryption fail") ||
				strings.Contains(err.Error(), "read content of packet fail") {
				logrus.Fatal("This is a fatal error, exiting.")
			}

		}

		time.Sleep(time.Duration(backoffSeconds) * time.Second)

		err = s.TestConnection()
		if err != nil {
			logrus.WithError(err).Warn("TestConnection")
			s.Connected = false
		}
	}
}

func (s *Server) Status() (ps PingStats, err error) {
	var resp []byte

	hostPort := fmt.Sprintf("%s:%d", s.Hostname, s.Port)
	resp, ps.Delay, err = bot.PingAndList(hostPort)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"addr":  hostPort,
			"error": err,
		}).Debug("bot.PingAndList Error")
		return ps, err
	}

	json := string(resp)

	ps.PlayersOnline = gjson.Get(json, "players.online").Int()
	ps.PlayersMax = gjson.Get(json, "players.max").Int()
	ps.MOTD = gjson.Get(json, "description.text").String()
	ps.Version = gjson.Get(json, "version.name").String()
	ps.Protocol = gjson.Get(json, "verison.protocol").Int()

	ps.PlayersOnline--

	config.WriteInt("players", ps.PlayersOnline)

	return ps, nil
}

func (s *Server) OnGameStart() error {
	logrus.WithFields(s.LogFields(nil)).Info("Minecraft start")
	return nil //if err isn't nil, HandleGame() will return it.
}

func (s *Server) OnHealthChange(health float32) error {
	logrus.WithFields(s.LogFields(logrus.Fields{
		"health": health,
	})).Info("Health change")
	return nil
}

func (s *Server) OnChatMsg(c chat.Message, pos byte, uuid uuid.UUID) error {
	f := s.LogFields(logrus.Fields{
		"pos":   pos,
		"event": "chat",
	})

	if ipc.ServerChatStream != nil {
		ipc.ServerChatStream <- c
	} else {
		logrus.WithFields(f).Info(c.ClearString())
	}

	return nil
}

func (s *Server) OnDisconnect(c chat.Message) error {
	logrus.WithFields(s.LogFields(logrus.Fields{
		"message": c.ClearString(),
	})).Info("Minecraft disconnect")

	s.Connected = false

	return nil
}

func (s *Server) OnPluginMessage(channel string, data []byte) error {
	logrus.WithFields(s.LogFields(logrus.Fields{
		"channel": channel,
		"data":    data,
		"string":  string(data),
	})).Info("Minecraft Plugin Message")
	return nil
}

func (s *Server) OnDeath() error {
	logrus.WithFields(s.LogFields(nil)).Info("Minecraft death, respawning")
	go func() {
		time.Sleep(time.Second * 5)
		err := s.Player.Respawn()
		if err != nil {
			logrus.WithError(err).Error("Respawn failed")
		}
	}()
	return nil
}

func (s *Server) Whisper(who, message string) error {
	if who == "" {
		return errors.New("Cannot send a message to nobody")
	}
	command := fmt.Sprintf("/tell %s %s", who, message)
	logrus.WithFields(logrus.Fields{
		"who":     who,
		"command": command,
	}).Debug("Whisper")

	if !s.Connected {
		return fmt.Errorf("mcserver client connection is nil")
	}

	//err := s.Client.Chat(command)
	err := s.Client.Conn.WritePacket(pk.Marshal(
		packetid.ServerboundChat,
		pk.String(command),
	))

	return err
}

func (s *Server) Say(command string) error {
	logrus.WithFields(logrus.Fields{
		"command": command,
	}).Debug("Say")

	// err := s.Client.Chat(command)
	err := s.Client.Conn.WritePacket(pk.Marshal(
		packetid.ServerboundChat,
		pk.String(command),
	))

	return err
}

func ChatMsgClass(m chat.Message) string {
	if m.Translate == "commands.message.display.incoming" {
		return "whisper"
	}

	if m.Translate == "sleep.skipping_night" {
		return "ignore"
	}

	if m.Translate == "commands.message.display.outgoing" {
		return "ignore"
	}

	if m.Translate == "commands.give.success.single" {
		return "ignore"
	}

	if m.Translate == "commands.data.block.modified" {
		return "ignore"
	}

	if strings.HasPrefix(m.Translate, "chat.type.text") {
		return "chat"
	}

	if strings.HasPrefix(m.Translate, "gamemode") {
		return "ignore"
	}

	if strings.HasPrefix(m.Translate, "chat.type.emote") {
		return "chat"
	}

	if strings.HasPrefix(m.Translate, "chat.type.announcement") {
		return "announcement"
	}

	if strings.HasPrefix(m.Translate, "death.") {
		return "death"
	}

	text := m.ClearString()

	if strings.HasPrefix(text, "<") {
		return "chat"
	}

	if strings.HasPrefix(text, "* ") {
		return "chat"
	}

	if strings.Contains(text, "joined the game") {
		return "join"
	}

	if strings.Contains(text, "left the game") {
		return "join"
	}

	if strings.Contains(text, "went to bed") {
		return "ignore"
	}

	if strings.Contains(text, "is now AFK") {
		return "ignore"
	}

	if strings.Contains(text, "is no longer AFK") {
		return "ignore"
	}

	if strings.Contains(text, "No player was found") {
		return "ignore"
	}

	if strings.Contains(text, "Expected whitespace") {
		return "ignore"
	}

	if strings.Contains(text, "Couldn't grant advancement") {
		return "ignore"
	}

	if strings.Contains(text, "Granted the advancement") {
		return "ignore"
	}

	if strings.HasPrefix(text, "[WEB]") {
		return "chat"
	}

	if strings.HasPrefix(text, "[") {
		return "ignore"
	}

	return "other"
}

func GetPlayerNameFromWhisper(data string) (string, error) {
	r := regexp.MustCompile(`^([^ ]+) whispers to you:`)
	res := r.FindStringSubmatch(data)

	if len(res) != 2 {
		return "", fmt.Errorf("Unable to parse Whisper data")
	}

	return res[1], nil
}
