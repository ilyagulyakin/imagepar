package transform

import "image"

//В Go у *image.RGBA пиксели лежат не как матрица pixels[y][x], а как один длинный слайс Pix []uint8
//Строка 0: [R00,G00,B00,A00, R01,G01,B01,A01, R02,G02,B02,A02]
//Строка 1: [R10,G10,B10,A10, R11,G11,B11,A11, R12,G12,B12,A12] итд

// Gray вычисляет яркость по BT.601 для указанной полосы строк.
func Gray(src *image.RGBA, dst *image.Gray, y0, y1 int) {
	bounds := src.Bounds()
	width := bounds.Dx()

	for y := y0; y < y1; y++ {
		srcRow := src.Pix[y*src.Stride : y*src.Stride+width*4]
		dstRow := dst.Pix[y*dst.Stride : y*dst.Stride+width]

		for x := 0; x < width; x++ {
			r := float64(srcRow[x*4+0])
			g := float64(srcRow[x*4+1])
			b := float64(srcRow[x*4+2])

			gray := 0.299*r + 0.587*g + 0.114*b
			dstRow[x] = ClampRound(gray)
		}
	}
}
