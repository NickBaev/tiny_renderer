package main

import (
	"bufio"
	"fmt"
	"os"
)

type Vertex struct {
	X, Y, Z float64
}

type Face struct {
	first, second, third int
}

type Model struct {
	vertices []Vertex
	faces    []Face
}

func readModel(filename string) *Model {
	model := &Model{}
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case 'v':
			if line[1] == ' ' {
				var x, y, z float64
				_, err := fmt.Sscanf(line, "v %f %f %f", &x, &y, &z)
				if err != nil {
					panic(err)
				}
				model.vertices = append(model.vertices, Vertex{X: x, Y: y, Z: z})
			}
		case 'f':
			// Формат грани: f v/vt/vn v/vt/vn v/vt/vn. Нам нужны только индексы
			// вершин (первый номер каждой группы); индексы текстур и нормалей пропускаем.
			var first, second, third int
			var vt, vn int
			_, err := fmt.Sscanf(line, "f %d/%d/%d %d/%d/%d %d/%d/%d",
				&first, &vt, &vn, &second, &vt, &vn, &third, &vt, &vn)
			if err != nil {
				panic(err)
			}
			model.faces = append(model.faces, Face{first: first - 1, second: second - 1, third: third - 1})
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return model
}
