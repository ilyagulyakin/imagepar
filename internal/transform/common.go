package transform

import "math"

// ClampRound округляет значение к ближайшему целому и обрезает в [0, 255].
// Используется везде, где происходит запись в итоговый буфер.
func ClampRound(v float64) uint8 {
	r := math.Round(v)

	if r < 0 {
		return 0
	}
	if r > 255 {
		return 255
	}

	return uint8(r)
}

// StripBounds возвращает границы строк [y0, y1) для полосы i.
// Общая функция для seq и par, чтобы гарантировать одинаковую нарезку.
func StripBounds(imgHeight, stripH, i int) (y0, y1 int) {
	y0 = i * stripH
	y1 = min(y0+stripH, imgHeight)
	return y0, y1
}
