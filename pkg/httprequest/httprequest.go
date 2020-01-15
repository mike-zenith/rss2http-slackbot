package httprequest

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
)

type MakeRequestWithPickedItem = func(item interface{}) (response []byte, e error)
type ItemToReadableStream = func(item interface{}) io.Reader

func NewPostTemplatedRequestClient(client *http.Client, url string, itemToPostData ItemToReadableStream) MakeRequestWithPickedItem {
	return func(item interface{}) (response []byte, err error) {
		res, err := client.Post(url, "application/json", itemToPostData(item))
		if err != nil {
			return nil, err
		}
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("status error: %v", res.StatusCode)
		}

		return ioutil.ReadAll(res.Body)
	}
}

func NewTemplatedItemToReadableStream(baseTemplate string) ItemToReadableStream {
	tpl, err := template.New("templatedItems").Parse(baseTemplate)
	if err != nil {
		panic(err)
	}
	return func (item interface{}) io.Reader {
		buffer := bytes.NewBuffer([]byte{})
		tpl.Execute(buffer, item)
		return buffer
	}
}
