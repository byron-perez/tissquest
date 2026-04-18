package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
)

// s3Client is initialized once globally to be reused across Lambda invocations.
var s3Client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}
	s3Client = s3.NewFromConfig(cfg)
}

// sizes maps variant names to their target dimensions.
// Names must match the ImageSize constants in the Go app (internal/core/slide).
var sizes = map[string]struct{ W, H int }{
	"thumb":   {W: 300, H: 200},
	"preview": {W: 800, H: 600},
}

func handler(ctx context.Context, s3Event events.S3Event) error {
	// APP_BASE_URL is set in the Lambda environment, e.g. "https://yourapp.com"
	appBaseURL := strings.TrimRight(os.Getenv("APP_BASE_URL"), "/")

	for _, record := range s3Event.Records {
		bucket := record.S3.Bucket.Name
		region := record.AWSRegion

		// S3 keys are URL-encoded in the event payload.
		key, err := url.QueryUnescape(record.S3.Object.Key)
		if err != nil {
			return fmt.Errorf("failed to unescape key %s: %v", record.S3.Object.Key, err)
		}

		// Recursive protection — only process files uploaded to slides/original/.
		// The resized outputs go to slides/thumb/ and slides/preview/, so they
		// won't re-trigger this handler.
		if !strings.HasPrefix(key, "slides/original/") {
			log.Printf("Skipping %s — not in slides/original/", key)
			continue
		}

		filename := filepath.Base(key)
		slideID := strings.TrimSuffix(filename, filepath.Ext(filename))
		log.Printf("Processing slide %s from bucket %s", slideID, bucket)

		// Download the original image.
		resp, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return fmt.Errorf("failed to download %s: %v", key, err)
		}

		img, err := imaging.Decode(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to decode image %s: %v", key, err)
		}

		// Generate each size variant, upload it, then notify the app.
		for sizeName, dims := range sizes {
			resized := imaging.Fit(img, dims.W, dims.H, imaging.Lanczos)

			buf := new(bytes.Buffer)
			if err := imaging.Encode(buf, resized, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
				return fmt.Errorf("failed to encode %s variant: %v", sizeName, err)
			}

			destKey := fmt.Sprintf("slides/%s/%s.jpg", sizeName, slideID)

			_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
				Bucket:      aws.String(bucket),
				Key:         aws.String(destKey),
				Body:        bytes.NewReader(buf.Bytes()),
				ContentType: aws.String("image/jpeg"),
			})
			if err != nil {
				return fmt.Errorf("failed to upload %s: %v", destKey, err)
			}
			log.Printf("Uploaded variant: %s", destKey)

			// Only notify the app after the S3 upload succeeded.
			// If this call fails we log and continue — the app will fall back
			// to the original image gracefully until the variant is registered.
			if appBaseURL != "" {
				variantURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, destKey)
				if err := notifyApp(appBaseURL, slideID, sizeName, variantURL); err != nil {
					log.Printf("Warning: failed to notify app for %s/%s: %v", slideID, sizeName, err)
				}
			} else {
				log.Printf("APP_BASE_URL not set — skipping callback for %s/%s", slideID, sizeName)
			}
		}
	}

	return nil
}

// notifyApp calls PATCH /slides/{id}/images/{size} on the Go app so it can
// record the variant URL in the database.
func notifyApp(baseURL, slideID, sizeName, variantURL string) error {
	endpoint := fmt.Sprintf("%s/slides/%s/images/%s", baseURL, slideID, sizeName)
	body := fmt.Sprintf(`{"url":%q}`, variantURL)

	req, err := http.NewRequest(http.MethodPatch, endpoint, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("app returned unexpected status %d", resp.StatusCode)
	}

	log.Printf("App notified: slide %s size %s → %s", slideID, sizeName, variantURL)
	return nil
}

func main() {
	lambda.Start(handler)
}
