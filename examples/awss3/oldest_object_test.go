package awss3

import (
	"bytes"
	"context"
	"github.com/areknoster/hypert"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"testing"
)

func TestOldestObjectDownloader(t *testing.T) {
	downloader := setupDownloader(t, false)
	ctx := context.Background()
	gotBytes, err := downloader.DownloadOldestObject(ctx, "hypert-example-download-oldest-test-bucket")
	if err != nil {
		t.Fatal(err)
	}
	const expectedContent = "oldest\n"
	if !bytes.Equal(gotBytes, []byte(expectedContent)) {
		t.Fatalf("expected '%s', got '%s'", expectedContent, string(gotBytes))
	}
}

func setupDownloader(t *testing.T, recordModeOn bool) *OldestObjectDownloader {
	hypertClient := hypert.TestClient(
		t,
		recordModeOn,
		hypert.WithRequestSanitizer(hypert.HeadersSanitizer("Authorization", "X-Amz-Security-Token")),
		hypert.WithRequestValidator(
			hypert.ComposedRequestValidator(
				hypert.PathValidator(),
				hypert.MethodValidator(),
				hypert.QueryParamsValidator(),
			),
		),
	)

	var s3Client *s3.Client
	if !recordModeOn {
		s3Client = s3.New(s3.Options{
			Region:     "us-east-1",
			HTTPClient: hypertClient,
		})
	} else {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			t.Fatal(err)
		}
		s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.HTTPClient = hypertClient
			o.Region = "us-east-1"
		})
	}

	return NewOldestObjectDownloader(s3Client)

}
