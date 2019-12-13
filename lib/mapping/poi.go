package mapping

import (
	"fmt"

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
	Dimension   string
	Distance    int
}

func NewPOI() POI {
	return POI{}
}

func (p POI) String() string {
	switch p.Class {
	case "portal":
		return p.Description
	case "base":
		return fmt.Sprintf("%s's Base", p.Description)
	default:
		return p.Description
	}
}

func (p POI) Fields() string {
	return "x, y, z, class, description, owner, dimension"
}

func (p POI) Update() error {
	logrus.WithField("poi", p).Info("map.NewPOI")

	query := `INSERT INTO poi (` + p.Fields() + `)
			  SELECT $1, $2, $3, $4, $5, $6, $7`

	_, err := db.DB.Exec(query, p.X, p.Y, p.Z, p.Class, p.Description, p.Owner, p.Dimension)

	return err
}

func NearestPOI(class string, x, z int, dimension string) (POI, error) {
	p := NewPOI()

	logrus.WithFields(logrus.Fields{
		"class":     class,
		"x":         x,
		"z":         z,
		"dimension": dimension,
	}).Trace("NearestPOI Called")

	query := `SELECT ` + p.Fields() + `,
				  (point(x,z) <-> point($1, $2))::int AS distance
	          FROM poi WHERE dimension = $3 AND class = $4 AND deleted IS NULL
			  ORDER BY distance LIMIT 1`

	rows, err := db.DB.Query(query, x, z, dimension, class)
	if err != nil {
		return p, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&p.X,
			&p.Y,
			&p.Z,
			&p.Class,
			&p.Description,
			&p.Owner,
			&p.Dimension,
			&p.Distance,
		)
		logrus.WithFields(logrus.Fields{
			"class":       p.Class,
			"x":           p.X,
			"y":           p.Y,
			"z":           p.Z,
			"description": p.Description,
			"err":         err,
		}).Trace("NearestPOI Loop")
		return p, err
	}

	return p, fmt.Errorf("Unknown error in NearestPOI")
}
