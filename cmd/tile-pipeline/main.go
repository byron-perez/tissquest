// tile-pipeline converts slide source images into Deep Zoom Image (DZI) tile sets,
// stores them in S3, and updates the slide records with the DZI URLs.
//
// Single mode (one slide):
//
//	tile-pipeline -slide <id> -magnification <int> -microns-per-pixel <float>
//	tile-pipeline -slide <id> -magnification <int> -microns-per-pixel <float> -source /path/to/image.jpg
//
// Batch mode (all slides with an image but no DZI):
//
//	tile-pipeline -batch [-magnification <int>] [-microns-per-pixel <float>]
//
// Prerequisites:
//   - libvips must be installed (pacman -S libvips / apt install libvips-tools / brew install vips)
//   - .env file present in the working directory
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"mcba/tissquest/cmd/tile-pipeline/tiling"
	"mcba/tissquest/internal/persistence/migration"
	"mcba/tissquest/internal/persistence/repositories"
	persistencestorage "mcba/tissquest/internal/persistence/storage"
	"mcba/tissquest/internal/services"
)

func main() {
	slideID    := flag.Uint("slide", 0, "ID of the slide to tile (single mode)")
	batch      := flag.Bool("batch", false, "Tile all slides that have an image but no DZI yet")
	baseMag    := flag.Int("magnification", 40, "Base magnification used at capture, e.g. 40")
	mpp        := flag.Float64("microns-per-pixel", 0.25, "Spatial calibration: µm per pixel")
	localSource := flag.String("source", "", "Path to a local image file (single mode only)")
	flag.Parse()

	if !*batch && *slideID == 0 {
		fmt.Fprintln(os.Stderr, "error: provide -slide <id> for single mode or -batch for batch mode")
		flag.Usage()
		os.Exit(1)
	}

	if err := godotenv.Load(); err != nil {
		log.Fatal("could not load .env file")
	}

	migration.RunMigration()

	s3, err := persistencestorage.NewS3Storage(
		os.Getenv("AWS_REGION"),
		os.Getenv("S3_BUCKET"),
		os.Getenv("AWS_ACCESS_KEY_ID"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
	)
	if err != nil {
		log.Fatalf("S3 init failed: %v", err)
	}

	svc := services.NewSlideService(s3, repositories.NewSlideRepository())

	if *batch {
		runBatch(svc, s3, *baseMag, *mpp)
	} else {
		runSingle(svc, s3, *slideID, *baseMag, *mpp, *localSource)
	}
}

func runSingle(svc *services.SlideService, s3 *persistencestorage.S3Storage, slideID uint, baseMag int, mpp float64, localSource string) {
	dziURL, err := tiling.Run(tiling.Params{
		SlideID:           slideID,
		BaseMagnification: baseMag,
		MicronsPerPixel:   mpp,
		S3:                s3,
		Region:            os.Getenv("AWS_REGION"),
		Bucket:            os.Getenv("S3_BUCKET"),
		LocalSource:       localSource,
	})
	if err != nil {
		log.Fatalf("tiling failed: %v", err)
	}
	if err := svc.SetDziMetadata(slideID, dziURL, baseMag, mpp); err != nil {
		log.Fatalf("failed to update slide record: %v", err)
	}
	fmt.Printf("✓ slide %d tiled\n  dzi_url: %s\n", slideID, dziURL)
}

func runBatch(svc *services.SlideService, s3 *persistencestorage.S3Storage, baseMag int, mpp float64) {
	repo := repositories.NewSlideRepository()
	pending, err := repo.GetPendingTiling()
	if err != nil {
		log.Fatalf("failed to query pending slides: %v", err)
	}

	if len(pending) == 0 {
		fmt.Println("✓ no slides pending tiling")
		return
	}

	fmt.Printf("found %d slide(s) pending tiling\n\n", len(pending))

	ok, failed := 0, 0
	for _, sl := range pending {
		// Use the slide's own magnification if set, otherwise fall back to the flag value
		mag := sl.Magnification
		if mag == 0 {
			mag = baseMag
		}
		slideMpp := sl.MicronsPerPixel
		if slideMpp == 0 {
			slideMpp = mpp
		}

		fmt.Printf("→ slide %d (%s, %d×)... ", sl.ID, sl.Name, mag)
		dziURL, err := tiling.Run(tiling.Params{
			SlideID:           sl.ID,
			BaseMagnification: mag,
			MicronsPerPixel:   slideMpp,
			S3:                s3,
			Region:            os.Getenv("AWS_REGION"),
			Bucket:            os.Getenv("S3_BUCKET"),
		})
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			failed++
			continue
		}
		if err := svc.SetDziMetadata(sl.ID, dziURL, mag, slideMpp); err != nil {
			fmt.Printf("FAILED (db update): %v\n", err)
			failed++
			continue
		}
		fmt.Println("✓")
		ok++
	}

	fmt.Printf("\nbatch complete: %d tiled, %d failed\n", ok, failed)
	if failed > 0 {
		os.Exit(1)
	}
}
