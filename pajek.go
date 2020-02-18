package main


import (
    "fmt"
    "os"
    "runtime"
    "sync"
    "time"
    "strings"
    "io/ioutil"
    "net/http"
)

var urlQueue chan string
var urlQueueMutex sync.Mutex

var visitedUrls map[string]bool
var visitedUrlsCntr uint64 = 0
var visitedUrlsMutex sync.Mutex

func enqueueLinks(links []string, baseUrl string) {
    for i := range links {
        link := links[i]
        url := getFullUrl(baseUrl, link)

        // Url gets discarded, if queue is full.
        urlQueueMutex.Lock()
        if len(urlQueue) < cap(urlQueue) {
            urlQueue <- url
        }
        urlQueueMutex.Unlock()
    }
}

func crawl() {
    // Construct HTTP client
    client := http.Client{
        Timeout: time.Duration(3 * time.Second),
    }

    for {
        // Pop a URL from queue
        url := <-urlQueue

        // Check if this URL was already visited.
        // If not, add URL to visited map, so future crawls will ignore it.
        visitedUrlsMutex.Lock()
        if visitedUrls[url] == true {
            visitedUrlsMutex.Unlock()
            continue
        } else {
            visitedUrls[url] = true
            visitedUrlsCntr++
        }
        visitedUrlsMutex.Unlock()

        fmt.Println("Fetching:", url)
        resp, err := client.Get(url)
        if err != nil {
            fmt.Println(err)
            continue
        }

        contentType := resp.Header["Content-Type"][0]
        if strings.HasPrefix(contentType, "text/html;") {
            body, err := ioutil.ReadAll(resp.Body)
            if err != nil {
                fmt.Println(err)
                continue
            }

            parseBody(body, url)

            links := findLinks(body)
            enqueueLinks(links, url)
        }

        resp.Body.Close()
    }
}

func main() {
    argsWithoutProg := os.Args[1:]

    if len(argsWithoutProg) == 0 {
        fmt.Println("Too few input arguments. Expected 1 or more.")
        os.Exit(1)
    }

    urlQueue = make(chan string, 5*10^5)
    for i := range argsWithoutProg {
        urlQueue <- argsWithoutProg[i]
    }

    visitedUrls = make(map[string]bool)

    // Spawn workers
    for i := 0; i < runtime.NumCPU(); i++ {
        go crawl()
    }

    select {}
}
