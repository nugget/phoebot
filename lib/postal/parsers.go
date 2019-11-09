package postal

import (
	"fmt"
	"regexp"
	"strconv"
)

func ParseInt(nbt, fieldName string) int {
	i := 0

	pattern := fmt.Sprintf(`%s: ([0-9-]+)`, fieldName)
	re := regexp.MustCompile(pattern)
	res := re.FindStringSubmatch(nbt)
	if len(res) == 2 {
		i, _ = strconv.Atoi(res[1])
	}

	return i
}
