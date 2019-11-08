package mcserver

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/nugget/phoebot/lib/ipc"

	"github.com/Tnze/go-mc/authenticate"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type Server struct {
	Client      *bot.Client
	Hostname    string
	Port        int
	Email       string
	password    string
	Connected   bool
	ConnectTime time.Time
	authData    authenticate.Response
}

type PingStats struct {
	Delay         time.Duration
	PlayersOnline int64
	PlayersMax    int64
	Description   string
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
	return err
}

func (s *Server) Connect() (err error) {
	if s.Connected {
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"token":       s.authData.AccessToken,
		"connectTime": s.ConnectTime,
		"connected":   s.Connected,
	}).Debug("mcserver Connect()")

	// Super-old token
	// s.authData.AccessToken = "6432e3d646a348aca3e46aca488ba333"

	if s.authData.AccessToken == "" {
		s.authData, err = authenticate.Authenticate(s.Email, s.password)
		if err != nil {
			return err
		}
		logrus.WithFields(logrus.Fields{
			"name":  s.authData.SelectedProfile.Name,
			"uuid":  s.authData.SelectedProfile.ID,
			"token": s.authData.AccessToken,
		}).Info("Authenticated with mojang")

		s.Client.Name, s.Client.Auth.UUID, s.Client.AsTk = s.authData.SelectedProfile.Name, s.authData.SelectedProfile.ID, s.authData.AccessToken
	}

	err = s.Client.JoinServer(s.Hostname, s.Port)
	if err != nil {
		s.Connected = false

		if strings.Contains(err.Error(), "auth fail") {
			logrus.WithError(err).Error("Authentication Failure, clearing access token")
			s.authData.AccessToken = ""
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

	s.Client.Events.GameStart = s.OnGameStart
	s.Client.Events.ChatMsg = s.OnChatMsg
	s.Client.Events.Disconnect = s.OnDisconnect
	s.Client.Events.PluginMessage = s.OnPluginMessage
	s.Client.Events.Die = s.OnDieMessage

	go s.HandleGame()

	return nil
}

func (s *Server) TestConnection() error {
	if s.Connected == false {
		return fmt.Errorf("We are not connected")
	}

	inventory := s.Client.MainInventory()
	err := s.Client.SwingArm(0)

	//fmt.Printf("Client: %+v\n", Client)
	//fmt.Printf("PlayInfo: %+v\n", Client.PlayInfo)
	//fmt.Printf("Chunks: %d\n", len(Client.Wd.Chunks))
	//fmt.Printf("inv: %+v\n", inventory)

	logrus.WithFields(logrus.Fields{
		"slot0":      inventory[0],
		"chunkCount": len(s.Client.Wd.Chunks),
		"gamemode":   s.Client.PlayInfo.Gamemode,
		"dimension":  s.Client.PlayInfo.Dimension,
		"difficulty": s.Client.PlayInfo.Difficulty,
		"swingarm":   err,
	}).Trace("TestConnection")

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"slot0":      inventory[0],
			"chunkCount": len(s.Client.Wd.Chunks),
			"gamemode":   s.Client.PlayInfo.Gamemode,
			"dimension":  s.Client.PlayInfo.Dimension,
			"difficulty": s.Client.PlayInfo.Difficulty,
			"swingarm":   err,
		}).Debug("TestConnection")
		return fmt.Errorf("Can't swing my arm: %v", err)
	}

	return nil
}

func (s *Server) LogFields(f logrus.Fields) logrus.Fields {
	if f == nil {
		f = logrus.Fields{}
	}

	conn := s.Client.Conn()

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

	resp, ps.Delay, err = bot.PingAndList(s.Hostname, s.Port)
	if err != nil {
		return ps, err
	}

	json := string(resp)

	ps.PlayersOnline = gjson.Get(json, "players.online").Int()
	ps.PlayersMax = gjson.Get(json, "players.max").Int()
	ps.Description = gjson.Get(json, "description.text").String()
	ps.Version = gjson.Get(json, "version.name").String()
	ps.Protocol = gjson.Get(json, "verison.protocol").Int()

	ps.PlayersOnline--

	return ps, nil
}

func (s *Server) OnGameStart() error {
	logrus.WithFields(s.LogFields(nil)).Info("Minecraft start")
	return nil //if err isn't nil, HandleGame() will return it.
}

func (s *Server) OnChatMsg(c chat.Message, pos byte) error {
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

func (s *Server) OnDieMessage() error {
	logrus.WithFields(s.LogFields(nil)).Info("Minecraft death, respawning")
	err := s.Client.Respawn()
	if err != nil {
		logrus.WithError(err).Error("Respawn failed")
	}
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

	err := s.Client.Chat(command)
	return err
}

func (s *Server) Say(command string) error {
	logrus.WithFields(logrus.Fields{
		"command": command,
	}).Debug("Say")
	err := s.Client.Chat(command)
	return err
}

func ChatMsgClass(m chat.Message) string {
	if m.Translate == "commands.message.display.incoming" {
		return "whisper"
	}

	if m.Translate == "commands.message.display.outgoing" {
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
