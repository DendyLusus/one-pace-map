package main

import (
	"regexp"
	"strconv"
)

// episodeRe matches the episode number from One Pace filenames like:
//
//	[One Pace][1] Romance Dawn 01 [1080p][En Sub][7E9134A5].mp4
//	[One Pace][3-5] Romance Dawn 03 [1080p][En Sub][4AADE29F].mp4
//	[One Pace][79-81] Arlong Park 05 Extended [1080p][En Dub][0326DF53].mp4
//
// We extract the "NN" after the arc name — there may be a variant word like "Extended" before [.
var episodeRe = regexp.MustCompile(`(?i)\]\s+\S.*?\s+(\d+)\s+(?:\S+\s+)?\[`)

// ParseEpisodeNum extracts the episode sequence number from a One Pace filename.
// Returns 0 if no number is found.
func ParseEpisodeNum(filename string) int {
	m := episodeRe.FindStringSubmatch(filename)
	if m == nil {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return n
}

// DeduplicateFiles keeps only the first file per episode number.
// Files with episode number 0 (unparseable) are always kept.
func DeduplicateFiles(files []ResolvedFile) []ResolvedFile {
	seen := make(map[int]bool)
	var out []ResolvedFile
	for _, f := range files {
		if f.EpisodeNum == 0 || !seen[f.EpisodeNum] {
			seen[f.EpisodeNum] = true
			out = append(out, f)
		}
	}
	return out
}
