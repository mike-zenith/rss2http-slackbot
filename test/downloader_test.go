package test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"slackbot-rss/pkg/downloader"
)

func createServer(t *testing.T, handler http.Handler) *httptest.Server {
	ts := httptest.NewUnstartedServer(handler)
	ts.Start()
	t.Logf("Created unstarted test server at %v", ts.URL)
	return ts
}

func createHTTPHandler(t *testing.T, respondBody []byte, respondStatus int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server received request: %v", r)
		t.Logf("Sending success response")
		w.WriteHeader(respondStatus)
		fmt.Fprintln(w, respondBody)
	})
}

func createDownloader(s *httptest.Server) downloader.Downloader {
	return downloader.NewHttpAwareDownloader(s.Client())
}

func TestDownloadOK(t *testing.T) {
	expectedResponse := []byte("Hello")

	ts := createServer(t, createHTTPHandler(t, expectedResponse, 200))
	defer ts.Close()

	dl := createDownloader(ts)
	content, err := dl(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(content, expectedResponse) {
		t.Errorf("Expected (%s) response did not match given: %s",  expectedResponse, content)
	}

}

func TestDownloadReturnsError(t *testing.T) {
	expectedStatus := 400
	ts := createServer(t, createHTTPHandler(t, []byte{}, expectedStatus))
	defer ts.Close()

	dl := createDownloader(ts)
	content, err := dl(ts.URL)
	if err == nil {
		t.Fatalf("Expecting error, nothing returned")
	}
	if content != nil {
		t.Fatalf("Should not have returned response")
	}
	if strings.Contains(err.Error(), fmt.Sprintf("%v", expectedStatus)) == false {
		t.Errorf("Error message does not contain status code: %v", err)
	}
}

func createTempDir() string {
	dir, _ := ioutil.TempDir("", "downloader")
	return dir
}

func prepareDummyCachedData(t *testing.T, dir string, url string, hasher downloader.Hasher, contents []byte) {
	filename := hasher(url)
	path := dir + "/" + filename
	ioutil.WriteFile(path, contents, 0666)
	t.Logf("Writing %v to %s", contents, path)
}

func TestGetCachedOrDownloadReadsFromCache(t *testing.T) {
	url := "http://127.0.0.1:8000/g.xml"
	cachedContent := []byte("Okay")

	hasher := downloader.NewUrlToSha1Hasher()
	dir := createTempDir()
	defer  os.RemoveAll(dir)

	prepareDummyCachedData(t, dir, url, hasher, cachedContent)

	contents, err := downloader.GetCachedOrDownload(
		url,
		downloader.NewUrlToHashedFilenameReader(dir, hasher),
		func (url string, contents []byte) error { return nil },
		func (url string) ([]byte, error) { return nil, nil })

	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(contents,cachedContent) == false {
		t.Errorf("Expected cached contents, received: %v", contents)
	}
}

func TestGetCachedOrDownloadFallsToDownloadOnError(t *testing.T) {
	url := "http://127.0.0.1:8000/g.xml"

	expectedDownloadedContent := []byte("This is magical content")

	dl := func (url string) ([]byte, error) { return expectedDownloadedContent, nil }

	contents, err := downloader.GetCachedOrDownload(
		url,
		func (url string) ([]byte, error) { return nil, errors.New("bumbum") },
		func (url string, contents []byte) error { return nil },
		dl)

	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(contents, expectedDownloadedContent) == false {
		t.Errorf("Expected downloaded contents, received: %v", contents)
	}
}

func TestGetCachedOrDownloadStoresDownloaded(t *testing.T) {
	url := "http://127.0.0.1:8000/g.xml"

	hasher := downloader.NewUrlToSha1Hasher()
	dir := createTempDir()
	defer  os.RemoveAll(dir)

	expectedDownloadedContent := []byte("This is magical content")
	dl := func (url string) ([]byte, error) { return expectedDownloadedContent, nil }

	_, err := downloader.GetCachedOrDownload(
		url,
		func (url string) ([]byte, error) { return nil, errors.New("download instead") },
		downloader.NewUrlToHashedFilenameWriter(dir, hasher),
		dl)

	if err != nil {
		t.Fatal(err)
	}

	contents, err := ioutil.ReadFile(dir + "/" + hasher(url))
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(contents, expectedDownloadedContent) == false {
		t.Errorf("Expected downloaded contents, received: %v", contents)
	}
}

func TestGetCachedOrDownloadReturnsDownloadedWhenWritingFails(t *testing.T) {
	url := "http://127.0.0.1:8000/g.xml"

	expectedDownloadedContent := []byte("This is magical content")
	dl := func (url string) ([]byte, error) { return expectedDownloadedContent, nil }
	writeError := errors.New("big woofy write error")

	contents, err := downloader.GetCachedOrDownload(
		url,
		func (url string) ([]byte, error) { return nil, errors.New("download instead") },
		func (url string, c []byte) error { return writeError },
		dl)

	if bytes.Equal(contents, expectedDownloadedContent) == false {
		t.Errorf("Expected downloaded contents, received: %v", contents)
	}
	if err != writeError {
		t.Errorf("Expected to return the write error, got: %v", writeError)
	}
}

func TestGetCachedOrDownloadSkipsCacheWhenUnreadable(t *testing.T) {
	url := "http://127.0.0.1:8000/g.xml"
	downloadedContent := []byte("Okay")

	hasher := downloader.NewUrlToSha1Hasher()

	contents, err := downloader.GetCachedOrDownload(
		url,
		downloader.NewUrlToHashedFilenameReader("./nodirectory", hasher),
		func (url string, contents []byte) error { return nil },
		func (url string) ([]byte, error) { return downloadedContent, nil })

	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(contents,downloadedContent) == false {
		t.Errorf("Expected cached contents, received: %v", contents)
	}
}

