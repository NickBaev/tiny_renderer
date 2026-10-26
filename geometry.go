package main

import "math"

type Vec2 struct{ X, Y float64 }

type Vec3 struct{ X, Y, Z float64 }

type Vec4 struct{ X, Y, Z, W float64 }

func (a Vec3) Add(b Vec3) Vec3      { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) Sub(b Vec3) Vec3      { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) Scale(k float64) Vec3 { return Vec3{a.X * k, a.Y * k, a.Z * k} }
func (a Vec3) Dot(b Vec3) float64   { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }
func (a Vec3) Norm() float64        { return math.Sqrt(a.Dot(a)) }
func (a Vec3) Normalized() Vec3     { return a.Scale(1 / a.Norm()) }
func (a Vec3) XYZ1() Vec4           { return Vec4{a.X, a.Y, a.Z, 1} }
func (a Vec4) Scale(k float64) Vec4 { return Vec4{a.X * k, a.Y * k, a.Z * k, a.W * k} }
func (a Vec4) XY() Vec2             { return Vec2{a.X, a.Y} }
func (a Vec4) XYZ() Vec3            { return Vec3{a.X, a.Y, a.Z} }
func (a Vec4) arr() [4]float64      { return [4]float64{a.X, a.Y, a.Z, a.W} }
func vec4(a [4]float64) Vec4        { return Vec4{a[0], a[1], a[2], a[3]} }
func (a Vec3) Cross(b Vec3) Vec3 {
	return Vec3{a.Y*b.Z - a.Z*b.Y, a.Z*b.X - a.X*b.Z, a.X*b.Y - a.Y*b.X}
}

// Mat3 хранится по строкам.
type Mat3 [3]Vec3

func (m Mat3) Det() float64 { return m[0].Dot(m[1].Cross(m[2])) }

// InvertTranspose возвращает (M⁻¹)ᵀ. Для 3×3 строки матрицы алгебраических
// дополнений — это попарные векторные произведения строк M.
func (m Mat3) InvertTranspose() Mat3 {
	k := 1 / m.Det()
	return Mat3{
		m[1].Cross(m[2]).Scale(k),
		m[2].Cross(m[0]).Scale(k),
		m[0].Cross(m[1]).Scale(k),
	}
}

func (m Mat3) MulVec(v Vec3) Vec3 { return Vec3{m[0].Dot(v), m[1].Dot(v), m[2].Dot(v)} }

// Mat4 хранится по строкам: m[row][col].
type Mat4 [4][4]float64

func (a Mat4) Mul(b Mat4) Mat4 {
	var r Mat4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				r[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return r
}

func (m Mat4) MulVec(v Vec4) Vec4 {
	a := v.arr()
	var r [4]float64
	for i := 0; i < 4; i++ {
		for k := 0; k < 4; k++ {
			r[i] += m[i][k] * a[k]
		}
	}
	return vec4(r)
}
