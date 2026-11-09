package transform

import (
	"image"
	"math"
)

// Distort применяет бочкообразную дисторсию к полосе изображения.
func Distort(src *image.RGBA, dst *image.RGBA, y0, y1 int, k float64) {
	bounds := src.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())
	cx := w / 2.0
	cy := h / 2.0

	for y := y0; y < y1; y++ {
		dstRow := dst.Pix[y*dst.Stride : y*dst.Stride+bounds.Dx()*4]
		fy := float64(y)

		for x := 0; x < bounds.Dx(); x++ {
			fx := float64(x)

			// Нормированные координаты
			nx := (fx - cx) / cx
			ny := (fy - cy) / cy
			r2 := nx*nx + ny*ny
			factor := 1.0 + k*r2

			// Исходные координаты
			sx := cx + nx*factor*cx
			sy := cy + ny*factor*cy

			// Билинейная интерполяция
			x0 := int(math.Floor(sx))
			y0i := int(math.Floor(sy))
			x1 := x0 + 1
			y1i := y0i + 1

			// Проверка границ (если хотя бы один сосед за пределами — черный пиксель)
			if x0 < 0 || x1 >= int(w) || y0i < 0 || y1i >= int(h) {
				dstRow[x*4+0] = 0
				dstRow[x*4+1] = 0
				dstRow[x*4+2] = 0
				dstRow[x*4+3] = 255
				continue
			}

			dx := sx - float64(x0)
			dy := sy - float64(y0i)

			// Веса
			w00 := (1 - dx) * (1 - dy)
			w10 := dx * (1 - dy)
			w01 := (1 - dx) * dy
			w11 := dx * dy

			// Чтение 4 соседей (упрощенно для одного канала, здесь для RGB)
			idx00 := y0i*src.Stride + x0*4
			idx10 := y0i*src.Stride + x1*4
			idx01 := y1i*src.Stride + x0*4
			idx11 := y1i*src.Stride + x1*4

			for c := 0; c < 3; c++ {
				v00 := float64(src.Pix[idx00+c])
				v10 := float64(src.Pix[idx10+c])
				v01 := float64(src.Pix[idx01+c])
				v11 := float64(src.Pix[idx11+c])

				val := w00*v00 + w10*v10 + w01*v01 + w11*v11
				dstRow[x*4+c] = ClampRound(val)
			}
			dstRow[x*4+3] = 255
		}
	}
}
