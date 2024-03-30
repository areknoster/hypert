package hypert

import (
	"os"
	"sync"
	"testing"
)

func TestNewSequentialNamingScheme(t *testing.T) {
	// Test creating a new SequentialNamingScheme
	dir := t.Name() + "testdir"
	_, err := NewSequentialNamingScheme(dir)
	if err != nil {
		t.Fatalf("Error creating SequentialNamingScheme: %v", err)
	}

	// Check if the directory was created
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Directory %s was not created", dir)
	}

	// Clean up the test directory
	err = os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("Error removing test directory: %v", err)
	}
}

func TestSequentialNamingScheme_FileNames(t *testing.T) {
	// Create a new SequentialNamingScheme for testing
	dir := "testdir"
	scheme, err := NewSequentialNamingScheme(dir)
	if err != nil {
		t.Fatalf("Error creating SequentialNamingScheme: %v", err)
	}

	// Clean up the test directory
	defer os.RemoveAll(dir)

	type reqRespPair struct {
		req  string
		resp string
	}
	const reqCount = 3
	reqRespPairs := make(chan reqRespPair, reqCount)

	// Create a bunch of request/response pairs
	var wg sync.WaitGroup
	for i := 0; i < reqCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, resp := scheme.FileNames(RequestData{})

			reqRespPairs <- reqRespPair{req, resp}
		}()
	}
	wg.Wait()
	close(reqRespPairs)
	expectedPairs := map[reqRespPair]struct{}{
		{dir + "/0.req.http", dir + "/0.resp.http"}: {},
		{dir + "/1.req.http", dir + "/1.resp.http"}: {},
		{dir + "/2.req.http", dir + "/2.resp.http"}: {},
	}
	for pair := range reqRespPairs {
		if _, ok := expectedPairs[pair]; !ok {
			t.Errorf("Unexpected request/response pair: %v", pair)
		}
	}
}
