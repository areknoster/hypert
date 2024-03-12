package nasa

import (
	"context"
	"github.com/areknoster/htttest"
	"testing"
	"time"
)

func TestImagesLister_ListNASAImages(t *testing.T) {
	httpClient := htttest.NewDefaultTestClient(t)
	lister := NewImagesLister(httpClient, "DEMO_KEY")
	d1 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := d1.Add(3 * 24 * time.Hour)
	images, err := lister.ListNASAImages(context.Background(), d1, d2)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 4 {
		t.Fatalf("expected four images, got %d", len(images))
	}
}
