# Hypert - HTTP API Testing Made Easy
[![build-img]][build-url]
[![pkg-img]][pkg-url]
[![reportcard-img]][reportcard-url]
[![coverage-img]][coverage-url]
[![tag-img]][tag-url]
[![license-img]][license-url]

Hypert is an open-source Go library that simplifies testing of HTTP API clients. It provides a convenient way to record and replay HTTP interactions, making it easy to create reliable and maintainable tests for your API clients.

## Features

- Record and replay HTTP interactions
- Request sanitization to remove sensitive information
- Request validation to ensure the integrity of recorded requests
- Seamless integration with Go's `http.Client`
- Extensible and configurable options

## Getting Started

1. Install Hypert:

```bash
go get github.com/areknoster/hypert
```

2. Use `hypert.TestClient` to create an `http.Client` instance for testing:
```go
func TestMyAPI(t *testing.T) {
	httpClient := hypert.TestClient(t, true) // true to record real requests
	// Use the client to make API requests. 
	// The requests and responses would be stored in ./testdata/TestMyAPI
	myAPI := NewMyAPI(httpClient, os.GetEnv("API_SECRET")) 
	// Make an API request with your adapter.
	// use static arguments, so that validation against recorded requests can happen
	stuff, err := myAPI.GetStuff(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)) 
	if err != nil {
		t.Fatalf("failed to get stuff: %v", err)
	}
    // Assertions on the actual API response
	if stuff.ID != "ID-FROM-RESP" {
		t.Errorf("stuff ")
	}
}
```
After you're done with building and testing your integration, change the mode to replay
```go
func TestMyAPI(t *testing.T) {
    httpClient := hypert.TestClient(t, false) // false to replay stored requests
    // Now client would validate requests against what's stored in ./testdata/TestMyAPI/*.req.http 
    // and load the response from  ./testdata/TestMyAPI/*.resp.http
    myAPI := NewMyAPI(httpClient, os.GetEnv("API_SECRET"))
    // HTTP requests are validated against what was prevously recorded. 
    // This behaviour can be customized using WithRequestValidator option
    stuff, err := myAPI.GetStuff(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)) 
    if err != nil {
        t.Fatalf("failed to get stuff: %v", err)
    }
    // Same assertions that were true on actual API responses should be true for replayed API responses.
	if stuff.ID != "ID-FROM-RESP" {
        t.Errorf("stuff ")
    }
}
```
Now your tests:
- are deterministic
- are fast
- bring the same confidence as integration tests
## Configuration

Hypert provides various options to customize its behavior:

- `WithNamingScheme`: Set the naming scheme for recorded requests
- `WithParentHTTPClient`: Set a custom parent `http.Client`
- `WithRequestSanitizer`: Configure the request sanitizer to remove sensitive information
- `WithRequestValidator`: Set a custom request validator

## Examples

Check out the [examples](examples/) directory for sample usage of Hypert in different scenarios.

## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue on the [GitHub repository](https://github.com/areknoster/hypert). If you'd like to contribute code, please fork the repository and submit a pull request.

## License

Hypert is released under the [MIT License](LICENSE).

---

[build-img]: https://github.com/areknoster/hypert/workflows/build/badge.svg
[build-url]: https://github.com/areknoster/hypert/actions
[pkg-img]: https://pkg.go.dev/badge/areknoster/hypert/
[pkg-url]: https://pkg.go.dev/github.com/areknoster/hypert/
[reportcard-img]: https://goreportcard.com/badge/github.com/areknoster/hypert/
[reportcard-url]: https://goreportcard.com/report/github.com/areknoster/hypert/
[coverage-img]: https://codecov.io/gh/areknoster/hypert//branch/main/graph/badge.svg
[coverage-url]: https://codecov.io/gh/areknoster/hypert/
[license-img]: https://img.shields.io/github/license/areknoster/hypert/
[license-url]: https://github.com/areknoster/hypert/blob/main/LICENSE
[tag-img]: https://img.shields.io/github/v/tag/areknoster/hypert
[tag-url]: https://github.com/areknoster/hypert/tags
