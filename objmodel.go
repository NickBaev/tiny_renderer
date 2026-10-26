package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Model struct {
	verts []Vec3   // массив вершин
	faces [][3]int // индексы вершин каждого треугольника
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
		case strings.HasPrefix(line, "f "):
			// Формат грани: f v/vt/vn v/vt/vn v/vt/vn. Нам нужны только индексы
			// вершин (первый номер каждой группы); индексы текстур и нормалей пропускаем.
			var face [3]int
			var vt, vn int
			if _, err := fmt.Sscanf(line, "f %d/%d/%d %d/%d/%d %d/%d/%d",
				&face[0], &vt, &vn, &face[1], &vt, &vn, &face[2], &vt, &vn); err != nil {
				return nil, fmt.Errorf("%s:%d: %w", filename, lineNo, err)
			}
			for i := range face {
				face[i]-- // в OBJ индексы с единицы
			}
			model.faces = append(model.faces, face)
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
	return model, nil
}

func (m *Model) nverts() int { return len(m.verts) }
func (m *Model) nfaces() int { return len(m.faces) }

func (m *Model) vert(i int) Vec3 { return m.verts[i] }

// faceVert возвращает nthvert-ю вершину (0..2) грани iface.
func (m *Model) faceVert(iface, nthvert int) Vec3 { return m.verts[m.faces[iface][nthvert]] }
