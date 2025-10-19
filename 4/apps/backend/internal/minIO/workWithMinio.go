package minIO

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

func UploadFileFromReader(minioClient *minio.Client, bucketName, objectName string, reader io.Reader, size int64, contentType string) error {
	ctx := context.Background()

	_, err := minioClient.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return err
	}

	return nil
}

func GetPhoto(minioClient *minio.Client, bucketName, objectName string) (*minio.Object, error) {
	ctx := context.Background()

	object, err := minioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return object, nil
}

func RemovePhoto(minioClient *minio.Client, bucketName, objectName string) error {
	ctx := context.Background()

	err := minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}
