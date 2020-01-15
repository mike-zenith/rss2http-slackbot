package downloader

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
)

type Downloader = func (url string) ([]byte, error)
type CacheReader = func (url string) ([]byte, error)
type CacheWriter = func (url string, contents []byte) error
type Hasher = func (url string) string

func NewUrlToSha1Hasher() Hasher {
	return func (url string) string {
		s := sha1.New()
		s.Write([]byte(url))
		return hex.EncodeToString(s.Sum(nil))
	}
}

func NewUrlToHashedFilenameReader(dir string, hasher Hasher) CacheReader {
	return func (url string) ([]byte, error) {
		filename := hasher(url)
		return ioutil.ReadFile(path.Join(dir, filename))
	}
}

func NewUrlToHashedFilenameWriter(dir string, hasher Hasher) CacheWriter {
	return func (url string, contents []byte) error {
		filename := hasher(url)
		return ioutil.WriteFile(path.Join(dir, filename), contents, 0666)
	}
}

func GetCachedOrDownload(url string, reader CacheReader, writer CacheWriter, downloader Downloader) ([]byte, error) {
	cached, err := reader(url)
	if err == nil {
		return cached, nil
	}
	contents, err := downloader(url)
	if err != nil {
		return nil, err
	}
	if err := writer(url, contents); err != nil {
		return contents, err
	}
	return contents, nil
}

func NewHttpAwareDownloader(client *http.Client) Downloader {
	return func(url string) (bytes []byte, err error) {
		return Download(client, url)
	}
}

func Download(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}
