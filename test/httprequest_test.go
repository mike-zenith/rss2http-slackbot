package test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slackbot-rss/pkg/httprequest"
	"slackbot-rss/pkg/parser"
	"testing"
)

func TestNewPostEncodedRequestThroughClientOK(t *testing.T) {
	item := parser.ParsedPodcast{
		Title: "Pluginek",
		Link: "https://shows.acast.com/5aeff6d96eb47cc259946df2/pluginek",
		Publicated: "Mon, 13 Jan 2020 07:00:06 GMT",
	}
	expectedJsonResponse := map[string]string {
		"channel": "#kerekasztal",
		"username": "webhookbot",
		"text": "Hallgasd meg most a <a href='https://shows.acast.com/5aeff6d96eb47cc259946df2/pluginek'>Pluginek</a> adást, ami naplivágot látott Mon, 13 Jan 2020 07:00:06 GMT",
		"icon_emoji": ":ghost:",
	}
	baseTemplate := `{
		"channel": "#kerekasztal", 
		"username": "webhookbot", 
		"text": "Hallgasd meg most a <a href='{{ .Link }}'>{{ .Title }}</a> adást, ami naplivágot látott {{ .Publicated }}", 
		"icon_emoji": ":ghost:"
	}`

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response, _ := ioutil.ReadAll(r.Body)
		w.Write(response)
	}))

	ts.Start()

	client := httprequest.NewPostTemplatedRequestClient(
		ts.Client(),
		ts.URL,
		httprequest.NewTemplatedItemToReadableStream(baseTemplate))

	rawResponse, err := client(item)
	if err != nil {
		t.Fatal(err)
	}

	var response map[string]string
	if err := json.Unmarshal(rawResponse, &response); err != nil {
		t.Fatalf("Error parsing %s \n\r Error receivedr: %v", rawResponse, err)
	}

	if reflect.DeepEqual(response, expectedJsonResponse) == false {
		t.Errorf("Failed asserting that expected response returned. Expected:\n\r%v\n\rGot:\n\r%v\n\r", expectedJsonResponse, response)
	}
}