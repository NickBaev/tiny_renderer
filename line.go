package main

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
