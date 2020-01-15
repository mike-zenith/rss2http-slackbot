package parser

import (
	"github.com/mmcdole/gofeed"
	"math/rand"
	"time"
)

type ParsedPodcast struct {
	Title, Link, Publicated string
}

func RSSPodcastParser(in []byte) ([]ParsedPodcast, error) {
	feedParser := gofeed.NewParser()
	parsed, err := feedParser.ParseString(string(in))
	if err != nil {
		return nil, err
	}
	var parsedPodcasts = make([]ParsedPodcast, len(parsed.Items))
	for i, item := range parsed.Items {
		parsedPodcasts[i] = ParsedPodcast{Title: item.Title, Link: item.Link, Publicated: item.Published}
	}

	return parsedPodcasts, nil
}

type PickOne = func (from []ParsedPodcast) ParsedPodcast

func PickRandom(from []ParsedPodcast) ParsedPodcast {
	rand.Seed(time.Now().Unix())
	n := rand.Intn(len(from))
	return from[n]
}