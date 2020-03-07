package coreprotect

import (
	"fmt"
	"strings"
)

var BlockdataLookup map[int]string

func Blockdata(code int) (string, string, error) {
	if BlockdataLookup == nil {
		BlockdataLookup = make(map[int]string)
	}

	var eq string

	_, ok := BlockdataLookup[code]
	if !ok {
		query := `SELECT data FROM co_blockdata_map WHERE id = ?`

		rows, err := DB.Query(query, code)
		if err != nil {
			return "", "", err
		}
		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&eq)
			if err != nil {
				return "", "", err
			}

			BlockdataLookup[code] = eq
		}

	}

	sp := strings.Split(eq, "=")

	k := sp[0]
	v := sp[1]

	if len(sp) == 2 {
		return k, v, nil
	}

	return "", "", fmt.Errorf("Unable to parse eq=%s", eq)
}
