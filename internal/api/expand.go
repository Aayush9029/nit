package api

import (
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

var tcoRe = regexp.MustCompile(`https?://t\.co/[A-Za-z0-9]+`)

var expandClient = &http.Client{
	Timeout: 5 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse // don't follow, just capture Location
	},
}

// ExpandTcoLinks replaces t.co short links with their expanded URLs.
// Expands up to 10 links concurrently per call.
func ExpandTcoLinks(text string) string {
	links := tcoRe.FindAllString(text, 10)
	if len(links) == 0 {
		return text
	}

	// Deduplicate
	unique := make(map[string]bool)
	for _, l := range links {
		unique[l] = true
	}

	type result struct {
		short, long string
	}

	results := make(chan result, len(unique))
	var wg sync.WaitGroup

	for short := range unique {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			resp, err := expandClient.Head(u)
			if err != nil {
				return
			}
			resp.Body.Close()
			loc := resp.Header.Get("Location")
			if loc != "" && loc != u {
				results <- result{short: u, long: loc}
			}
		}(short)
	}

	go func() { wg.Wait(); close(results) }()

	for r := range results {
		text = strings.ReplaceAll(text, r.short, r.long)
	}
	return text
}
