package storage

type ImageStorage interface {
	Upload(filename string, contentType string, data []byte) (url string, err error)
}
