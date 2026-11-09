package transform

import (
	"image"
	"math"
)

// Sobel вычисляет градиент для полосы.
func Sobel(src *image.Gray, dst *image.Gray, y0, y1 int) {
	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	for y := y0; y < y1; y++ {
		dstRow := dst.Pix[y*dst.Stride : y*dst.Stride+w]

		// Зажимаем индексы строк для "повтора края"
		yPrev := y - 1
		if yPrev < 0 {
			yPrev = 0
		}
		yNext := y + 1
		if yNext >= h {
			yNext = h - 1
		}

		rowPrev := src.Pix[yPrev*src.Stride : yPrev*src.Stride+w]
		rowCur := src.Pix[y*src.Stride : y*src.Stride+w]
		rowNext := src.Pix[yNext*src.Stride : yNext*src.Stride+w]

		for x := 0; x < w; x++ {
			xPrev := x - 1
			if xPrev < 0 {
				xPrev = 0
			}
			xNext := x + 1
			if xNext >= w {
				xNext = w - 1
			}

			// Ядра Собеля
			gx := float64(rowPrev[xPrev]) + 2.0*float64(rowCur[xPrev]) + float64(rowNext[xPrev]) -
				float64(rowPrev[xNext]) - 2.0*float64(rowCur[xNext]) - float64(rowNext[xNext])

			gy := float64(rowPrev[xPrev]) + 2.0*float64(rowPrev[x]) + float64(rowPrev[xNext]) -
				float64(rowNext[xPrev]) - 2.0*float64(rowNext[x]) - float64(rowNext[xNext])

			// Модуль градиента
			mag := math.Hypot(gx, gy)
			dstRow[x] = ClampRound(mag)
		}
	}
}
