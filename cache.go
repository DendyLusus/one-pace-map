package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

const cacheTTL = 24 * time.Hour

type cacheEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Arcs      []Arc     `json:"arcs"`
}

func cacheDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "one-pace-map")
}

func cacheFile() string {
	return filepath.Join(cacheDir(), "arcs.json")
}

func loadCache() ([]Arc, bool) {
	data, err := os.ReadFile(cacheFile())
	if err != nil {
		return nil, false
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		log.Printf("WARN corrupt cache file, ignoring: %v", err)
		return nil, false
	}
	if time.Since(entry.Timestamp) > cacheTTL {
		return nil, false
	}
	return entry.Arcs, true
}

func saveCache(arcs []Arc) error {
	dir := cacheDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	entry := cacheEntry{Timestamp: time.Now(), Arcs: arcs}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "arcs-*.json.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, cacheFile())
}
