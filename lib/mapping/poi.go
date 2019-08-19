package mapping

import (
	"github.com/sirupsen/logrus"

	"github.com/nugget/phoebot/lib/db"
)

type POI struct {
	X           int
	Y           int
	Z           int
	Class       string
	Description string
	Owner       string
}

func NewPOI() POI {
	return POI{}
}

func UpdatePOI(p POI) error {
	logrus.WithField("poi", p).Info("map.NewPOI")

	query := `INSERT INTO poi (x, y, z, class, description, owner)
			  SELECT $1, $2, $3, $4, $5, $6`

	_, err := db.DB.Exec(query, p.X, p.Y, p.Z, p.Class, p.Description, p.Owner)

	return err
}
