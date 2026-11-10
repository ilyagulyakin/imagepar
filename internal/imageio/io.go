package imageio

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

// Load читает изображение из файла и всегда возвращает *image.RGBA,
// независимо от исходного формата и цветовой модели.
func Load(path string) (*image.RGBA, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("imageio.Load: open %q: %w", path, err)
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("imageio.Load: decode %q: %w", path, err)
	}

	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
	return dst, nil
}

// SavePNG записывает изображение в PNG. Создаёт директорию, если её нет.
func SavePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("imageio.SavePNG: mkdir for %q: %w", path, err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("imageio.SavePNG: create %q: %w", path, err)
	}

	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("imageio.SavePNG: close %q: %w", path, cerr)
		}
	}()

	w := bufio.NewWriter(f)
	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := enc.Encode(w, img); err != nil {
		return fmt.Errorf("imageio.SavePNG: encode %q: %w", path, err)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("imageio.SavePNG: flush %q: %w", path, err)
	}
	return nil
}
