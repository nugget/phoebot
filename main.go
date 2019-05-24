package main

import (
	"fmt"

	"github.com/nugget/phoebot/serverpro"
)

func main() {
	fmt.Println("vim-go")

	lv, err := serverpro.LatestVersion("Paper")

	fmt.Println(lv, err)
}
