package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"imgpar/internal/app"
	"imgpar/internal/imageio"
	"imgpar/internal/pool"
)

func main() {
	input := flag.String("input", "", "path to input image (required)")
	outDir := flag.String("out", "result", "output directory")
	mode := flag.String("mode", "seq", "mode: seq or par")
	threads := flag.Int("threads", 4, "number of worker threads")
	stripH := flag.Int("strip", 64, "strip height in rows")
	k := flag.Float64("k", 0.4, "distortion coefficient")
	flag.Parse()

	if *input == "" {
		log.Fatal("error: --input is required")
	}
	if *stripH < 1 {
		log.Fatal("error: --strip must be >= 1")
	}
	if *mode != "seq" && *mode != "par" {
		log.Fatal("error: --mode must be 'seq' or 'par'")
	}
	if *mode == "par" && *threads < 1 {
		log.Fatal("error: --threads must be >= 1")
	}

	t0 := time.Now()
	fmt.Printf("Loading image %s...\n", *input)
	src, err := imageio.Load(*input)
	if err != nil {
		log.Fatalf("failed to load image: %v", err)
	}
	bounds := src.Bounds()
	fmt.Printf("Loaded %dx%d image in %v\n", bounds.Dx(), bounds.Dy(), time.Since(t0))

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("failed to create output dir: %v", err)
	}

	bufs := app.NewBuffers(bounds)

	fmt.Printf("Processing in %s mode...\n", *mode)
	t1 := time.Now()

	switch *mode {
	case "seq":
		sched := pool.NewSyncScheduler()
		app.Build(sched, src, bufs, *stripH, *k, *outDir)
		sched.Wait()

	case "par":
		wp := pool.New(*threads)
		sched := pool.NewScheduler(wp)
		app.Build(sched, src, bufs, *stripH, *k, *outDir)
		sched.Wait()
		wp.Stop()
	}

	fmt.Printf("Processing finished in %v\n", time.Since(t1))
	fmt.Printf("All results saved to %s\n", *outDir)
}
