package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const pixeldrainAPIBase = "https://pixeldrain.net/api"

// isTransient returns true for HTTP status codes worth retrying.
func isTransient(code int) bool {
	return code == http.StatusTooManyRequests ||
		code >= http.StatusInternalServerError
}

// ResolvePlaylist fetches a Pixeldrain list and returns its files.
// Retries up to 3 times with backoff on transient failures.
func ResolvePlaylist(listID string) (*PixeldrainList, error) {
	url := fmt.Sprintf("%s/list/%s", pixeldrainAPIBase, listID)
	backoff := [3]time.Duration{0, 2 * time.Second, 5 * time.Second}

	for attempt, delay := range backoff {
		if delay > 0 {
			time.Sleep(delay)
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("building request for pixeldrain list %s: %w", listID, err)
		}
		req.Header.Set("User-Agent", userAgent)

		resp, err := httpClient.Do(req)
		if err != nil {
			if attempt == len(backoff)-1 {
				return nil, fmt.Errorf("fetching pixeldrain list %s: %w", listID, err)
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if !isTransient(resp.StatusCode) {
				return nil, fmt.Errorf("pixeldrain list %s returned %d: %s", listID, resp.StatusCode, body)
			}
			if attempt == len(backoff)-1 {
				return nil, fmt.Errorf("pixeldrain list %s returned %d: %s", listID, resp.StatusCode, body)
			}
			continue
		}

		var list PixeldrainList
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			resp.Body.Close()
			if attempt == len(backoff)-1 {
				return nil, fmt.Errorf("decoding pixeldrain list %s: %w", listID, err)
			}
			continue
		}
		resp.Body.Close()
		return &list, nil
	}

	return nil, fmt.Errorf("pixeldrain list %s: all retries exhausted", listID)
}

// FileDownloadURL returns the direct download URL for a Pixeldrain file.
func FileDownloadURL(fileID string) string {
	return fmt.Sprintf("%s/file/%s?download", pixeldrainAPIBase, fileID)
}
