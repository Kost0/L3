package minIO

import (
	"context"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/wb-go/wbf/zlog"
)

func InitMinio() *minio.Client {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatal(("Error initialization minio client"))
	}

	bucketName := "images"

	makeBucket(minioClient, bucketName)

	return minioClient
}

func makeBucket(minioClient *minio.Client, bucketName string) {
	ctx := context.Background()
	location := "us-east-1"

	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err == nil && exists {
		zlog.Logger.Info().Msg("Bucket exists")
		return
	}

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region: location,
	})
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to create bucket")
	}

	zlog.Logger.Info().Msg("Bucket created")
}
