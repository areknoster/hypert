package awss3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"log"
	"time"
)

type OldestObjectDownloader struct {
	s3Client *s3.Client
}

func NewOldestObjectDownloader(s3Client *s3.Client) *OldestObjectDownloader {
	return &OldestObjectDownloader{s3Client: s3Client}
}

func (d *OldestObjectDownloader) DownloadOldestObject(ctx context.Context, bucket string) ([]byte, error) {
	var continuationToken *string
	oldestKey := ""
	oldestTime := time.Now()
	for {
		objects, err := d.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            &bucket,
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list objects in bucket %s: %w", bucket, err)
		}
		for _, object := range objects.Contents {
			if object.LastModified.Before(oldestTime) {
				oldestTime = *object.LastModified
				oldestKey = *object.Key
			}
		}
		if !*objects.IsTruncated {
			break
		}
		continuationToken = objects.NextContinuationToken
	}
	if oldestKey == "" {
		return nil, nil
	}

	getObjectOutput, err := d.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &oldestKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s from bucket %s: %w", oldestKey, bucket, err)
	}
	readBytes, err := io.ReadAll(getObjectOutput.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object %s from bucket %s: %w", oldestKey, bucket, err)
	}
	err = getObjectOutput.Body.Close()
	if err != nil {
		log.Printf("failed to close object %s from bucket %s: %v", oldestKey, bucket, err)
	}
	return readBytes, nil
}
