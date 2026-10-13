package blob

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Store interface {
	Ping(context.Context) error
	Put(context.Context, string, []byte, string) error
	Get(context.Context, string) ([]byte, error)
}

type FileStore struct{ Root string }

func (s FileStore) Ping(_ context.Context) error {
	if s.Root == "" {
		return os.ErrInvalid
	}
	return os.MkdirAll(s.Root, 0750)
}
func (s FileStore) Put(_ context.Context, key string, data []byte, _ string) error {
	path, err := s.path(key)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0640)
}
func (s FileStore) Get(_ context.Context, key string) ([]byte, error) {
	path, err := s.path(key)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}
func (s FileStore) path(key string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(key))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", os.ErrPermission
	}
	return filepath.Join(s.Root, clean), nil
}

type S3Store struct {
	client *minio.Client
	bucket string
}

func NewS3(endpoint, accessKey, secretKey, bucket string, secure bool) (*S3Store, error) {
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
	client, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(accessKey, secretKey, ""), Secure: secure})
	if err != nil {
		return nil, err
	}
	return &S3Store{client: client, bucket: bucket}, nil
}
func (s *S3Store) Ping(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !exists {
		return os.ErrNotExist
	}
	return nil
}
func (s *S3Store) Put(ctx context.Context, key string, data []byte, mime string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: mime})
	return err
}
func (s *S3Store) Ready(ctx context.Context, create bool) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if !create {
		return os.ErrNotExist
	}
	return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
}
func (s *S3Store) Get(ctx context.Context, key string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return io.ReadAll(obj)
}
