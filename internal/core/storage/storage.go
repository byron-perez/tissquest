package storage

import (
	"context"
	"io"
)

type MediaStore interface {
	Upload(ctx context.Context, key, contentType string, r io.Reader, size int64) (url string, err error)
	Delete(ctx context.Context, key string) error
	DeletePrefix(ctx context.Context, prefix string) error
	PublicURL(key string) string
}
