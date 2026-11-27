package tiling

import (
	"context"
	"fmt"
	"log"
	"mime"
	"os"
	"os/exec"
	"path/filepath"

	corestorage "mcba/tissquest/internal/core/storage"
)

type Params struct {
	SlideID     uint
	Store       corestorage.MediaStore
	LocalSource string
	JPEGQuality int
	TileSize    int
	UseRGB      bool
}

type Result struct {
	DziURL string
	Bytes  int64
	Tiles  int
}

func Run(p Params) (Result, error) {
	if p.Store == nil {
		return Result{}, fmt.Errorf("storage is required")
	}
	if p.LocalSource == "" {
		return Result{}, fmt.Errorf("local source image is required")
	}
	if p.JPEGQuality <= 0 {
		p.JPEGQuality = 75
	}

	workDir, err := os.MkdirTemp("", fmt.Sprintf("tissquest-tile-%d-*", p.SlideID))
	if err != nil {
		return Result{}, fmt.Errorf("create work dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	outputBase := filepath.Join(workDir, "output")
	if err := runVips(p.LocalSource, outputBase, p.JPEGQuality, p.TileSize, p.UseRGB); err != nil {
		return Result{}, fmt.Errorf("vips dzsave: %w", err)
	}

	ctx := context.Background()
	if err := p.Store.DeletePrefix(ctx, fmt.Sprintf("slides/%d/", p.SlideID)); err != nil {
		return Result{}, fmt.Errorf("delete old tiles: %w", err)
	}

	dziKey := fmt.Sprintf("slides/%d/dzi/output.dzi", p.SlideID)
	tilesPrefix := fmt.Sprintf("slides/%d/dzi/output_files", p.SlideID)

	dziBytes, err := uploadFile(ctx, p.Store, outputBase+".dzi", dziKey, "application/xml")
	if err != nil {
		return Result{}, fmt.Errorf("upload .dzi: %w", err)
	}

	tilesBytes, tiles, err := uploadDir(ctx, p.Store, outputBase+"_files", tilesPrefix)
	if err != nil {
		return Result{}, fmt.Errorf("upload tiles: %w", err)
	}

	return Result{
		DziURL: p.Store.PublicURL(dziKey),
		Bytes:  dziBytes + tilesBytes,
		Tiles:  tiles,
	}, nil
}

func runVips(srcPath, outputBase string, quality int, tileSize int, useRGB bool) error {
	effectiveSrc := srcPath
	if useRGB {
		effectiveSrc = srcPath + "[rgb]"
	}
	cmd := exec.Command("vips", "dzsave",
		effectiveSrc,
		outputBase,
		"--tile-size", fmt.Sprintf("%d", tileSize),
		"--overlap", "1",
		"--suffix", fmt.Sprintf(".jpg[Q=%d]", quality),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("running: %s", cmd.String())
	return cmd.Run()
}

func uploadFile(ctx context.Context, store corestorage.MediaStore, localPath, key, contentType string) (int64, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return 0, err
	}

	if _, err := store.Upload(ctx, key, contentType, f, info.Size()); err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func uploadDir(ctx context.Context, store corestorage.MediaStore, localDir, prefix string) (int64, int, error) {
	var total int64
	var count int

	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}

		key := prefix + "/" + filepath.ToSlash(rel)
		contentType := mime.TypeByExtension(filepath.Ext(path))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		size, err := uploadFile(ctx, store, path, key, contentType)
		if err != nil {
			return err
		}
		total += size
		count++
		return nil
	})

	return total, count, err
}
