package main

import (
	"math/rand"
	"testing"
)

func BenchmarkLine(b *testing.B) {
	fb := NewTGAImage(64, 64, RGB)
	c := NewColor(uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		line(rand.Intn(64), rand.Intn(64), rand.Intn(64), rand.Intn(64), fb, c)
	}
}
