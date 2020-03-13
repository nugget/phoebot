package coreprotect

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nugget/phoebot/lib/phoelib"
	"github.com/sirupsen/logrus"
)

type ContainerLog struct {
	Epoch       int64
	Timestamp   time.Time
	Player      string
	WorldID     int
	World       string
	X           int
	Y           int
	Z           int
	Material    string
	Amount      int
	Action      string
	ActionCode  int
	RolledBack  int
	Preposition string
}

func (c *ContainerLog) Scan(rows *sql.Rows) error {
	return rows.Scan(
		&c.Player,
		&c.WorldID,
		&c.X,
		&c.Y,
		&c.Z,
		&c.Material,
		&c.ActionCode,
		&c.RolledBack,
		&c.Epoch,
		&c.Amount,
	)
}

func (c *ContainerLog) Parse() error {
	c.Timestamp = time.Unix(c.Epoch, 0)

	if c.ActionCode == 1 {
		c.Action = "placed"
		c.Preposition = "into"
	} else {
		c.Action = "took"
		c.Preposition = "from"
	}

	c.Material = strings.Title(strings.TrimPrefix(c.Material, "minecraft:"))
	c.World = WorldFromWid(c.WorldID)

	return nil
}

func ContainerActivity(world string, lastScan time.Time, x, y, z int) (l []ContainerLog, err error) {
	wid := WidFromWorld(world)
	epoch := lastScan.Unix()

	query := `SELECT
				u.user, wid, x, y, z, m.material, c.action, c.rolled_back, max(c.time), sum(c.amount)
		   	  FROM co_container c 
			  LEFT JOIN (co_user u, co_material_map m) on (c.type = m.rowid and c.user = u.rowid)
			  WHERE c.rolled_back = 0 
			    AND c.wid = ?
			    AND c.x = ?
				AND c.y = ?
				AND c.z >= ?
				AND c.time > ?
			  GROUP BY u.user, wid, x, y, z, m.material, c.action, c.rolled_back
			  ORDER BY max(c.time)`

	logrus.WithFields(logrus.Fields{
		"lastScan": lastScan,
		"epoch":    epoch,
		"world":    world,
		"wid":      wid,
		"x":        x,
		"y":        y,
		"z":        z,
	}).Trace("Inspecting container activity")

	rows, err := DB.Query(query, wid, x, y, z, epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		c := ContainerLog{}
		c.Scan(rows)
		if err != nil {
			return nil, err
		}
		c.Parse()

		l = append(l, c)
	}

	return l, nil
}

func ScanContainers(world string, lastScan time.Time, sx, sy, sz, fx, fy, fz int) (l []ContainerLog, err error) {
	wid := WidFromWorld(world)
	epoch := lastScan.Unix()

	sx, sy, sz, fx, fy, fz = phoelib.Rebound(sx, sy, sz, fx, fy, fz)

	query := `SELECT
				u.user, wid, x, y, z, m.material, c.action, c.rolled_back, max(c.time), sum(c.amount)
		   	  FROM co_container c 
			  LEFT JOIN (co_user u, co_material_map m) on (c.type = m.rowid and c.user = u.rowid)
			  WHERE c.rolled_back = 0 
			    AND c.wid = ?
			    AND c.x >= ? AND c.x <= ? 
				AND c.y >= ? AND c.y <= ? 
				AND c.z >= ? AND c.z <= ?
				AND c.time > ?
			  GROUP BY u.user, wid, x, y, z, m.material, c.action, c.rolled_back
			  ORDER BY max(c.time)`

	logrus.WithFields(logrus.Fields{
		"lastScan": lastScan,
		"epoch":    epoch,
		"world":    world,
		"wid":      wid,
		"start":    fmt.Sprintf("(%d, %d, %d)", sx, sy, sz),
		"finish":   fmt.Sprintf("(%d, %d, %d)", fx, fy, fz),
	}).Trace("Looking for container activity")

	rows, err := DB.Query(query, wid, sx, fx, sy, fy, sz, fz, epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		c := ContainerLog{}
		c.Scan(rows)
		if err != nil {
			return nil, err
		}
		c.Parse()

		l = append(l, c)
	}

	return l, nil
}
