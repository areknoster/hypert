package hypert

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

	const reqCount = 3
	reqRespPairs := make(chan [2]string, reqCount)

	// Create a bunch of request/response pairs
	var wg sync.WaitGroup
	for i := 0; i < reqCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, resp := scheme.FileNames(RequestData{})

			reqRespPairs <- [2]string{req, resp}
		}()
	}
	wg.Wait()
	close(reqRespPairs)
	expectedPairs := map[[2]string]struct{}{
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

func TestSequentialNamingScheme_ConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	scheme, err := NewSequentialNamingScheme(dir)
	if err != nil {
		t.Fatalf("Error creating SequentialNamingScheme: %v", err)
	}

	const numGoroutines = 100
	fileNames := make(chan string, numGoroutines*2)
	var wg sync.WaitGroup

	// Test concurrent access
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, resp := scheme.FileNames(RequestData{})
			fileNames <- req
			fileNames <- resp
		}()
	}

	wg.Wait()
	close(fileNames)

	// Check that all filenames are unique
	seen := make(map[string]bool)
	for fileName := range fileNames {
		if seen[fileName] {
			t.Errorf("Duplicate filename generated: %s", fileName)
		}
		seen[fileName] = true
	}

	if len(seen) != numGoroutines*2 {
		t.Errorf("Expected %d unique filenames, got %d", numGoroutines*2, len(seen))
	}
}

func TestNewPathBasedNamingScheme(t *testing.T) {
	dir := t.TempDir()
	scheme, err := NewPathBasedNamingScheme(dir)
	if err != nil {
		t.Fatalf("Error creating PathBasedNamingScheme: %v", err)
	}

	if scheme.dir != dir {
		t.Errorf("Expected dir %s, got %s", dir, scheme.dir)
	}

	// Check if the directory was created
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Directory %s was not created", dir)
	}
}

func TestPathBasedNamingScheme_FileNames(t *testing.T) {
	dir := t.TempDir()
	scheme, err := NewPathBasedNamingScheme(dir)
	if err != nil {
		t.Fatalf("Error creating PathBasedNamingScheme: %v", err)
	}

	tests := []struct {
		name        string
		url         string
		expectSame  bool
		description string
	}{
		{
			name:        "same_url_different_calls",
			url:         "https://example.com/api/users",
			expectSame:  false,
			description: "Same URL should generate different filenames with counter",
		},
		{
			name:        "different_urls",
			url:         "https://example.com/api/posts",
			expectSame:  false,
			description: "Different URLs should generate different filenames",
		},
		{
			name:        "url_with_query_params",
			url:         "https://example.com/api/users?page=1&limit=10",
			expectSame:  false,
			description: "URLs with query params should be considered different",
		},
		{
			name:        "url_with_different_query_params",
			url:         "https://example.com/api/users?page=2&limit=10",
			expectSame:  false,
			description: "URLs with different query params should be different",
		},
	}

	generatedFiles := make(map[string][]string)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("Error parsing URL: %v", err)
			}

			data := RequestData{URL: u}
			req, resp := scheme.FileNames(data)

			// Check that files have correct extensions
			if !strings.HasSuffix(req, ".req.http") {
				t.Errorf("Request file should end with .req.http, got: %s", req)
			}
			if !strings.HasSuffix(resp, ".resp.http") {
				t.Errorf("Response file should end with .resp.http, got: %s", resp)
			}

			// Check that files are in the correct directory
			if !strings.HasPrefix(req, dir) {
				t.Errorf("Request file should be in directory %s, got: %s", dir, req)
			}
			if !strings.HasPrefix(resp, dir) {
				t.Errorf("Response file should be in directory %s, got: %s", dir, resp)
			}

			generatedFiles[tt.url] = append(generatedFiles[tt.url], req, resp)
		})
	}

	// Test that calling the same URL multiple times generates different filenames
	u, err := url.Parse("https://example.com/api/unique-test-url")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}
	data := RequestData{URL: u}

	req1, resp1 := scheme.FileNames(data)
	req2, resp2 := scheme.FileNames(data)

	if req1 == req2 || resp1 == resp2 {
		t.Error("Same URL should generate different filenames on subsequent calls")
	}

	// The second call should have a counter suffix (format: hash-1)
	baseName1 := filepath.Base(req1)
	baseName2 := filepath.Base(req2)
	hashPart1 := strings.TrimSuffix(baseName1, ".req.http")
	hashPart2 := strings.TrimSuffix(baseName2, ".req.http")

	// First call should not have counter, second should have "-1"
	if strings.Contains(hashPart1, "-") {
		t.Error("First call should not have counter suffix")
	}
	if !strings.HasSuffix(hashPart2, "-1") {
		t.Errorf("Second call should have '-1' suffix, got: %s", hashPart2)
	}
}

func TestPathBasedNamingScheme_ConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	scheme, err := NewPathBasedNamingScheme(dir)
	if err != nil {
		t.Fatalf("Error creating PathBasedNamingScheme: %v", err)
	}

	const numGoroutines = 50
	results := make(chan []string, numGoroutines)
	var wg sync.WaitGroup

	// Test concurrent access with same URL
	u, err := url.Parse("https://example.com/test")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}
	data := RequestData{URL: u}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, resp := scheme.FileNames(data)
			results <- []string{req, resp}
		}()
	}

	wg.Wait()
	close(results)

	// Check that all generated filenames are unique
	seen := make(map[string]bool)
	for result := range results {
		for _, fileName := range result {
			if seen[fileName] {
				t.Errorf("Duplicate filename generated: %s", fileName)
			}
			seen[fileName] = true
		}
	}
}

func TestNewContentHashNamingScheme(t *testing.T) {
	dir := t.TempDir()
	scheme, err := NewContentHashNamingScheme(dir)
	if err != nil {
		t.Fatalf("Error creating ContentHashNamingScheme: %v", err)
	}

	if scheme.dir != dir {
		t.Errorf("Expected dir %s, got %s", dir, scheme.dir)
	}

	// Check if the directory was created
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Directory %s was not created", dir)
	}
}

func TestContentHashNamingScheme_FileNames(t *testing.T) {
	dir := t.TempDir()
	scheme, err := NewContentHashNamingScheme(dir)
	if err != nil {
		t.Fatalf("Error creating ContentHashNamingScheme: %v", err)
	}

	tests := []struct {
		name        string
		url         string
		body        []byte
		expectSame  bool
		description string
	}{
		{
			name:        "same_path_same_content",
			url:         "https://example.com/api/users",
			body:        []byte(`{"name": "John"}`),
			expectSame:  true,
			description: "Same path and content should generate same filenames",
		},
		{
			name:        "same_path_different_content",
			url:         "https://example.com/api/users",
			body:        []byte(`{"name": "Jane"}`),
			expectSame:  false,
			description: "Same path but different content should generate different filenames",
		},
		{
			name:        "different_path_same_content",
			url:         "https://example.com/api/posts",
			body:        []byte(`{"name": "John"}`),
			expectSame:  false,
			description: "Different path with same content should generate different filenames",
		},
		{
			name:        "empty_content",
			url:         "https://example.com/api/empty",
			body:        nil,
			expectSame:  false,
			description: "Empty content should be handled correctly",
		},
		{
			name:        "root_path",
			url:         "https://example.com/",
			body:        []byte(`test`),
			expectSame:  false,
			description: "Root path should be handled correctly",
		},
	}

	generatedFiles := make(map[string][]string)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("Error parsing URL: %v", err)
			}

			data := RequestData{
				URL:       u,
				BodyBytes: tt.body,
			}
			req, resp := scheme.FileNames(data)

			// Check that files have correct extensions
			if !strings.HasSuffix(req, ".req.http") {
				t.Errorf("Request file should end with .req.http, got: %s", req)
			}
			if !strings.HasSuffix(resp, ".resp.http") {
				t.Errorf("Response file should end with .resp.http, got: %s", resp)
			}

			// Check that files are in the correct directory
			if !strings.HasPrefix(req, dir) {
				t.Errorf("Request file should be in directory %s, got: %s", dir, req)
			}
			if !strings.HasPrefix(resp, dir) {
				t.Errorf("Response file should be in directory %s, got: %s", dir, resp)
			}

			// Check that filenames contain valid hash (16 characters)
			baseName := filepath.Base(req)
			hashPart := strings.TrimSuffix(baseName, ".req.http")
			if len(hashPart) != 16 {
				t.Errorf("Hash part should be 16 characters, got %d: %s", len(hashPart), hashPart)
			}

			generatedFiles[tt.name] = []string{req, resp}
		})
	}

	// Test that same path and content generate same filenames
	u, err := url.Parse("https://example.com/test")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}
	data := RequestData{
		URL:       u,
		BodyBytes: []byte("test content"),
	}

	req1, resp1 := scheme.FileNames(data)
	req2, resp2 := scheme.FileNames(data)

	if req1 != req2 || resp1 != resp2 {
		t.Error("Same path and content should generate identical filenames")
	}
}

func TestContentHashNamingScheme_EdgeCases(t *testing.T) {
	dir := t.TempDir()
	scheme, err := NewContentHashNamingScheme(dir)
	if err != nil {
		t.Fatalf("Error creating ContentHashNamingScheme: %v", err)
	}

	tests := []struct {
		name string
		data RequestData
	}{
		{
			name: "nil_url",
			data: RequestData{
				URL:       nil,
				BodyBytes: []byte("test"),
			},
		},
		{
			name: "empty_path_and_body",
			data: RequestData{
				URL:       &url.URL{Path: ""},
				BodyBytes: nil,
			},
		},
		{
			name: "large_body",
			data: RequestData{
				URL:       &url.URL{Path: "/test"},
				BodyBytes: make([]byte, 10000), // Large body
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic and should return valid filenames
			req, resp := scheme.FileNames(tt.data)

			if req == "" || resp == "" {
				t.Error("FileNames should not return empty strings")
			}

			if !strings.HasSuffix(req, ".req.http") || !strings.HasSuffix(resp, ".resp.http") {
				t.Error("FileNames should return files with correct extensions")
			}
		})
	}
}

// Benchmark tests
func BenchmarkSequentialNamingScheme(b *testing.B) {
	dir := b.TempDir()
	scheme, err := NewSequentialNamingScheme(dir)
	if err != nil {
		b.Fatalf("Error creating SequentialNamingScheme: %v", err)
	}

	data := RequestData{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scheme.FileNames(data)
	}
}

func BenchmarkPathBasedNamingScheme(b *testing.B) {
	dir := b.TempDir()
	scheme, err := NewPathBasedNamingScheme(dir)
	if err != nil {
		b.Fatalf("Error creating PathBasedNamingScheme: %v", err)
	}

	u, err := url.Parse("https://example.com/api/test")
	if err != nil {
		b.Fatalf("failed to parse URL: %v", err)
	}
	data := RequestData{URL: u}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scheme.FileNames(data)
	}
}

func BenchmarkContentHashNamingScheme(b *testing.B) {
	dir := b.TempDir()
	scheme, err := NewContentHashNamingScheme(dir)
	if err != nil {
		b.Fatalf("Error creating ContentHashNamingScheme: %v", err)
	}

	u, err := url.Parse("https://example.com/api/test")
	if err != nil {
		b.Fatalf("failed to parse URL: %v", err)
	}
	data := RequestData{
		URL:       u,
		BodyBytes: []byte(`{"test": "data"}`),
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scheme.FileNames(data)
	}
}
