package test

import (
	"reflect"
	"testing"
	"io/ioutil"
	"slackbot-rss/pkg/parser"
)

func getDummyRSS() []byte {
	content, err := ioutil.ReadFile("./rss.xml")
	if err != nil {
		panic(err)
	}
	return content
}

func TestRRSParserReturnsAllParsedContent(t *testing.T) {
	rss := getDummyRSS()
	podcasts, err := parser.RSSPodcastParser(rss)

	expectedResult := []parser.ParsedPodcast{
		parser.ParsedPodcast{
			Title: "Pluginek",
			Link: "https://shows.acast.com/5aeff6d96eb47cc259946df2/pluginek",
			Publicated: "Mon, 13 Jan 2020 07:00:06 GMT",
		},
		parser.ParsedPodcast{
			Title: "Egyetemes oroszlánbajusz",
			Link: "https://shows.acast.com/5aeff6d96eb47cc259946df2/egyetemes-oroszlanbajusz",
			Publicated: "Mon, 06 Jan 2020 07:12:41 GMT",
		},
		parser.ParsedPodcast{
			Title: "Svájc",
			Link: "https://shows.acast.com/5aeff6d96eb47cc259946df2/svajc",
			Publicated: "Mon, 16 Dec 2019 07:00:57 GMT",
		},
	}

	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(podcasts, expectedResult) == false {
		t.Errorf("Did not return expected result. Got: %s", podcasts)
	}
}




