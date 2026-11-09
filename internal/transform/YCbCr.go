package transform

import "image"

// YCbCr разбивает изображение на три канала для указанной полосы строк.
func YCbCr(src *image.RGBA, yBuf, cbBuf, crBuf *image.Gray, y0, y1 int) {
	bounds := src.Bounds()
	width := bounds.Dx()

	for y := y0; y < y1; y++ {
		srcRow := src.Pix[y*src.Stride : y*src.Stride+width*4]
		yRow := yBuf.Pix[y*yBuf.Stride : y*yBuf.Stride+width]
		cbRow := cbBuf.Pix[y*cbBuf.Stride : y*cbBuf.Stride+width]
		crRow := crBuf.Pix[y*crBuf.Stride : y*crBuf.Stride+width]

		for x := 0; x < width; x++ {
			r := float64(srcRow[x*4+0])
			g := float64(srcRow[x*4+1])
			b := float64(srcRow[x*4+2])

			yVal := 0.299*r + 0.587*g + 0.114*b
			cbVal := 128.0 - 0.168736*r - 0.331264*g + 0.5*b
			crVal := 128.0 + 0.5*r - 0.418688*g - 0.081312*b

			yRow[x] = ClampRound(yVal)
			cbRow[x] = ClampRound(cbVal)
			crRow[x] = ClampRound(crVal)
		}
	}
}
