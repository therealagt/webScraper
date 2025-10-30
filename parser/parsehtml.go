package parser

import (
	"regexp"
	"strings"
	"time"
)

/* search parameters */
func ParseFunc(url string, html []byte, keyword string) (string, string, time.Time) {
	htmlStr := string(html)
	var price string 
	if strings.Contains(htmlStr, keyword) {
        re := regexp.MustCompile(keyword + `\s*([0-9]+(?:[.,][0-9]+)?)`)
        match := re.FindStringSubmatch(htmlStr)
        if len(match) > 1 {
            price = "found: " + match[1]
        } else {
            price = "found, but no number"
        }
    } else {
        price = "not found"
    }

	var title string
    afterSlash := url[strings.LastIndex(url, "/")+1:]
	firstDot := strings.Index(afterSlash, ".")
	secondDot := strings.Index(afterSlash[firstDot+1:], ".")
	if firstDot != -1 && secondDot != -1 {
		title = afterSlash[firstDot+1 : firstDot+1+secondDot]
	} else {
		title = afterSlash
	}
	completedAt := time.Now()
	return price, title, completedAt
}
