package main

import (
	"flag"
	"fmt"
)

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

func main() {
	var (
		x     int
		z     int
		scale int
	)

	flag.IntVar(&x, "x", 0, "X Position")
	flag.IntVar(&z, "z", 0, "X Position")
	flag.IntVar(&scale, "scale", 1, "Map Scale")
	flag.Parse()

	tlX, tlZ, brX, brZ := MapBoundaries(x, z, scale)
	fmt.Printf("Map spans from (%d, %d) to (%d, %d)\n", tlX, tlZ, brX, brZ)

}
