// Package tiling handles the image-to-DZI conversion and S3 upload.
// It is intentionally isolated from the rest of the app so it can be
// tested or replaced without touching the service or handler layers.
package tiling

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	persistencestorage "mcba/tissquest/internal/persistence/storage"
)

// Params holds everything Run needs to do its job.
type Params struct {
	SlideID           uint
	BaseMagnification int
	MicronsPerPixel   float64
	S3                *persistencestorage.S3Storage
	Region            string
	Bucket            string
	// LocalSource is an optional path to a local image file.
	// When set, the pipeline skips the S3 download and uses this file directly.
	// The file is also uploaded to slides/original/<id> in S3 for record-keeping.
	LocalSource string
}

// Run tiles the source image with vips, uploads the result to S3,
// and returns the public URL of the .dzi descriptor.
func Run(p Params) (dziURL string, err error) {
	// 1. Resolve source image — local file or download from S3
	var srcPath string
	var cleanup func()

	if p.LocalSource != "" {
		srcPath = p.LocalSource
		cleanup = func() {} // caller owns the file, we don't delete it

		// Also upload the local file as the canonical original in S3
		originalKey := fmt.Sprintf("slides/original/%d%s", p.SlideID, filepath.Ext(p.LocalSource))
		ct := mime.TypeByExtension(filepath.Ext(p.LocalSource))
		if ct == "" {
			ct = "image/jpeg"
		}
		log.Printf("uploading source image to s3://%s/%s", p.Bucket, originalKey)
		if err := uploadFile(p.S3, p.Bucket, p.LocalSource, originalKey, ct); err != nil {
			return "", fmt.Errorf("upload source image: %w", err)
		}
	} else {
		srcKey := fmt.Sprintf("slides/original/%d.png", p.SlideID)
		log.Printf("downloading source image from s3://%s/%s", p.Bucket, srcKey)
		srcPath, cleanup, err = downloadFromS3(p.S3, p.Bucket, srcKey)
		if err != nil {
			return "", fmt.Errorf("download source image: %w", err)
		}
	}
	defer cleanup()

	// 2. Run vips dzsave in a temp directory
	workDir, err := os.MkdirTemp("", fmt.Sprintf("tissquest-tile-%d-*", p.SlideID))
	if err != nil {
		return "", fmt.Errorf("create work dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	outputBase := filepath.Join(workDir, "output")
	if err := runVips(srcPath, outputBase); err != nil {
		return "", fmt.Errorf("vips dzsave: %w", err)
	}

	// 3. Upload .dzi descriptor and all tiles to S3
	dziKey := fmt.Sprintf("slides/%d/dzi/output.dzi", p.SlideID)
	tilesPrefix := fmt.Sprintf("slides/%d/dzi/output_files", p.SlideID)

	log.Printf("uploading .dzi to s3://%s/%s", p.Bucket, dziKey)
	if err := uploadFile(p.S3, p.Bucket, outputBase+".dzi", dziKey, "application/xml"); err != nil {
		return "", fmt.Errorf("upload .dzi: %w", err)
	}

	tilesDir := outputBase + "_files"
	log.Printf("uploading tiles from %s to s3://%s/%s/", tilesDir, p.Bucket, tilesPrefix)
	if err := uploadDir(p.S3, p.Bucket, tilesDir, tilesPrefix); err != nil {
		return "", fmt.Errorf("upload tiles: %w", err)
	}

	dziURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", p.Bucket, p.Region, dziKey)
	return dziURL, nil
}

// runVips shells out to vips dzsave. outputBase must be a full path without extension.
func runVips(srcPath, outputBase string) error {
	cmd := exec.Command("vips", "dzsave",
		srcPath,
		outputBase,
		"--tile-size", "256",
		"--overlap", "1",
		"--suffix", ".jpg[Q=85]",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("running: %s", cmd.String())
	return cmd.Run()
}

// downloadFromS3 fetches a key into a local temp file.
// Returns the local path and a cleanup function to delete it.
func downloadFromS3(store *persistencestorage.S3Storage, bucket, key string) (path string, cleanup func(), err error) {
	client := store.Client()
	out, err := client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", nil, fmt.Errorf("GetObject %s: %w", key, err)
	}
	defer out.Body.Close()

	tmp, err := os.CreateTemp("", "tissquest-src-*.png")
	if err != nil {
		return "", nil, err
	}
	if _, err := io.Copy(tmp, out.Body); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", nil, err
	}
	tmp.Close()
	return tmp.Name(), func() { os.Remove(tmp.Name()) }, nil
}

// uploadFile uploads a single local file to S3 under the given key.
func uploadFile(store *persistencestorage.S3Storage, bucket, localPath, s3Key, contentType string) error {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return err
	}
	_, err = store.Upload(s3Key, contentType, data)
	return err
}

// uploadDir walks localDir recursively and uploads every file to S3,
// preserving the relative path structure under s3Prefix.
func uploadDir(store *persistencestorage.S3Storage, bucket, localDir, s3Prefix string) error {
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(localDir, path)
		key := s3Prefix + "/" + filepath.ToSlash(rel)
		ct := mime.TypeByExtension(filepath.Ext(path))
		if ct == "" {
			ct = "application/octet-stream"
		}
		return uploadFile(store, bucket, path, key, ct)
	})
}
