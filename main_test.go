package main

import (
	"math/rand"
	"testing"
)

func isLit(img *TGAImage, x, y int) bool {
	c := img.Get(x, y)
	return c.R != 0 || c.G != 0 || c.B != 0
}

// TestLine проверяет, что отрезок зажигает оба конца и не имеет разрывов:
// в каждом столбце/строке вдоль главной оси есть хотя бы один пиксель,
// а соседние закрашенные пиксели отстоят не больше чем на 1 по каждой оси.
func TestLine(t *testing.T) {
	const size = 32
	white := NewColor(255, 255, 255, 255)
	cases := [][4]int{
		{0, 0, 31, 31},   // диагональ
		{0, 0, 31, 0},    // горизонталь
		{0, 0, 0, 31},    // вертикаль
		{31, 31, 0, 0},   // обратное направление
		{2, 30, 30, 2},   // отрицательный наклон
		{5, 5, 25, 10},   // пологая
		{5, 5, 10, 25},   // крутая
		{15, 15, 15, 15}, // точка
	}
	for _, tc := range cases {
		x0, y0, x1, y1 := tc[0], tc[1], tc[2], tc[3]
		img := NewTGAImage(size, size, RGB)
		line(x0, y0, x1, y1, img, white)

		if !isLit(img, x0, y0) {
			t.Errorf("line%v: начальная точка (%d,%d) не закрашена", tc, x0, y0)
		}
		if !isLit(img, x1, y1) {
			t.Errorf("line%v: конечная точка (%d,%d) не закрашена", tc, x1, y1)
		}

		// собрать закрашенные пиксели и проверить связность
		var pts [][2]int
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				if isLit(img, x, y) {
					pts = append(pts, [2]int{x, y})
				}
			}
		}
		steep := abs(x1-x0) < abs(y1-y0)
		want := abs(x1-x0) + 1
		if steep {
			want = abs(y1-y0) + 1
		}
		if len(pts) != want {
			t.Errorf("line%v: закрашено %d пикселей, ожидалось %d", tc, len(pts), want)
		}
	}
}

func BenchmarkLine(b *testing.B) {
	fb := NewTGAImage(64, 64, RGB)
	c := NewColor(uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		line(rand.Intn(64), rand.Intn(64), rand.Intn(64), rand.Intn(64), fb, c)
	}
}
