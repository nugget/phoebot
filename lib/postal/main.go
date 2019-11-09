package postal

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nugget/phoebot/lib/console"
	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/discord"
	"github.com/nugget/phoebot/lib/ipc"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/nugget/phoebot/lib/player"
	"github.com/nugget/phoebot/models"

	"github.com/sirupsen/logrus"
)

type Item struct {
	Slot  int    `nbt:"Slot"`
	ID    string `nbt:"id"`
	Count int    `nbt:"Count"`
}

type Mailbox struct {
	ContainerID string
	ID          string `nbt:"id"`
	X           int    `nbt:"x"`
	Y           int    `nbt:"y"`
	Z           int    `nbt:"z"`
	Dimension   string
	CustomName  string `nbt:"CustomName"`
	PlayerName  string
	Items       []Item `nbt:"Items"`
	PlayerID    string
	NBT         string
}

func NewMailbox(data string) (Mailbox, error) {
	var m Mailbox

	m.NBT = data
	err := m.Parse()

	return m, err
}

func (m *Mailbox) Log(desc string) error {
	logrus.WithFields(logrus.Fields{
		"nbt": m.NBT,
	}).Debug("Mailbox NBT")

	logrus.WithFields(logrus.Fields{
		"id":          m.ID,
		"x":           m.X,
		"y":           m.Y,
		"z":           m.Z,
		"dimension":   m.Dimension,
		"customName":  m.CustomName,
		"playerName":  m.PlayerName,
		"playerID":    m.PlayerID,
		"isContainer": m.IsContainer(),
		"containerID": m.ContainerID,
	}).Info(desc)

	return nil
}

func (m *Mailbox) IsContainer() bool {
	re := regexp.MustCompile(`Items:`)
	res := re.FindString(m.NBT)
	return (res != "")
}

func (m *Mailbox) ContainerIDFromLocation() error {
	query := `SELECT containerID FROM container WHERE x = $1 AND y = $2 AND z = $3 AND dimension = $4 AND deleted IS NULL ORDER BY changed DESC LIMIT 1`

	phoelib.LogSQL(query, m.X, m.Y, m.Z, m.Dimension)
	rows, err := db.DB.Query(query, m.X, m.Y, m.Z, m.Dimension)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&m.ContainerID)
		if err != nil {
			return err
		}

		logrus.Trace("Successfully linked containerID from coordinates")
	}

	return nil
}

func (m *Mailbox) FindPlayerName() (err error) {
	if m.PlayerName != "" {
		logrus.Trace("FindPlayerName no work to do")
		return nil
	}

	if m.PlayerID != "" {
		m.PlayerName, _ = player.GameNickFromPlayerID(m.PlayerID)
		logrus.WithFields(logrus.Fields{
			"PlayerID":   m.PlayerID,
			"PlayerName": m.PlayerName,
		}).Trace("FindPlayerName ID was not empty")
	}

	if m.PlayerName == "" {
		re := regexp.MustCompile("^([^']+)'s Mailbox")
		res := re.FindStringSubmatch(m.CustomName)
		if len(res) == 2 {
			logrus.Trace("FindPlayerName customname regexp hit")
			m.PlayerName = res[1]
		} else {
			logrus.Trace("FindPlayerName customname regexp miss")
			m.PlayerName = m.CustomName
			m.PlayerName = strings.ReplaceAll(m.PlayerName, "MAIL BOX", "")
			m.PlayerName = strings.ReplaceAll(m.PlayerName, "MAILBOX", "")
			m.PlayerName = strings.ReplaceAll(m.PlayerName, "'s", "")
			m.PlayerName = strings.TrimSpace(m.PlayerName)
		}
	}

	return nil
}

func (m *Mailbox) Parse() (err error) {
	m.X = ParseInt(m.NBT, "x")
	m.Y = ParseInt(m.NBT, "y")
	m.Z = ParseInt(m.NBT, "z")
	m.Dimension = "overworld"

	m.ID, err = console.GetString(m.X, m.Y, m.Z, "id")
	if err != nil {
		return err
	}

	// This is fine if it errors
	m.CustomName, _ = console.GetText(m.X, m.Y, m.Z, "CustomName")
	m.FindPlayerName()

	if m.PlayerID == "" {
		m.PlayerID, _ = player.PlayerIDFromGameNick(m.PlayerName)
	}

	err = m.ContainerIDFromLocation()

	return err
}

func (m *Mailbox) Update() (bool, error) {
	var err error

	if m.CustomName == "" {
		return false, nil
	}
	if m.ContainerID == "" {
		// This is a newly-discovered mailbox
		query := `INSERT INTO container (name, id, dimension, x, y, z, nbt, playerID)
              SELECT $1, $2, $3, $4, $5, $6, $7, $8`

		phoelib.LogSQL(query, m.CustomName, m.ID, m.Dimension, m.X, m.Y, m.Z, m.NBT, m.PlayerID)
		_, err := db.DB.Exec(query, m.CustomName, m.ID, m.Dimension, m.X, m.Y, m.Z, m.NBT, m.PlayerID)
		if err != nil {
			return false, err
		}

		m.Log("Stored new container in db")
		return false, nil
	}

	// This is an existing container we are updating
	m.NBT, err = console.GetBlock(m.X, m.Y, m.Z, "")
	if err != nil {
		return false, err
	}

	err = m.Parse()
	if err != nil {
		m.Log("Parse Failed")
		return false, err
	}

	query := `UPDATE container SET nbt = $2, name = $3, playerid = $4, id = $5
	          WHERE containerID = $1 AND
			  (nbt <> $2 OR name <> $3 OR playerid <> $4 OR id <> $5)
			  RETURNING containerID`

	phoelib.LogSQL(query, m.ContainerID, m.NBT, m.CustomName, m.PlayerID, m.ID)
	rows, err := db.DB.Query(query, m.ContainerID, m.NBT, m.CustomName, m.PlayerID, m.ID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cID string
		rows.Scan(&cID)
		m.Log("Updated changed container in db")
		return true, nil
	}

	return false, nil
}

func (m *Mailbox) Notify() error {
	message := fmt.Sprintf("You've got mail in %s at (%d, %d, %d)", m.CustomName, m.X, m.Y, m.Z)

	sentWhisper, sentDiscord := false, false

	if m.PlayerName != "" {
		w := models.Whisper{m.PlayerName, message}

		if ipc.ServerWhisperStream != nil {
			ipc.ServerWhisperStream <- w
			sentWhisper = true
		}
	}

	if m.PlayerID != "" {
		channel, err := discord.GetChannelByPlayerID(m.PlayerID)
		if err != nil {
			logrus.WithError(err).Trace("Notify Discord channel lookup failed")
		} else {
			discord.Session.ChannelMessageSend(channel.ID, message)
			sentDiscord = true
		}
	}

	logrus.WithFields(logrus.Fields{
		"Player":  m.PlayerName,
		"whisper": sentWhisper,
		"discord": sentDiscord,
	}).Debug("Sent notifications for updated mailbox")

	return nil
}

func PollContainers() error {
	query := `SELECT containerID, id, x, y, z, name, NBT, playerID
			  FROM container WHERE deleted IS NULL AND enabled IS TRUE`

	phoelib.LogSQL(query)
	rows, err := db.DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	polled, updated, failed := 0, 0, 0

	for rows.Next() {
		polled++

		m := Mailbox{}
		err := rows.Scan(
			&m.ContainerID,
			&m.ID,
			&m.X,
			&m.Y,
			&m.Z,
			&m.CustomName,
			&m.NBT,
			&m.PlayerID,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed Mailbox scan")
			failed++
			continue
		}

		m.FindPlayerName()

		changed, err := m.Update()
		if err != nil {
			logrus.WithError(err).Error("Failed Mailbox Update")
			failed++
			continue
		}
		if changed {
			updated++
			err := m.Notify()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error":       err,
					"name":        m.CustomName,
					"playerID":    m.PlayerID,
					"containerID": m.ContainerID,
				}).Error("Error sending notification")
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"polled":  polled,
		"updated": updated,
		"failed":  failed,
	}).Trace("PollContainers")

	return nil
}

func SearchForMailboxes(sx, sy, sz, fx, fy, fz int) error {
	if sx > fx {
		sx, fx = fx, sx
	}

	if sy > fy {
		sy, fy = fy, sy
	}

	if sz > fz {
		sz, fz = fz, sz
	}

	logrus.WithFields(logrus.Fields{
		"sx": sx,
		"sy": sy,
		"sz": sz,
		"fx": fx,
		"fy": fy,
		"fz": fz,
	}).Debug("SearchForMailboxes")

	failed, unassigned, updated := 0, 0, 0

	for x := sx; x <= fx; x++ {
		for y := sy; y <= fy; y++ {
			for z := sz; z <= fz; z++ {
				data, err := console.GetBlock(x, y, z, "")
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"x":   x,
						"y":   y,
						"z":   z,
						"err": err,
					}).Debug("GetBlock Failure")
					failed++
					continue
				}

				m, err := NewMailbox(data)
				if err != nil {
					logrus.WithError(err).Error("NewMailbox Failure")
					failed++
					continue
				}

				if m.CustomName == "MAIL BOX" {
					unassigned++
					continue
				}

				if m.IsContainer() {
					if m.ContainerID == "" {
						// This is a new container
						_, err := m.Update()
						if err != nil {
							logrus.WithError(err).Error("Mailbox Update() Failure")
							failed++
							continue
						} else {
							updated++
						}
					}
				}
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"sx":         sx,
		"sy":         sy,
		"sz":         sz,
		"fx":         fx,
		"fy":         fy,
		"fz":         fz,
		"failed":     failed,
		"unassigned": unassigned,
		"updated":    updated,
	}).Info("SearchForMailboxes Complete")

	return nil
}

func SearchServer(hostname string) (err error) {
	switch hostname {
	case "phoenixcraft.serv.nu":
		err = SearchForMailboxes(-35, 71, 152, -29, 69, 152)
		if err != nil {
			return err
		}
		err = SearchForMailboxes(-11, 71, 152, -5, 69, 152)
		if err != nil {
			return err
		}
	case "172.28.0.24":
		err = SearchForMailboxes(200, 81, 264, 204, 79, 264)
		if err != nil {
			return err
		}
	}

	return nil
}
