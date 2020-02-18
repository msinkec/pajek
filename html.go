package main


import (
    "fmt"
    "bytes"
    "strings"
    "io/ioutil"
    "golang.org/x/net/html"
)

func parseBody(data []byte, baseUrl string) {
    reader := ioutil.NopCloser(bytes.NewReader(data))
    z := html.NewTokenizer(reader)

    for {
        token := z.Next()
        if token == html.ErrorToken {
            break
        }

        tn, hasAttr := z.TagName()
        tagName := string(tn)
        if tagName == "script" && hasAttr {
            for {
                key, val, moreAttr := z.TagAttr()
                if string(key) == "src" {
                    srcUrl := getFullUrl(baseUrl, string(val))
                    fmt.Println(srcUrl)
                }

                if moreAttr { continue }
                break
            }
        }

        text := string(z.Text())

        if len(text) > 0 && strings.TrimSpace(text) != "" {
            //fmt.Println(text)
        }
    }

    reader.Close()
}

func findLinks(data []byte) []string {
    res := make([]string, 0)

    reader := ioutil.NopCloser(bytes.NewReader(data))
    z := html.NewTokenizer(reader)

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

    reader.Close()
    return res
}

