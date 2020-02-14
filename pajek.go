package main


import (
    "fmt"
    "os"
    "runtime"
    "sync"
    "time"
    "strings"
    "io"
    "regexp"
    "net/http"
    "golang.org/x/net/html"
)

var urlQueue chan string
var urlQueueMutex sync.Mutex

var visitedUrls map[string]bool
var visitedUrlsMutex sync.Mutex

func findLinks(body io.ReadCloser) []string {
    res := make([]string, 0)

    z := html.NewTokenizer(body)
    for {
        token := z.Next()
        if token == html.ErrorToken {
            break
        }

        tn, hasAttr := z.TagName()
        tagName := string(tn)
        if tagName == "a" && hasAttr {
            for {
                key, val, moreAttr := z.TagAttr()
                if string(key) == "href" {
                    res = append(res, string(val))
                }

                if moreAttr { continue }
                break
            }
        }
    }

    return res
}

func enqueueLinks(baseUrl string, links []string) {
    pathSepRegex, err := regexp.Compile("[a-z 0-9]/")
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    for i := range links {
        link := links[i]
        var sb strings.Builder

        if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
            sb.WriteString(link)
        } else if strings.HasPrefix(link, "//") {
            prefix := strings.Split(baseUrl, ":")[0]
            sb.WriteString(prefix)
            sb.WriteRune(':')
            sb.WriteString(link)
        } else if strings.HasPrefix(link, "/") {
            splitIdx := pathSepRegex.FindStringIndex(baseUrl)
            if splitIdx == nil {
                sb.WriteString(baseUrl)
            } else {
                sb.WriteString(baseUrl[:splitIdx[1]])
            }
            sb.WriteString(link)
        } else {
            sb.WriteString(baseUrl)
            if !strings.HasSuffix(baseUrl, "/") {
                sb.WriteRune('/')
            }
            sb.WriteString(link)
        }

        // TODO: What to do if queue is full?
        urlQueueMutex.Lock()
        if len(urlQueue) < cap(urlQueue) {
            fmt.Println(sb.String())
            urlQueue <- sb.String()
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
        }
        visitedUrlsMutex.Unlock()

        //fmt.Println("Fetching:", url)
        resp, err := client.Get(url)
        if err != nil {
            fmt.Println(err)
            continue
        }

        /*
        body, err := ioutil.ReadAll(resp.Body)
        fmt.Println(body)
        */

        contentType := resp.Header["Content-Type"][0]
        if strings.HasPrefix(contentType, "text/html;") {
            links := findLinks(resp.Body)
            enqueueLinks(url, links)
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

    urlQueue = make(chan string, 5000)
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
