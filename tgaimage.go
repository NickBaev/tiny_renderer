package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
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
	switch img.format {
	case Grayscale:
		img.data[idx] = c.B
	case RGB:
		img.data[idx] = c.B
		img.data[idx+1] = c.G
		img.data[idx+2] = c.R
	case RGBA:
		img.data[idx] = c.B
		img.data[idx+1] = c.G
		img.data[idx+2] = c.R
		img.data[idx+3] = c.A
	}
}

// Fill заливает всё изображение цветом c.
func (img *TGAImage) Fill(c TGAColor) {
	for y := 0; y < img.height; y++ {
		for x := 0; x < img.width; x++ {
			img.Set(x, y, c)
		}
	}
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

// ReadTGAFile загружает несжатое или RLE-сжатое изображение Grayscale/RGB/RGBA.
func ReadTGAFile(filename string) (*TGAImage, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("открытие TGA: %w", err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	var header tgaHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("чтение заголовка TGA: %w", err)
	}
	if header.Width == 0 || header.Height == 0 {
		return nil, fmt.Errorf("некорректный размер TGA: %d×%d", header.Width, header.Height)
	}
	if header.ColorMapType != 0 {
		return nil, fmt.Errorf("TGA с цветовой палитрой не поддерживается")
	}
	if header.ImageDescriptor&0x10 != 0 {
		return nil, fmt.Errorf("TGA с направлением справа налево не поддерживается")
	}

	var format Format
	switch header.BitsPerPixel {
	case 8:
		format = Grayscale
	case 24:
		format = RGB
	case 32:
		format = RGBA
	default:
		return nil, fmt.Errorf("глубина TGA %d бит не поддерживается", header.BitsPerPixel)
	}

	switch header.DataTypeCode {
	case 2, 10:
		if format == Grayscale {
			return nil, fmt.Errorf("несоответствие типа и глубины TGA")
		}
	case 3, 11:
		if format != Grayscale {
			return nil, fmt.Errorf("несоответствие типа и глубины TGA")
		}
	default:
		return nil, fmt.Errorf("тип TGA %d не поддерживается", header.DataTypeCode)
	}

	if _, err := io.CopyN(io.Discard, r, int64(header.IDLength)); err != nil {
		return nil, fmt.Errorf("пропуск идентификатора TGA: %w", err)
	}

	img := NewTGAImage(int(header.Width), int(header.Height), format)
	if header.DataTypeCode == 10 || header.DataTypeCode == 11 {
		if err := readTGARLE(r, img.data, int(format)); err != nil {
			return nil, fmt.Errorf("распаковка TGA: %w", err)
		}
	} else if _, err := io.ReadFull(r, img.data); err != nil {
		return nil, fmt.Errorf("чтение пикселей TGA: %w", err)
	}

	// При верхнем начале координат переворачиваем строки во внутренний формат снизу вверх.
	if header.ImageDescriptor&0x20 != 0 {
		rowSize := img.width * int(img.format)
		row := make([]byte, rowSize)
		for y := 0; y < img.height/2; y++ {
			top := y * rowSize
			bottom := (img.height - 1 - y) * rowSize
			copy(row, img.data[top:top+rowSize])
			copy(img.data[top:top+rowSize], img.data[bottom:bottom+rowSize])
			copy(img.data[bottom:bottom+rowSize], row)
		}
	}
	return img, nil
}

func readTGARLE(r io.Reader, data []byte, bytesPerPixel int) error {
	pixel := make([]byte, bytesPerPixel)
	for offset := 0; offset < len(data); {
		var packet uint8
		if err := binary.Read(r, binary.LittleEndian, &packet); err != nil {
			return err
		}
		count := int(packet&0x7f) + 1
		size := count * bytesPerPixel
		if size > len(data)-offset {
			return fmt.Errorf("RLE-пакет выходит за границы изображения")
		}

		if packet&0x80 != 0 {
			if _, err := io.ReadFull(r, pixel); err != nil {
				return err
			}
			for range count {
				copy(data[offset:offset+bytesPerPixel], pixel)
				offset += bytesPerPixel
			}
		} else {
			if _, err := io.ReadFull(r, data[offset:offset+size]); err != nil {
				return err
			}
			offset += size
		}
	}
	return nil
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

// WritePNGFile сохраняет изображение в PNG-файл. Так как в TGA начало координат
// в левом нижнем углу, строки переворачиваются по вертикали под формат PNG.
func (img *TGAImage) WritePNGFile(filename string) error {
	out := image.NewNRGBA(image.Rect(0, 0, img.width, img.height))
	for y := 0; y < img.height; y++ {
		for x := 0; x < img.width; x++ {
			c := img.Get(x, y)
			out.SetNRGBA(x, img.height-1-y, color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A})
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	if err := png.Encode(w, out); err != nil {
		return fmt.Errorf("кодирование PNG: %w", err)
	}
	return w.Flush()
}
