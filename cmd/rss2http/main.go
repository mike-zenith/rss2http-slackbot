package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"slackbot-rss/pkg/downloader"
	"slackbot-rss/pkg/httprequest"
	"slackbot-rss/pkg/parser"
	"time"
)

type InputArguments struct {
	rss string
	post string
	tpl string
	cache string
}

func getInputArguments() InputArguments {
	currDir, _ :=  os.Getwd()
	defaults := InputArguments{
		rss: "https://feed.pippa.io/public/shows/5aeff6d96eb47cc259946df2",
		post: "https://127.0.0.1:9000/test",
		tpl: `{
		"channel": "#kerekasztal", 
		"username": "webhookbot", 
		"text": "Hallgasd meg most a <a href='{{ .Link }}'>{{ .Title }}</a> adást, ami naplivágot látott {{ .Publicated }}", 
		"icon_emoji": ":ghost:"
		}`,
		cache: path.Join(currDir, "tmp"),
	}
	input := InputArguments{}
	flag.StringVar(&input.rss, "rss", defaults.rss, "Http url of the rss feed")
	flag.StringVar(&input.post, "post", defaults.post, "Http url to post")
	flag.StringVar(&input.tpl, "tpl", defaults.tpl, "Golang template string that will be sent to post")
	flag.StringVar(&input.cache, "cache", defaults.cache, "Cache directory")
	flag.Parse()
	return input
}

func rss2http(input InputArguments, w io.Writer) error {
	hasher := downloader.NewUrlToSha1Hasher()

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
	}
	httpClient := &http.Client{Transport: tr}

	xmlResponse, err := downloader.GetCachedOrDownload(
		input.rss,
		downloader.NewUrlToHashedFilenameReader(input.cache, hasher),
		downloader.NewUrlToHashedFilenameWriter(input.cache, hasher),
		downloader.NewHttpAwareDownloader(httpClient))

	if err != nil && len(xmlResponse) == 0 {
		fmt.Fprintf(w, "Error while downloading/saving: %v", err)
		return err
	}
	if len(xmlResponse) == 0 {
		fmt.Fprintf(w,"Empty response returned: %v\n", xmlResponse)
		return errors.New("empty response")
	}

	podcasts, err := parser.RSSPodcastParser(xmlResponse)
	if err != nil {
		fmt.Fprintf(w,"Error while parsing xml: %v\n", err)
		return err
	}

	client := httprequest.NewPostTemplatedRequestClient(
		httpClient,
		input.post,
		httprequest.NewTemplatedItemToReadableStream(input.tpl))

	_, err = client(parser.PickRandom(podcasts))
	if err != nil {
		fmt.Fprintf(w,"Error while posting data to given url: %v\n", err)
		return err
	}
	return nil
}

func main() {
	input := getInputArguments()
	e := rss2http(input, os.Stderr)
	if e != nil {
		os.Exit(1)
	}
}




