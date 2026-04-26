// Package main implements op-map, a CLI that scrapes onepace.net for arc metadata
// and resolves Pixeldrain playlists to downloadable episode links.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"sync"
	"sync/atomic"
)

func main() {
	noCache := flag.Bool("no-cache", false, "bypass local cache and re-scrape")
	arcFilter := flag.String("arc", "", "only process a specific arc by slug (e.g. romance-dawn)")
	resolve := flag.Bool("resolve", false, "resolve pixeldrain playlists to individual files")
	jsonOut := flag.Bool("json", false, "output as JSON")
	dub := flag.String("dub", "en", "preferred dub language: en or ja")
	flag.Parse()

	if *dub != "en" && *dub != "ja" {
		log.Fatalf("Invalid --dub value %q: must be en or ja", *dub)
	}

	arcs, err := FetchArcs(*noCache)
	if err != nil {
		log.Fatalf("Error fetching arcs: %v", err)
	}

	if *arcFilter != "" {
		filtered := make([]Arc, 0, 1)
		for _, a := range arcs {
			if a.Slug == *arcFilter {
				filtered = append(filtered, a)
			}
		}
		if len(filtered) == 0 {
			log.Fatalf("Arc %q not found", *arcFilter)
		}
		arcs = filtered
	}

	type indexedArc struct {
		index      int
		ra         ResolvedArc
		playlistID string
	}

	pending := make([]indexedArc, 0, len(arcs))
	results := make([]ResolvedArc, 0, len(arcs))

	for _, arc := range arcs {
		pg, pl, ok := SelectBestPlaylist(arc, *dub)
		if !ok {
			log.Printf("SKIP %s: no playlists available", arc.Slug)
			continue
		}

		ra := ResolvedArc{
			Slug:       arc.Slug,
			Title:      arc.Title,
			Special:    arc.Special,
			PlaylistID: pl.ID,
			Resolution: pl.Resolution,
			Sub:        pg.Sub,
			Dub:        pg.Dub,
			Variant:    pg.Variant,
			Episodes:   []ResolvedFile{},
		}
		results = append(results, ra)
		if *resolve {
			pending = append(pending, indexedArc{index: len(results) - 1, ra: ra, playlistID: pl.ID})
		}
	}

	// Resolve playlists concurrently (max 5 in-flight).
	if len(pending) > 0 {
		var mu sync.Mutex
		var failures atomic.Int32
		sem := make(chan struct{}, 5)
		var wg sync.WaitGroup

		for _, item := range pending {
			wg.Add(1)
			go func(it indexedArc) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				resolvedList, err := ResolvePlaylist(it.playlistID)
				if err != nil {
					log.Printf("WARN %s: %v", it.ra.Slug, err)
					failures.Add(1)
					return
				}
				files := make([]ResolvedFile, 0, len(resolvedList.Files))
				for _, f := range resolvedList.Files {
					files = append(files, ResolvedFile{
						EpisodeNum: ParseEpisodeNum(f.Name),
						FileID:     f.ID,
						FileName:   f.Name,
						Size:       f.Size,
						URL:        FileDownloadURL(f.ID),
					})
				}

				mu.Lock()
				results[it.index].Episodes = DeduplicateFiles(files)
				mu.Unlock()
			}(item)
		}
		wg.Wait()

		if int(failures.Load()) == len(pending) {
			log.Fatalf("All %d playlist resolves failed", len(pending))
		}
	}

	if len(results) == 0 {
		log.Fatalf("No arcs matched — site structure may have changed")
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(results); err != nil {
			log.Fatalf("Error encoding JSON: %v", err)
		}
		return
	}

	printTable(results)
}
