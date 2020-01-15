RSS2HTTP 
===

Pick a random item from an atom rss feed and post it to a given url after running it through a template.

Useful to advertise an article through a slack webhook. 

# How it works

* Checks **cache directory** for already downloaded feed based on hashed full url *(errors at this point are ignored)*
* Downloads file or **throws error** that stops the program
* Saves **file** as a cache *(errors are ignored)*
* Runs an atom **xml parser** that creates feed items
* Picks **one random** item
* Runs a **template** method based on custom string to **create request body**
* Sends a **POST** request to a given url

# Usage

**You have to manually create and clean cache directory**

```
$ build/rss2http --post https://slack.webhook.url

$ cat cmd/rss2http/main.go
...
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
...

```


# Setup

```
$ docker-compose up -d
$ docker-compose exec app bash
root@fab9e94fa30c:/go/src/app# go test ./test
root@fab9e94fa30c:/go/src/app# go build -o build/rss2http ./cmd/rss2http
```

## Notes / howto

### Template variables
```
$ cat pkg/parser/parser.go
...
type ParsedPodcast struct {
	Title, Link, Publicated string
}
...
```

### Add your own parser / etc

At the moment there is no built-in way to support 3rd party, custom parsers.

Build your own:

* Create a downloader based on signatures
```
$ cat pkg/downloader/downloader.go
...
type Downloader = func (url string) ([]byte, error)
type CacheReader = func (url string) ([]byte, error)
type CacheWriter = func (url string, contents []byte) error
type Hasher = func (url string) string
...
func GetCachedOrDownload(url string, reader CacheReader, writer CacheWriter, downloader Downloader) ([]byte, error) {
```
* Add your parser
* Create a new command








 