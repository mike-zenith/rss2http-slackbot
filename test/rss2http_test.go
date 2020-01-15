package test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
)

func buildBinary() string {
	if err := os.Chdir(".."); err != nil {
		fmt.Println("Could not change directory")
		os.Exit(1)
	}

	build := exec.Command("go", "build", "-o", "build/rss2http", "./cmd/rss2http")
	out, err := build.CombinedOutput()
	if err != nil {
		fmt.Printf("Could not build binary: %s %v\n", out, err)
		os.Exit(1)
	}
	if err := os.Chdir("test"); err != nil {
		fmt.Println("Could not change directory")
		os.Exit(1)
	}

	return "../build/rss2http"
}

func getServerWithHandler(h func(http.ResponseWriter, *http.Request)) *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(h))
}

func TestRss2Http(t *testing.T) {
	binary := buildBinary()

	t.Run("FAIL --post", func (t *testing.T) {
		ts := getServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		})
		ts.Start()
		defer ts.Close()

		cmd := exec.Command(binary, "--post", ts.URL)
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Errorf("Expected error. Got: \n%s\n", out)
		}
	})

	t.Run("OK --post", func (t *testing.T) {
		ts := getServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(400)
				return
			}
			w.WriteHeader(200)
		})
		ts.Start()
		defer ts.Close()

		cmd := exec.Command(binary, "--post", ts.URL)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Output running command: \n%s\n%s\n", out, err)
		}
	})

	t.Run("OK --tpl", func (t *testing.T) {
		tpl := []byte(`{"dummy": "template"}`)
		ts := getServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				return
			}
			if !bytes.Equal(tpl, body) {
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(200)
		})
		ts.Start()
		defer ts.Close()

		cmd := exec.Command(binary, "--post", ts.URL, "--tpl", string(tpl))
		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("OK --cache", func (t *testing.T) {
		dir, _ := ioutil.TempDir("", "rss2httpcli_cache")
		defer os.RemoveAll(dir)
		cmd := exec.Command(binary, "--cache", dir)
		cmd.Run()
		files, e := ioutil.ReadDir(dir)
		if e != nil {
			t.Errorf("Could not read dir: %v", dir)
		}
		if len(files) != 1 {
			t.Errorf("Unexpected cache files found: %v", files)
		}
	})

}



