package main

import (
	"fmt"
	"strings"
)

func printTable(results []ResolvedArc) {
	hasEpisodes := false
	for _, ra := range results {
		if len(ra.Episodes) > 0 {
			hasEpisodes = true
			break
		}
	}

	if hasEpisodes {
		printEpisodeTable(results)
	} else {
		printArcTable(results)
	}
}

func printArcTable(results []ResolvedArc) {
	slugWidth := len("ARC")
	for _, ra := range results {
		if len(ra.Slug) > slugWidth {
			slugWidth = len(ra.Slug)
		}
	}

	header := fmt.Sprintf("  # │ %-*s │ RES    │ VARIANT    │ SUB │ DUB │ PLAYLIST", slugWidth, "ARC")
	sep := fmt.Sprintf("────┼─%s─┼────────┼────────────┼─────┼─────┼──────────", strings.Repeat("─", slugWidth))
	fmt.Println(header)
	fmt.Println(sep)

	for i, ra := range results {
		variant := ra.Variant
		if variant == "" {
			variant = "standard"
		}
		sub := ra.Sub
		if sub == "" {
			sub = "—"
		}
		fmt.Printf(" %2d │ %-*s │ %4dp  │ %-10s │ %-3s │ %-3s │ %s\n",
			i+1, slugWidth, ra.Slug, ra.Resolution, variant, sub, ra.Dub, ra.PlaylistID)
	}

	fmt.Printf("\n%d arcs\n", len(results))
}

func printEpisodeTable(results []ResolvedArc) {
	slugWidth := len("ARC")
	nameWidth := len("FILE")
	for _, ra := range results {
		if len(ra.Slug) > slugWidth {
			slugWidth = len(ra.Slug)
		}
		for _, ep := range ra.Episodes {
			if len(ep.FileName) > nameWidth {
				nameWidth = len(ep.FileName)
			}
		}
	}
	if nameWidth > 80 {
		nameWidth = 80
	}

	header := fmt.Sprintf("  # │ %-*s │ EP  │ SIZE (MB) │ FILE ID  │ %-*s", slugWidth, "ARC", nameWidth, "FILE")
	sep := fmt.Sprintf("────┼─%s─┼─────┼───────────┼──────────┼─%s", strings.Repeat("─", slugWidth), strings.Repeat("─", nameWidth))
	fmt.Println(header)
	fmt.Println(sep)

	row := 0
	for _, ra := range results {
		if len(ra.Episodes) == 0 {
			row++
			fmt.Printf(" %2d │ %-*s │  —  │    —      │    —     │ (no files resolved)\n",
				row, slugWidth, ra.Slug)
			continue
		}
		for _, ep := range ra.Episodes {
			row++
			name := ep.FileName
			if len(name) > nameWidth {
				name = name[:nameWidth-1] + "…"
			}
			fmt.Printf("%3d │ %-*s │ %3d │ %9.1f │ %-8s │ %s\n",
				row, slugWidth, ra.Slug, ep.EpisodeNum,
				float64(ep.Size)/(1024*1024), ep.FileID, name)
		}
	}

	total := 0
	for _, ra := range results {
		total += len(ra.Episodes)
	}
	fmt.Printf("\n%d arcs, %d episodes\n", len(results), total)
}
