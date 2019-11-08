package postal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func ParseText(nbt, fieldName string) string {
	t := ""

	pattern := fmt.Sprintf(`%s: '{"text":"([^"]+)"`, fieldName)
	re := regexp.MustCompile(pattern)
	res := re.FindStringSubmatch(nbt)
	if len(res) == 2 {
		t = res[1]
	}

	t = strings.ReplaceAll(t, `\'`, `'`)

	return t
}

func ParseString(nbt, fieldName string) string {
	t := ""

	pattern := fmt.Sprintf(`%s: "([^"]+)"`, fieldName)
	re := regexp.MustCompile(pattern)
	res := re.FindStringSubmatch(nbt)
	if len(res) == 2 {
		t = res[1]
	}

	t = strings.ReplaceAll(t, `\'`, `'`)

	return t
}

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
