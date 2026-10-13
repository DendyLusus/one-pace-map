package main

import "encoding/json"

// Arc represents a One Pace story arc scraped from the watch page.
type Arc struct {
	Slug        string      `json:"slug"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Special     bool        `json:"special"`
	Chapters    Span        `json:"chapters"`
	Episodes    Span        `json:"episodes"`
	PlayGroups  []PlayGroup `json:"playlistGroups"`
}

// Span is a set of chapter or episode ranges covered by an arc.
type Span struct {
	Terms        []SpanTerm `json:"terms"`
	CoverStories bool       `json:"coverStories"`
}

// UnmarshalJSON tolerates the empty string left behind when the page sends
// "$undefined" in place of a span object.
func (s *Span) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		*s = Span{}
		return nil
	}
	type plain Span
	return json.Unmarshal(b, (*plain)(s))
}

// SpanTerm is either a numeric range (From/To) or free text (e.g. "Episode of East Blue").
type SpanTerm struct {
	From int    `json:"from"`
	To   int    `json:"to"`
	Text string `json:"text"`
}

// PlayGroup is a set of resolution-keyed playlists sharing the same language/variant.
type PlayGroup struct {
	Sub       string     `json:"sub"`
	Dub       string     `json:"dub"`
	Variant   string     `json:"variant"`
	Playlists []Playlist `json:"items"`
}

// Playlist maps a resolution to a Pixeldrain list ID.
type Playlist struct {
	ID         string `json:"id"`
	Resolution int    `json:"resolution"`
}

// PixeldrainList is the Pixeldrain API response for a list.
type PixeldrainList struct {
	Success   bool             `json:"success"`
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	FileCount int              `json:"file_count"`
	Files     []PixeldrainFile `json:"files"`
}

// PixeldrainFile is a single file entry in a Pixeldrain list.
type PixeldrainFile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// ResolvedArc holds the final output for one arc.
type ResolvedArc struct {
	Slug       string         `json:"slug"`
	Title      string         `json:"title"`
	Special    bool           `json:"special"`
	PlaylistID string         `json:"playlist_id"`
	Resolution int            `json:"resolution"`
	Sub        string         `json:"sub"`
	Dub        string         `json:"dub"`
	Variant    string         `json:"variant"`
	Episodes   []ResolvedFile `json:"episodes"`
}

// ResolvedFile is one downloadable episode file.
type ResolvedFile struct {
	EpisodeNum int    `json:"episode_num"`
	FileID     string `json:"file_id"`
	FileName   string `json:"file_name"`
	Size       int64  `json:"size"`
	URL        string `json:"url"`
}
