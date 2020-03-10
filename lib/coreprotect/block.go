package coreprotect

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Actions = map[int]string{
	0: "destroyed",
	1: "placed",
	2: "toggled",
}

type Block struct {
	RowID        int64
	Epoch        int64
	Timestamp    time.Time
	User         string
	UserID       int
	World        string
	WorldID      int
	X            int
	Y            int
	Z            int
	TypeID       int
	Material     string
	Data         int
	Meta         []byte
	BlockdataRaw []byte
	Blockdata    map[string]string
	Action       string
	ActionCode   int
	RolledBack   bool
}

func (b *Block) Parse() error {
	if b.Blockdata == nil {
		b.Blockdata = make(map[string]string)
	}
	b.Action = Actions[b.ActionCode]

	if len(b.BlockdataRaw) > 0 {
		for _, bd := range strings.Split(string(b.BlockdataRaw), ",") {
			bdi, err := strconv.ParseInt(bd, 10, 0)
			if err != nil {
				return err
			}
			k, v, err := Blockdata(int(bdi))
			if err != nil {
				return err
			}
			b.Blockdata[k] = v

		}
	}

	return nil
}

func (b *Block) MountedOn() (Block, error) {
	var (
		x int
		y int
		z int
	)

	if !strings.Contains(b.Material, "wall_sign") {
		// This is not a wall sign
		return Block{}, fmt.Errorf("%s is not a wall sign", b.Material)
	}

	x = b.X
	y = b.Y
	z = b.Z

	// test is at -186 72 -181
	facing := b.Blockdata["facing"]
	switch facing {
	case "west":
		x++
	case "east":
		x--
	case "south":
		z--
	case "north":
		z++
	default:
		z++
	}

	retBlock, err := GetBlock(b.WorldID, x, y, z)
	if err != nil {
		return Block{}, err
	}

	//fmt.Printf("Sign at %d %d %d facing %s is attached to %s at %d %d %d\n",
	//	b.X, b.Y, b.Z, facing,
	//	retBlock.Material, retBlock.X, retBlock.Y, retBlock.Z,
	//)

	return retBlock, nil
}

func GetBlock(wid, x, y, z int) (b Block, err error) {
	query := `SELECT b.rowid, b.time, u.user, b.user as userid, b.wid, w.world, b.x, b.y, b.z, b.type, m.material, b.data, b.meta, b.blockdata, b.action, b.rolled_back
			  FROM co_block b
			  LEFT JOIN (co_user u, co_material_map m, co_world w) on (b.type = m.rowid and b.user = u.rowid and w.rowid = b.wid)
			  WHERE b.wid = ? AND b.x = ? AND b.y = ? AND b.z = ? AND 
			  ORDER BY b.time DESC LIMIT 1`

	rows, err := DB.Query(query, wid, x, y, z)
	if err != nil {
		return Block{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&b.RowID,
			&b.Epoch,
			&b.User,
			&b.UserID,
			&b.WorldID,
			&b.World,
			&b.X,
			&b.Y,
			&b.Z,
			&b.TypeID,
			&b.Material,
			&b.Data,
			&b.Meta,
			&b.BlockdataRaw,
			&b.ActionCode,
			&b.RolledBack,
		)
		if err != nil {
			return Block{}, err
		}
		err = b.Parse()
		if err != nil {
			return Block{}, err
		}
	}

	return b, nil
}
