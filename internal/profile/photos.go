// Package profile validates profile image uploads before storage.
package profile

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const MaxPhotoBytes = 5 << 20

var allowedMIMEs = map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}

func ReadUpload(fileHeader *multipart.FileHeader) ([]byte, string, string, error) {
	if fileHeader == nil {
		return nil, "", "", errors.New("a profile image file is required")
	}
	if fileHeader.Size > MaxPhotoBytes {
		return nil, "", "", errors.New("profile images must be 5 MB or smaller")
	}
	source, err := fileHeader.Open()
	if err != nil {
		return nil, "", "", err
	}
	defer source.Close()
	raw, err := io.ReadAll(io.LimitReader(source, MaxPhotoBytes+1))
	if err != nil {
		return nil, "", "", err
	}
	if int64(len(raw)) > MaxPhotoBytes {
		return nil, "", "", errors.New("profile images must be 5 MB or smaller")
	}
	mimeType := http.DetectContentType(raw)
	extension, allowed := allowedMIMEs[mimeType]
	if !allowed {
		return nil, "", "", errors.New("profile images must be JPEG, PNG, or WebP")
	}
	return raw, extension, mimeType, nil
}
func SaveUpload(fileHeader *multipart.FileHeader, directory string) (string, error) {
	raw, extension, _, err := ReadUpload(fileHeader)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(directory, "profiles"), 0750); err != nil {
		return "", err
	}
	name := uuid.NewString() + extension
	if err := os.WriteFile(filepath.Join(directory, "profiles", name), raw, 0640); err != nil {
		return "", err
	}
	return "/uploads/profiles/" + name, nil
}
func LocalPath(publicURL, directory string) (string, bool) {
	const prefix = "/uploads/profiles/"
	if !strings.HasPrefix(publicURL, prefix) {
		return "", false
	}
	name := strings.TrimPrefix(publicURL, prefix)
	if filepath.Base(name) != name {
		return "", false
	}
	return filepath.Join(directory, name), true
}
func ValidateEmbeddedPhoto(content []byte) error {
	mimeType := http.DetectContentType(content)
	if _, ok := allowedMIMEs[mimeType]; !ok {
		return fmt.Errorf("unsupported profile image")
	}
	return nil
}
