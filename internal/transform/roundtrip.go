package transform

import "image"

// Roundtrip переводит YCbCr обратно в RGB.
func Roundtrip(yBuf, cbBuf, crBuf *image.Gray, dst *image.RGBA, y0, y1 int) {
	bounds := dst.Bounds()
	width := bounds.Dx()

	for y := y0; y < y1; y++ {
		dstRow := dst.Pix[y*dst.Stride : y*dst.Stride+width*4]
		yRow := yBuf.Pix[y*yBuf.Stride : y*yBuf.Stride+width]
		cbRow := cbBuf.Pix[y*cbBuf.Stride : y*cbBuf.Stride+width]
		crRow := crBuf.Pix[y*crBuf.Stride : y*crBuf.Stride+width]

		for x := 0; x < width; x++ {
			yVal := float64(yRow[x])
			cbVal := float64(cbRow[x]) - 128.0
			crVal := float64(crRow[x]) - 128.0

			r := yVal + 1.402*crVal
			g := yVal - 0.344136*cbVal - 0.714136*crVal
			b := yVal + 1.772*cbVal

			dstRow[x*4+0] = ClampRound(r)
			dstRow[x*4+1] = ClampRound(g)
			dstRow[x*4+2] = ClampRound(b)
			dstRow[x*4+3] = 255 // Альфа всегда непрозрачная
		}
	}
}
