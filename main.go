package main

import (
	"math/rand/v2"
	"runtime"
	"sync"
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

func sortvec3(a, b, c, x1, x2, x3 int) (int, int, int, int, int, int) {
	if a > b {
		a, b = b, a
		x1, x2 = x2, x1
	}
	if a > c {
		a, c = c, a
		x1, x3 = x3, x1
	}
	if b > c {
		b, c = c, b
		x2, x3 = x3, x2
	}
	return a, b, c, x1, x2, x3
}

func triangle(x1, y1, x2, y2, x3, y3 int, framebuffer *TGAImage, color TGAColor) {
	yMin, yMid, yMax, xMin, xMid, xMax := sortvec3(y1, y2, y3, x1, x2, x3)
	if yMin == yMax || yMid == yMax || yMin == yMid || xMin == xMax || xMid == xMax || xMin == xMid {
		return // вырожденный треугольник
	}

	lineMax := lineX(xMin, yMin, xMax, yMax)
	lineDown := lineX(xMin, yMin, xMid, yMid)
	lineUp := lineX(xMid, yMid, xMax, yMax)
	for i := 0; i < len(lineMax); i++ {
		xStart := lineMax[i]
		y := yMin + i
		var xEnd int
		if y < yMid {
			xEnd = lineDown[i]
		} else {
			xEnd = lineUp[y-yMid]
		}
		lo, hi := xStart, xEnd
		if lo > hi {
			lo, hi = hi, lo
		}
		for x := lo; x <= hi; x++ {
			framebuffer.Set(x, y, color)
		}
	}

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

// lineX returns one x coordinate per scanline row, indexed by y-y0: the x where
// the edge enters that row. Bresenham is driven along the y axis (no floats, no
// per-row division), so x may advance several steps per row.
func lineX(x0, y0, x1, y1 int) []int {
	if y0 > y1 { // always iterate from low y to high y
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}
	dy := y1 - y0
	dx := x1 - x0
	xStep := 1
	if dx < 0 {
		xStep = -1
	}
	derror2 := abs(dx) * 2
	error2 := 0
	x := x0
	coords := make([]int, 0, dy+1)
	for y := y0; y <= y1; y++ {
		coords = append(coords, x)
		error2 += derror2
		for error2 > dy {
			x += xStep
			error2 -= dy * 2
		}
	}

	return coords
}

func modernTriangle(x1, y1, z1, x2, y2, z2, x3, y3, z3 int, framebuffer *TGAImage, zbuffer *TGAImage, color TGAColor) {
	bboxMinX := min(x1, x2, x3)
	bboxMinY := min(y1, y2, y3)
	bboxMaxX := max(x1, x2, x3)
	bboxMaxY := max(y1, y2, y3)

	// Вырожденный (нулевая площадь) треугольник: иначе проверка "одного знака"
	// залила бы весь bbox.
	area := edgeFunction(x3, y3, x1, y1, x2, y2)
	if area == 0 {
		return
	}

	// Приращения edge-функций на один шаг по x и по y. edgeFunction линейна, поэтому
	// в горячем цикле остаются только сложения — без умножений. Считаются один раз,
	// горутины их только читают.
	stepX0, stepY0 := y1-y2, x2-x1
	stepX1, stepY1 := y2-y3, x3-x2
	stepX2, stepY2 := y3-y1, x1-x3

	// Каждая горутина обрабатывает свой непрерывный диапазон строк (y).
	// Строки не пересекаются, значит и байты в framebuffer.data не пересекаются,
	// поэтому Set можно вызывать параллельно без блокировок.
	rows := bboxMaxY - bboxMinY + 1
	workers := runtime.NumCPU()
	if workers > rows {
		workers = rows
	}

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		yStart := bboxMinY + w*rows/workers
		yEnd := bboxMinY + (w+1)*rows/workers // не включительно
		wg.Add(1)
		go func(yStart, yEnd int) {
			defer wg.Done()
			// Каждая горутина стартует со своего угла (bboxMinX, yStart).
			edge0Row := edgeFunction(bboxMinX, yStart, x1, y1, x2, y2)
			edge1Row := edgeFunction(bboxMinX, yStart, x2, y2, x3, y3)
			edge2Row := edgeFunction(bboxMinX, yStart, x3, y3, x1, y1)
			for y := yStart; y < yEnd; y++ {
				edge0, edge1, edge2 := edge0Row, edge1Row, edge2Row
				for x := bboxMinX; x <= bboxMaxX; x++ {
					if edge0 >= 0 && edge1 >= 0 && edge2 >= 0 || edge0 <= 0 && edge1 <= 0 && edge2 <= 0 {
						z := uint8((edge0*z3 + edge1*z1 + edge2*z2) / area)
						if z <= zbuffer.Get(x, y).B {
							edge0 += stepX0
							edge1 += stepX1
							edge2 += stepX2
							continue
						}
						// color = TGAColor{z,z,z,255}
						zbuffer.Set(x, y, TGAColor{z, z, z, z})
						framebuffer.Set(x, y, color)
					}
					edge0 += stepX0
					edge1 += stepX1
					edge2 += stepX2
				}
				edge0Row += stepY0
				edge1Row += stepY1
				edge2Row += stepY2
			}
		}(yStart, yEnd)
	}
	wg.Wait()
}

// edgeFunction возвращает удвоенную знаковую площадь треугольника (A, B, P) —
// z-компоненту векторного произведения (B-A) × (P-A). Знак показывает, с какой
// стороны прямой AB лежит точка P=(x, y). Раскрытая форма даёт постоянные
// приращения по x (ay-by) и по y (bx-ax) для инкрементального растеризатора.
func edgeFunction(x, y, ax, ay, bx, by int) int {
	return y*(bx-ax) + x*(ay-by) + ax*by - ay*bx
}

func modernTriangleGradient(x1, y1, x2, y2, x3, y3 int, framebuffer *TGAImage) {
	bboxMinX := min(x1, x2, x3)
	bboxMinY := min(y1, y2, y3)
	bboxMaxX := max(x1, x2, x3)
	bboxMaxY := max(y1, y2, y3)

	// Вырожденный (нулевая площадь) треугольник: иначе проверка "одного знака"
	// залила бы весь bbox.
	area := edgeFunction(x3, y3, x1, y1, x2, y2)
	if area == 0 {
		return
	}

	// Приращения edge-функций на один шаг по x и по y. edgeFunction линейна, поэтому
	// в горячем цикле остаются только сложения — без умножений. Считаются один раз,
	// горутины их только читают.
	stepX0, stepY0 := y1-y2, x2-x1
	stepX1, stepY1 := y2-y3, x3-x2
	stepX2, stepY2 := y3-y1, x1-x3

	// Каждая горутина обрабатывает свой непрерывный диапазон строк (y).
	// Строки не пересекаются, значит и байты в framebuffer.data не пересекаются,
	// поэтому Set можно вызывать параллельно без блокировок.
	rows := bboxMaxY - bboxMinY + 1
	workers := runtime.NumCPU()
	if workers > rows {
		workers = rows
	}

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		yStart := bboxMinY + w*rows/workers
		yEnd := bboxMinY + (w+1)*rows/workers // не включительно
		wg.Add(1)
		go func(yStart, yEnd int) {
			defer wg.Done()
			// Каждая горутина стартует со своего угла (bboxMinX, yStart).
			edge0Row := edgeFunction(bboxMinX, yStart, x1, y1, x2, y2)
			edge1Row := edgeFunction(bboxMinX, yStart, x2, y2, x3, y3)
			edge2Row := edgeFunction(bboxMinX, yStart, x3, y3, x1, y1)
			for y := yStart; y < yEnd; y++ {
				edge0, edge1, edge2 := edge0Row, edge1Row, edge2Row
				for x := bboxMinX; x <= bboxMaxX; x++ {
					if edge0 >= 0 && edge1 >= 0 && edge2 >= 0 || edge0 <= 0 && edge1 <= 0 && edge2 <= 0 {
						b := uint8(float32(edge0) / float32(area) * 255)
						g := uint8(float32(edge1) / float32(area) * 255)
						r := uint8(float32(edge2) / float32(area) * 255)
						framebuffer.Set(x, y, TGAColor{b, g, r, 255})
					}
					edge0 += stepX0
					edge1 += stepX1
					edge2 += stepX2
				}
				edge0Row += stepY0
				edge1Row += stepY1
				edge2Row += stepY2
			}
		}(yStart, yEnd)
	}
	wg.Wait()
}

func renderObjModelWireFrame(model *Model, framebuffer *TGAImage, color TGAColor) {
	for _, face := range model.faces {
		if len(model.vertices) <= face.first || len(model.vertices) <= face.second || len(model.vertices) <= face.third {
			continue // Пропускаем некорректные грани
		}
		v0 := model.vertices[face.first]
		v1 := model.vertices[face.second]
		v2 := model.vertices[face.third]

		v0normalized := Vertex{X: (v0.X + 1.0) * float64(framebuffer.Width()) / 2, Y: (v0.Y + 1.0) * float64(framebuffer.Height()) / 2, Z: v0.Z}
		v1normalized := Vertex{X: (v1.X + 1.0) * float64(framebuffer.Width()) / 2, Y: (v1.Y + 1.0) * float64(framebuffer.Height()) / 2, Z: v1.Z}
		v2normalized := Vertex{X: (v2.X + 1.0) * float64(framebuffer.Width()) / 2, Y: (v2.Y + 1.0) * float64(framebuffer.Height()) / 2, Z: v2.Z}

		// Back-face culling: z-компонента (v1-v0) × (v2-v0) в экранных координатах.
		// Знак показывает направление обхода вершин; грани, смотрящие от камеры,
		// имеют обратный обход — их пропускаем.
		cross := (v1normalized.X-v0normalized.X)*(v2normalized.Y-v0normalized.Y) -
			(v2normalized.X-v0normalized.X)*(v1normalized.Y-v0normalized.Y)
		if cross <= 0 {
			continue
		}

		line(int(v0normalized.X), int(v0normalized.Y), int(v1normalized.X), int(v1normalized.Y), framebuffer, color)
		line(int(v1normalized.X), int(v1normalized.Y), int(v2normalized.X), int(v2normalized.Y), framebuffer, color)
		line(int(v2normalized.X), int(v2normalized.Y), int(v0normalized.X), int(v0normalized.Y), framebuffer, color)
		// modernTriangle(int(v0normalized.X), int(v0normalized.Y), int(v1normalized.X), int(v1normalized.Y), int(v2normalized.X), int(v2normalized.Y), framebuffer, color)
	}
}

func renderObjModelSurface(model *Model, framebuffer *TGAImage, zbuffer *TGAImage) {
	for _, face := range model.faces {
		if len(model.vertices) <= face.first || len(model.vertices) <= face.second || len(model.vertices) <= face.third {
			continue // Пропускаем некорректные грани
		}
		v0 := model.vertices[face.first]
		v1 := model.vertices[face.second]
		v2 := model.vertices[face.third]

		v0normalized := Vertex{X: (v0.X + 1.0) * float64(framebuffer.Width()) / 2, Y: (v0.Y + 1.0) * float64(framebuffer.Height()) / 2, Z: (v0.Z + 1.0) * 255 / 2}
		v1normalized := Vertex{X: (v1.X + 1.0) * float64(framebuffer.Width()) / 2, Y: (v1.Y + 1.0) * float64(framebuffer.Height()) / 2, Z: (v1.Z + 1.0) * 255 / 2}
		v2normalized := Vertex{X: (v2.X + 1.0) * float64(framebuffer.Width()) / 2, Y: (v2.Y + 1.0) * float64(framebuffer.Height()) / 2, Z: (v2.Z + 1.0) * 255 / 2}

		// cross := (v1normalized.X-v0normalized.X)*(v2normalized.Y-v0normalized.Y) -
		// 	(v2normalized.X-v0normalized.X)*(v1normalized.Y-v0normalized.Y)
		// if cross <= 0 {
		// 	continue
		// }

		color := NewColor(uint8(rand.IntN(256)), uint8(rand.IntN(256)), uint8(rand.IntN(256)), 255)
		// color := NewColor(uint8((v0.Z+1.0)*127.5), uint8((v1.Z+1.0)*127.5), uint8((v2.Z+1.0)*127.5), 255)
		modernTriangle(
			int(v0normalized.X), int(v0normalized.Y), int(v0normalized.Z),
			int(v1normalized.X), int(v1normalized.Y), int(v1normalized.Z),
			int(v2normalized.X), int(v2normalized.Y), int(v2normalized.Z),
			framebuffer, zbuffer, color)
		// modernTriangle(int(v0normalized.X), int(v0normalized.Y), int(v1normalized.X), int(v1normalized.Y), int(v2normalized.X), int(v2normalized.Y), framebuffer, color)
	}
}

func main() {
	const (
		width  = 1000
		height = 1000
	)

	framebuffer := NewTGAImage(width, height, RGB)
	zbuffer := NewTGAImage(width, height, Grayscale)
	model := readModel("diablo3_pose.obj")

	renderObjModelSurface(model, framebuffer, zbuffer)
	framebuffer.WritePNGFile(outDir + "/diablo3_culled.png")
}
