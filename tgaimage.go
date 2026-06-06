package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
)

// Format задаёт количество байт на пиксель.
type Format int

const (
	Grayscale Format = 1 // 1 байт на пиксель
	RGB       Format = 3 // 3 байта: B, G, R
	RGBA      Format = 4 // 4 байта: B, G, R, A
)

// TGAColor хранит цвет в порядке BGRA (как в TGA-формате).
type TGAColor struct {
	B, G, R, A uint8
}

// NewColor — удобный конструктор, принимающий привычный порядок RGBA.
func NewColor(r, g, b, a uint8) TGAColor {
	return TGAColor{B: b, G: g, R: r, A: a}
}

// TGAImage — буфер кадра. Начало координат (0,0) в левом нижнем углу.
type TGAImage struct {
	width  int
	height int
	format Format
	data   []byte
}

// NewTGAImage создаёт чёрное изображение w×h в заданном формате.
func NewTGAImage(w, h int, format Format) *TGAImage {
	return &TGAImage{
		width:  w,
		height: h,
		format: format,
		data:   make([]byte, w*h*int(format)),
	}
}

func (img *TGAImage) Width() int  { return img.width }
func (img *TGAImage) Height() int { return img.height }

// Set закрашивает пиксель (x, y). Координаты вне холста игнорируются.
func (img *TGAImage) Set(x, y int, c TGAColor) {
	if x < 0 || y < 0 || x >= img.width || y >= img.height {
		return
	}
	idx := (x + y*img.width) * int(img.format)
	bytes := []byte{c.B, c.G, c.R, c.A}
	copy(img.data[idx:idx+int(img.format)], bytes[:img.format])
}

// Get возвращает цвет пикселя (x, y).
func (img *TGAImage) Get(x, y int) TGAColor {
	if x < 0 || y < 0 || x >= img.width || y >= img.height {
		return TGAColor{}
	}
	idx := (x + y*img.width) * int(img.format)
	c := TGAColor{}
	switch img.format {
	case Grayscale:
		c.B, c.G, c.R = img.data[idx], img.data[idx], img.data[idx]
		c.A = 255
	case RGB:
		c.B, c.G, c.R, c.A = img.data[idx], img.data[idx+1], img.data[idx+2], 255
	case RGBA:
		c.B, c.G, c.R, c.A = img.data[idx], img.data[idx+1], img.data[idx+2], img.data[idx+3]
	}
	return c
}

// tgaHeader — 18-байтовый заголовок TGA-файла.
type tgaHeader struct {
	IDLength        uint8
	ColorMapType    uint8
	DataTypeCode    uint8
	ColorMapOrigin  uint16
	ColorMapLength  uint16
	ColorMapDepth   uint8
	XOrigin         uint16
	YOrigin         uint16
	Width           uint16
	Height          uint16
	BitsPerPixel    uint8
	ImageDescriptor uint8
}

// WriteTGAFile сохраняет изображение в несжатый TGA-файл.
func (img *TGAImage) WriteTGAFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	header := tgaHeader{
		DataTypeCode: 2, // несжатый true-color (для grayscale было бы 3)
		Width:        uint16(img.width),
		Height:       uint16(img.height),
		BitsPerPixel: uint8(img.format) * 8,
		// бит 5 (0x20) не выставлен → начало координат в левом нижнем углу
	}
	if img.format == Grayscale {
		header.DataTypeCode = 3
	}
	if err := binary.Write(w, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("запись заголовка: %w", err)
	}
	if _, err := w.Write(img.data); err != nil {
		return fmt.Errorf("запись пикселей: %w", err)
	}
	return w.Flush()
}
