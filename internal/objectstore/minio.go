package objectstore

import (
	"log"
	"time"

	"github.com/burntcarrot/wasmninja/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func ConnectMinio(cfg *config.MinioConfig) (*minio.Client, error) {
	var (
		client *minio.Client
		err    error
	)

	// Retry parameters
	retryAttempts := 5
	retryInterval := time.Second

	for i := 0; i < retryAttempts; i++ {
		client, err = minio.New(cfg.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: false,
		})

		if err == nil {
			// Connection successful, break out of the retry loop
			break
		}

		log.Printf("Failed to connect to Minio. Retrying in %s...", retryInterval)

		time.Sleep(retryInterval)
		retryInterval *= 2 // Exponential backoff
	}

	if err != nil {
		return nil, err
	}

	return client, nil
}
