package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Model struct {
	verts         []Vec3   // массив вершин
	normals       []Vec3   // массив нормалей
	uvs           []Vec2   // массив текстурных координат
	faces         [][3]int // индексы вершин каждого треугольника
	normalIndices [][3]int // индексы нормалей каждого треугольника
	uvIndices     [][3]int // индексы текстурных координат каждого треугольника
}

func readModel(filename string) (*Model, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	model := &Model{}
	scanner := bufio.NewScanner(file)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "v "):
			var v Vec3
			if _, err := fmt.Sscanf(line, "v %f %f %f", &v.X, &v.Y, &v.Z); err != nil {
				return nil, fmt.Errorf("%s:%d: %w", filename, lineNo, err)
			}
			model.verts = append(model.verts, v)
		case strings.HasPrefix(line, "vn "):
			var n Vec3
			if _, err := fmt.Sscanf(line, "vn %f %f %f", &n.X, &n.Y, &n.Z); err != nil {
				return nil, fmt.Errorf("%s:%d: %w", filename, lineNo, err)
			}
			model.normals = append(model.normals, n)
		case strings.HasPrefix(line, "vt "):
			var uv Vec2
			if _, err := fmt.Sscanf(line, "vt %f %f", &uv.X, &uv.Y); err != nil {
				return nil, fmt.Errorf("%s:%d: %w", filename, lineNo, err)
			}
			model.uvs = append(model.uvs, uv)
		case strings.HasPrefix(line, "f "):
			// Формат грани: f v/vt/vn v/vt/vn v/vt/vn.
			var face, normalFace, uvFace [3]int
			if _, err := fmt.Sscanf(line, "f %d/%d/%d %d/%d/%d %d/%d/%d",
				&face[0], &uvFace[0], &normalFace[0],
				&face[1], &uvFace[1], &normalFace[1],
				&face[2], &uvFace[2], &normalFace[2]); err != nil {
				return nil, fmt.Errorf("%s:%d: %w", filename, lineNo, err)
			}
			for i := range face {
				face[i]--       // в OBJ индексы с единицы
				normalFace[i]-- // в OBJ индексы с единицы
				uvFace[i]--     // в OBJ индексы с единицы
			}
			model.faces = append(model.faces, face)
			model.normalIndices = append(model.normalIndices, normalFace)
			model.uvIndices = append(model.uvIndices, uvFace)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for i, face := range model.faces {
		for _, v := range face {
			if v < 0 || v >= len(model.verts) {
				return nil, fmt.Errorf("%s: грань %d ссылается на несуществующую вершину %d", filename, i, v+1)
			}
		}
	}
	for i, face := range model.normalIndices {
		for _, n := range face {
			if n < 0 || n >= len(model.normals) {
				return nil, fmt.Errorf("%s: грань %d ссылается на несуществующую нормаль %d", filename, i, n+1)
			}
		}
	}
	for i, face := range model.uvIndices {
		for _, uv := range face {
			if uv < 0 || uv >= len(model.uvs) {
				return nil, fmt.Errorf("%s: грань %d ссылается на несуществующие текстурные координаты %d", filename, i, uv+1)
			}
		}
	}
	return model, nil
}

func (m *Model) nverts() int { return len(m.verts) }
func (m *Model) nfaces() int { return len(m.faces) }

func (m *Model) vert(i int) Vec3 { return m.verts[i] }

// faceVert возвращает nthvert-ю вершину (0..2) грани iface.
func (m *Model) faceVert(iface, nthvert int) Vec3 { return m.verts[m.faces[iface][nthvert]] }

// faceNormal возвращает нормаль nthvert-й вершины (0..2) грани iface.
func (m *Model) faceNormal(iface, nthvert int) Vec3 {
	return m.normals[m.normalIndices[iface][nthvert]]
}

// faceUV возвращает текстурные координаты nthvert-й вершины (0..2) грани iface.
func (m *Model) faceUV(iface, nthvert int) Vec2 {
	return m.uvs[m.uvIndices[iface][nthvert]]
}
