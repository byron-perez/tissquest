// tile-pipeline generates slide image variants and Deep Zoom Image tiles,
// uploads them to the configured S3-compatible store, and updates the slide record.
//
//	tile-pipeline -slide <id> -source /path/to/image
//	tile-pipeline -slide <id> -source /path/to/image -magnification 40 -microns-per-pixel 0.25
//	tile-pipeline -slide <id> -delete-media
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"mcba/tissquest/cmd/tile-pipeline/tiling"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/media"
	"mcba/tissquest/internal/persistence/repositories"
	persistencestorage "mcba/tissquest/internal/persistence/storage"
	"mcba/tissquest/internal/services"
)

func main() {
	slideID := flag.Uint("slide", 0, "ID of the slide to process")
	baseMag := flag.Int("magnification", 0, "Base magnification; defaults to the slide value")
	mpp := flag.Float64("microns-per-pixel", 0, "Microns per pixel; defaults to the slide value")
	localSource := flag.String("source", "", "Path to a local image file")
	quality := flag.Int("jpeg-quality", 75, "JPEG quality for DZI tiles")
	deleteMedia := flag.Bool("delete-media", false, "Delete stored media for the slide")
	tileSize := flag.Int("tile-size", 512, "Tile size for DZI (pixels)")
	useRGB := flag.Bool("rgb", true, "Use [rgb] loader option for WSI (default true)")
	sourceURL := flag.String("source-url", "", "URL of the source image (for provenance)")
	sourceTitle := flag.String("source-title", "", "Title / description of the source")
	sourceAuthor := flag.String("source-author", "", "Author / uploader")
	sourceLicense := flag.String("source-license", "", "License name, e.g. CC BY 4.0")
	sourceNotes := flag.String("source-notes", "", "Free‑form notes (cropping, scale, etc.)")
	flag.Parse()

	if *slideID == 0 {
		fmt.Fprintln(os.Stderr, "error: provide -slide <id>")
		flag.Usage()
		os.Exit(1)
	}

	_ = godotenv.Load()

	store, err := persistencestorage.NewS3StorageFromEnv()
	if err != nil {
		log.Fatalf("storage init failed: %v", err)
	}

	if *deleteMedia {
		if err := deleteSlideMedia(uint(*slideID), store); err != nil {
			log.Fatalf("delete media failed: %v", err)
		}
		fmt.Printf("deleted media for slide %d\n", *slideID)
		return
	}

	if *localSource == "" {
		fmt.Fprintln(os.Stderr, "error: -source is required")
		flag.Usage()
		os.Exit(1)
	}
	if *quality < 1 || *quality > 100 {
		log.Fatalf("invalid -jpeg-quality %d: must be between 1 and 100", *quality)
	}

	svc := services.NewSlideService(repositories.NewSlideRepository())
	sl, err := svc.GetByID(uint(*slideID))
	if err != nil {
		log.Fatalf("slide %d not found: %v", *slideID, err)
	}

	mag := *baseMag
	if mag == 0 {
		mag = sl.Magnification
	}
	if mag == 0 {
		mag = 40
	}

	slideMPP := *mpp
	if slideMPP == 0 {
		slideMPP = sl.MicronsPerPixel
	}
	if slideMPP == 0 {
		slideMPP = 0.25
	}

	if err := runSingle(svc, store, uint(*slideID), mag, slideMPP, *localSource, *quality, *tileSize, *useRGB, *sourceURL, *sourceTitle, *sourceAuthor, *sourceLicense, *sourceNotes); err != nil {
		log.Fatal(err)
	}
}

func runSingle(svc *services.SlideService, store *persistencestorage.S3Storage, slideID uint, baseMag int, mpp float64, source string, quality int, tileSize int, useRGB bool, sourceURL string, sourceTitle string, sourceAuthor string, sourceLicense string, sourceNotes string) error {
	workDir, err := os.MkdirTemp("", fmt.Sprintf("tissquest-variants-%d-*", slideID))
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	thumbPath := filepath.Join(workDir, "thumb.jpg")
	previewPath := filepath.Join(workDir, "preview.jpg")

	if err := media.Thumbnail(source, thumbPath, 300, 200, 80); err != nil {
		return fmt.Errorf("generate thumbnail: %w", err)
	}
	if err := media.Thumbnail(source, previewPath, 800, 600, 85); err != nil {
		return fmt.Errorf("generate preview: %w", err)
	}

	ctx := context.Background()
	thumbURL, thumbBytes, err := uploadFile(ctx, store, thumbPath, fmt.Sprintf("slides/thumb/%d.jpg", slideID), "image/jpeg")
	if err != nil {
		return fmt.Errorf("upload thumbnail: %w", err)
	}

	previewURL, previewBytes, err := uploadFile(ctx, store, previewPath, fmt.Sprintf("slides/preview/%d.jpg", slideID), "image/jpeg")
	if err != nil {
		return fmt.Errorf("upload preview: %w", err)
	}

	if err := svc.SetImageVariant(slideID, slide.ImageSizeThumb, thumbURL); err != nil {
		return fmt.Errorf("save thumbnail variant: %w", err)
	}
	if err := svc.SetImageVariant(slideID, slide.ImageSizePreview, previewURL); err != nil {
		return fmt.Errorf("save preview variant: %w", err)
	}

	result, err := tiling.Run(tiling.Params{
		SlideID:     slideID,
		Store:       store,
		LocalSource: source,
		JPEGQuality: quality,
		TileSize:    tileSize,
		UseRGB:      useRGB,
	})
	if err != nil {
		return fmt.Errorf("tiling failed: %w", err)
	}

	if err := svc.SetDziMetadata(slideID, result.DziURL, baseMag, mpp); err != nil {
		return fmt.Errorf("save DZI metadata: %w", err)
	}
	if err := svc.SetImageKey(slideID, fmt.Sprintf("slides/%d", slideID)); err != nil {
		return fmt.Errorf("save image key: %w", err)
	}
	if err := svc.SetSourceMetadata(slideID, sourceURL, sourceTitle, sourceAuthor, sourceLicense, sourceNotes); err != nil {
		return fmt.Errorf("save source metadata: %w", err)
	}

	fmt.Printf("slide %d processed\n", slideID)
	fmt.Printf("  thumbnail: %s (%d bytes)\n", thumbURL, thumbBytes)
	fmt.Printf("  preview:   %s (%d bytes)\n", previewURL, previewBytes)
	fmt.Printf("  dzi:       %s\n", result.DziURL)
	fmt.Printf("  tiles:     %d (%d bytes)\n", result.Tiles, result.Bytes)
	fmt.Printf("  total:     %d bytes\n", thumbBytes+previewBytes+result.Bytes)
	return nil
}

func uploadFile(ctx context.Context, store *persistencestorage.S3Storage, localPath, key, contentType string) (string, int64, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", 0, err
	}

	url, err := store.Upload(ctx, key, contentType, f, info.Size())
	if err != nil {
		return "", 0, err
	}
	return url, info.Size(), nil
}

func deleteSlideMedia(slideID uint, store *persistencestorage.S3Storage) error {
	ctx := context.Background()
	keys := []string{
		fmt.Sprintf("slides/thumb/%d.jpg", slideID),
		fmt.Sprintf("slides/preview/%d.jpg", slideID),
	}
	for _, key := range keys {
		if err := store.Delete(ctx, key); err != nil {
			return err
		}
	}
	return store.DeletePrefix(ctx, fmt.Sprintf("slides/%d/", slideID))
}
