package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	watchURL  = "https://onepace.net/en/watch"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

var (
	pushRe     = regexp.MustCompile(`self\.__next_f\.push\(\[1,"(.+?)"\]\)`)
	arcArrayRe = regexp.MustCompile(`\[[\s]*\{[\s]*"slug"[\s]*:`)
)

// FetchArcs returns arc data, using cache when available.
func FetchArcs(noCache bool) ([]Arc, error) {
	if !noCache {
		if arcs, ok := loadCache(); ok {
			return arcs, nil
		}
	}

	arcs, err := scrapeArcs()
	if err != nil {
		return nil, err
	}
	_ = saveCache(arcs)
	return arcs, nil
}

func scrapeArcs() ([]Arc, error) {
	var body []byte
	backoff := [3]time.Duration{0, 2 * time.Second, 5 * time.Second}

	for attempt, delay := range backoff {
		if delay > 0 {
			log.Printf("Retry %d/%d after %v...", attempt, len(backoff)-1, delay)
			time.Sleep(delay)
		}

		req, err := http.NewRequest("GET", watchURL, nil)
		if err != nil {
			return nil, fmt.Errorf("building request: %w", err)
		}
		req.Header.Set("User-Agent", userAgent)

		resp, err := httpClient.Do(req)
		if err != nil {
			if attempt == len(backoff)-1 {
				return nil, fmt.Errorf("fetching watch page: %w", err)
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			if attempt == len(backoff)-1 {
				return nil, fmt.Errorf("watch page returned %d", resp.StatusCode)
			}
			continue
		}

		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			if attempt == len(backoff)-1 {
				return nil, fmt.Errorf("reading watch page: %w", err)
			}
			continue
		}

		break
	}

	return extractArcs(string(body))
}

// extractArcs finds the arc JSON array embedded in the RSC payload.
func extractArcs(html string) ([]Arc, error) {
	matches := pushRe.FindAllStringSubmatch(html, -1)

	for _, m := range matches {
		payload := unescapeRSC(m[1])
		arcs, err := findArcArray(payload)
		if err == nil && len(arcs) > 0 {
			return arcs, nil
		}
	}

	arcs, err := findArcArray(html)
	if err == nil && len(arcs) > 0 {
		return arcs, nil
	}

	return nil, fmt.Errorf("could not find arc data in page (found %d RSC chunks)", len(matches))
}

func unescapeRSC(s string) string {
	s = strings.ReplaceAll(s, `\\`, `\`)
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, `\/`, "/")
	return s
}

// findArcArray scans text for a JSON array containing arc-shaped objects.
func findArcArray(text string) ([]Arc, error) {
	loc := arcArrayRe.FindStringIndex(text)
	if loc == nil {
		return nil, fmt.Errorf("no arc array found")
	}

	start := loc[0]
	jsonStr := extractJSONArray(text[start:])
	if jsonStr == "" {
		return nil, fmt.Errorf("could not extract JSON array")
	}

	// RSC payloads use the literal string "$undefined" for JavaScript undefined values;
	// replace with empty string so JSON unmarshal doesn't choke on unknown tokens.
	jsonStr = strings.ReplaceAll(jsonStr, `"$undefined"`, `""`)

	var arcs []Arc
	if err := json.Unmarshal([]byte(jsonStr), &arcs); err != nil {
		return nil, fmt.Errorf("parsing arc JSON: %w", err)
	}

	if len(arcs) == 0 || arcs[0].Slug == "" {
		return nil, fmt.Errorf("parsed array does not contain arcs")
	}

	return arcs, nil
}

// extractJSONArray extracts a balanced JSON array from the start of text.
func extractJSONArray(text string) string {
	if len(text) == 0 || text[0] != '[' {
		return ""
	}

	depth := 0
	inString := false
	escaped := false

	for i, c := range text {
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' && inString {
			escaped = true
			continue
		}
		if c == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}

		switch c {
		case '[', '{':
			depth++
		case ']', '}':
			depth--
			if depth == 0 {
				return text[:i+1]
			}
		}
	}
	return ""
}
