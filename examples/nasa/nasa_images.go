package nasa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

type ImagesLister struct {
	client *http.Client
	apiKey string
}

func NewImagesLister(c *http.Client, apiKey string) *ImagesLister {
	return &ImagesLister{
		client: c,
		apiKey: apiKey,
	}
}

type listImagesResponse []struct {
	URL string `json:"url"`
}

func (il *ImagesLister) ListNASAImages(ctx context.Context, from, to time.Time) ([]string, error) {
	q := url.Values{
		"api_key":    {il.apiKey},
		"start_date": {from.Format("2006-01-02")},
		"end_date":   {to.Format("2006-01-02")},
	}
	u := &url.URL{
		Scheme:   "https",
		Host:     "api.nasa.gov",
		Path:     "/planetary/apod",
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := il.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var images listImagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&images); err != nil {
		return nil, err
	}

	var urls []string
	for _, img := range images {
		urls = append(urls, img.URL)
	}
	return urls, nil
}
