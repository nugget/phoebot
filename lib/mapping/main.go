package mapping

import (
	"github.com/sirupsen/logrus"

	"github.com/nugget/phoebot/lib/db"
	"github.com/nugget/phoebot/lib/phoelib"
)

type Map struct {
	MapID  int
	Scale  int
	LeftX  int
	LeftZ  int
	RightX int
	RightZ int
}

func NewMap() Map {
	return Map{}
}

func (m *Map) Contains(x, z int) bool {
	if x >= m.LeftX && x <= m.RightX {
		if z >= m.LeftZ && z <= m.RightZ {
			return true
		}
	}

	return false
}

func MapSize(scale int) int {
	return 128 * scale
}

func NearestPos(val, scale int) int {
	mapSize := MapSize(scale)

	i := -64

	for {
		//	fmt.Printf("mapsize=%d val=%d i=%d\n", mapSize, val, i)
		if val > 0 {
			if val <= (i + mapSize) {
				return i
			}
			i += mapSize
		} else {
			if val >= (i - mapSize) {
				return i
			}
			i -= mapSize
		}

	}
}

func NearestTopLeft(x, z, scale int) (int, int) {
	cX := NearestPos(x, scale)
	cZ := NearestPos(z, scale)

	return cX, cZ
}

func MapBoundaries(x, z, scale int) (int, int, int, int) {
	tlX, tlZ := NearestTopLeft(x, z, scale)
	// fmt.Printf("Nearest Top Left to (%d, %d) is (%d, %d)\n", x, z, tlX, tlZ)

	brX := tlX + MapSize(scale) - 1
	brZ := tlZ + MapSize(scale) - 1

	return tlX, tlZ, brX, brZ
}

func Update(m Map) error {
	if m.MapID == 0 {
		logrus.WithField("map", m).Debug("Not saving mapID 0")
		return nil
	}

	logrus.WithField("map", m).Info("map.Update")

	query := `INSERT INTO map (mapid, scale, lx, lz, rx, rz)
			  SELECT $1, $2, $3, $4, $5, $6
			  ON CONFLICT (mapid) 
			     DO UPDATE SET scale = $2,
				    lx = $3, lz = $4,
					rx = $5, rz = $6`

	//phoelib.LogSQL(query, m.MapID, m.Scale, m.LeftX, m.LeftZ, m.RightX, m.RightZ)
	_, err := db.DB.Exec(query, m.MapID, m.Scale, m.LeftX, m.LeftZ, m.RightX, m.RightZ)

	return err
}

func GetByID(id int) (Map, error) {
	query := `SELECT mapid, scale, lx, lz, rx, rz FROM map WHERE mapid = $1 AND deleted IS NULL  ORDER BY changed DESC LIMIT 1`

	phoelib.LogSQL(query, id)
	rows, err := db.DB.Query(query, id)
	if err != nil {
		return Map{}, err
	}
	defer rows.Close()

	m := NewMap()
	err = rows.Scan(
		&m.MapID,
		&m.Scale,
		&m.LeftX,
		&m.LeftZ,
		&m.RightX,
		&m.RightZ,
	)
	if err != nil {
		return Map{}, err
	}

	return m, nil
}

func GetByPosition(x, z int) ([]Map, error) {
	var mapList []Map

	query := `SELECT mapid, scale, lx, lz, rx, rz FROM map
	          WHERE deleted IS NULL AND
					$1 >= lx AND $1 <= rx AND
					$2 >= lz AND $2 <= rz`

	phoelib.LogSQL(query, x, z)
	rows, err := db.DB.Query(query, x, z)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		m := NewMap()
		err := rows.Scan(
			&m.MapID,
			&m.Scale,
			&m.LeftX,
			&m.LeftZ,
			&m.RightX,
			&m.RightZ,
		)
		if err != nil {
			return nil, err
		}

		mapList = append(mapList, m)
	}

	return mapList, nil
}
