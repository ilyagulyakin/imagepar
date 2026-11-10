package app

import (
	"image"
	"path/filepath"

	"imgpar/internal/imageio"
	"imgpar/internal/pool"
	"imgpar/internal/transform"
)

type Buffers struct {
	Gray, Y, Cb, Cr, Sobel *image.Gray
	RT, Distort            *image.RGBA
}

func NewBuffers(b image.Rectangle) Buffers {
	return Buffers{
		Gray:    image.NewGray(b),
		Y:       image.NewGray(b),
		Cb:      image.NewGray(b),
		Cr:      image.NewGray(b),
		Sobel:   image.NewGray(b),
		RT:      image.NewRGBA(b),
		Distort: image.NewRGBA(b),
	}
}

func Build(sched *pool.Scheduler, src *image.RGBA, bufs Buffers, stripH int, k float64, outDir string) {
	H := src.Bounds().Dy()
	S := (H + stripH - 1) / stripH

	grayJobs := make([]*pool.Job, S)
	ycbcrJobs := make([]*pool.Job, S)
	distortJobs := make([]*pool.Job, S)
	sobelJobs := make([]*pool.Job, S)
	rtJobs := make([]*pool.Job, S)

	// 1. Создаём все задачи с известным deps
	for i := 0; i < S; i++ {
		i := i
		y0, y1 := transform.StripBounds(H, stripH, i)

		grayJobs[i] = sched.NewJob(func() {
			transform.Gray(src, bufs.Gray, y0, y1)
		}, 0)

		ycbcrJobs[i] = sched.NewJob(func() {
			transform.YCbCr(src, bufs.Y, bufs.Cb, bufs.Cr, y0, y1)
		}, 0)

		distortJobs[i] = sched.NewJob(func() {
			transform.Distort(src, bufs.Distort, y0, y1, k)
		}, 0)

		var deps int32 = 1 // своя полоса серая
		if i > 0 {
			deps++
		}
		if i < S-1 {
			deps++
		}
		sobelJobs[i] = sched.NewJob(func() {
			transform.Sobel(bufs.Gray, bufs.Sobel, y0, y1)
		}, deps)

		rtJobs[i] = sched.NewJob(func() {
			transform.Roundtrip(bufs.Y, bufs.Cb, bufs.Cr, bufs.RT, y0, y1)
		}, 1) // Очевидно обратный перевод возможен после прямого, поэтому deps 1
	}

	// 2. Связываем sobel с соседними полосами gray (с проверкой границ)
	for i := 0; i < S; i++ {
		if i > 0 {
			pool.AddDependency(grayJobs[i-1], sobelJobs[i])
		}
		pool.AddDependency(grayJobs[i], sobelJobs[i])
		if i < S-1 {
			pool.AddDependency(grayJobs[i+1], sobelJobs[i])
		}
		pool.AddDependency(ycbcrJobs[i], rtJobs[i])
	}

	// 3. Сохранения
	saveGray := sched.NewJob(func() {
		imageio.SavePNG(filepath.Join(outDir, "gray.png"), bufs.Gray) // без grayToRGBA
	}, int32(S))
	saveY := sched.NewJob(func() {
		imageio.SavePNG(filepath.Join(outDir, "ycbcr_y.png"), bufs.Y)
	}, int32(S))
	saveCb := sched.NewJob(func() {
		imageio.SavePNG(filepath.Join(outDir, "ycbcr_cb.png"), bufs.Cb)
	}, int32(S))
	saveCr := sched.NewJob(func() {
		imageio.SavePNG(filepath.Join(outDir, "ycbcr_cr.png"), bufs.Cr)
	}, int32(S))
	saveSobel := sched.NewJob(func() {
		imageio.SavePNG(filepath.Join(outDir, "sobel.png"), bufs.Sobel)
	}, int32(S))
	saveRT := sched.NewJob(func() {
		imageio.SavePNG(filepath.Join(outDir, "ycbcr_roundtrip.png"), bufs.RT) // RT — уже *image.RGBA
	}, int32(S))
	saveDistort := sched.NewJob(func() {
		imageio.SavePNG(filepath.Join(outDir, "distort.png"), bufs.Distort) // Distort — уже *image.RGBA
	}, int32(S))

	for i := 0; i < S; i++ {
		pool.AddDependency(grayJobs[i], saveGray)
		pool.AddDependency(ycbcrJobs[i], saveY)
		pool.AddDependency(ycbcrJobs[i], saveCb)
		pool.AddDependency(ycbcrJobs[i], saveCr)
		pool.AddDependency(sobelJobs[i], saveSobel)
		pool.AddDependency(rtJobs[i], saveRT)
		pool.AddDependency(distortJobs[i], saveDistort)
	}

	sched.Start()
}
