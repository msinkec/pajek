package main


import (
    "fmt"
    "os"
    "strings"
    "regexp"
)

func getFullUrl(baseUrl string, path string) string {
    pathSepRegex, err := regexp.Compile("[a-z 0-9]/")
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    var sb strings.Builder

    if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
        sb.WriteString(path)
    } else if strings.HasPrefix(path, "//") {
        prefix := strings.Split(baseUrl, ":")[0]
        sb.WriteString(prefix)
        sb.WriteRune(':')
        sb.WriteString(path)
    } else if strings.HasPrefix(path, "/") {
        splitIdx := pathSepRegex.FindStringIndex(baseUrl)
        if splitIdx == nil {
            sb.WriteString(baseUrl)
        } else {
            sb.WriteString(baseUrl[:splitIdx[1]])
        }
        sb.WriteString(path)
    } else {
        sb.WriteString(baseUrl)
        if !strings.HasSuffix(baseUrl, "/") {
            sb.WriteRune('/')
        }
        sb.WriteString(path)
    }

    return sb.String()
}

