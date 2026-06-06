package main

import (
	"math/rand"
	"os"
)

const outDir = "out"

var (
	white  = NewColor(255, 255, 255, 255)
	green  = NewColor(0, 255, 0, 255)
	red    = NewColor(255, 0, 0, 255)
	blue   = NewColor(64, 128, 255, 255)
	yellow = NewColor(255, 200, 0, 255)
)

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// line draws a line using the integer Bresenham algorithm (no floating point).
func line(x0, y0, x1, y1 int, framebuffer *TGAImage, color TGAColor) {
	steep := abs(x1-x0) < abs(y1-y0)
	if steep {
		x0, y0 = y0, x0
		x1, y1 = y1, x1
	}
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}
	dx := x1 - x0
	derror2 := abs(y1-y0) * 2
	yStep := 1
	if y1 < y0 {
		yStep = -1
	}
	error2 := 0
	y := y0
	for x := x0; x <= x1; x++ {
		if steep {
			framebuffer.Set(y, x, color)
		} else {
			framebuffer.Set(x, y, color)
		}
		error2 += derror2
		if error2 > dx {
			y += yStep
			error2 -= dx * 2
		}
	}
}

func main() {
	const (
		width  = 64
		height = 64
	)
	framebuffer := NewTGAImage(width, height, RGB)

	for i := 0; i < 10000; i++ {
		ax := rand.Intn(64)
		ay := rand.Intn(64)
		bx := rand.Intn(64)
		by := rand.Intn(64)
		color := NewColor(uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)))
		line(ax, ay, bx, by, framebuffer, color)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}
	if err := framebuffer.WriteTGAFile(outDir + "/framebuffer.tga"); err != nil {
		panic(err)
	}
}
