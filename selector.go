package main

// SelectBestPlaylist picks the best playlist for an arc according to the priority rules:
//
//	Variant:    extended > alternate > standard (empty)
//	Language:   prefer matching preferredDub, with sub=en ranked above sub=""
//	Resolution: 1080 > 720 > 480
func SelectBestPlaylist(arc Arc, preferredDub string) (PlayGroup, Playlist, bool) {
	var best *candidate

	for _, pg := range arc.PlayGroups {
		vr := variantRank(pg.Variant)
		lr := langRank(pg.Sub, pg.Dub, preferredDub)

		for _, pl := range pg.Playlists {
			rr := resolutionRank(pl.Resolution)

			c := candidate{
				group:          pg,
				playlist:       pl,
				variantRank:    vr,
				langRank:       lr,
				resolutionRank: rr,
			}

			if best == nil || betterThan(c, *best) {
				best = &c
			}
		}
	}

	if best == nil {
		return PlayGroup{}, Playlist{}, false
	}
	return best.group, best.playlist, true
}

type candidate struct {
	group          PlayGroup
	playlist       Playlist
	variantRank    int
	langRank       int
	resolutionRank int
}

func betterThan(a, b candidate) bool {
	if a.variantRank != b.variantRank {
		return a.variantRank > b.variantRank
	}
	if a.langRank != b.langRank {
		return a.langRank > b.langRank
	}
	return a.resolutionRank > b.resolutionRank
}

func variantRank(v string) int {
	switch v {
	case "extended":
		return 2
	case "alternate":
		return 1
	default:
		return 0
	}
}

func langRank(sub, dub, preferredDub string) int {
	altDub := "ja"
	if preferredDub == "ja" {
		altDub = "en"
	}

	switch {
	case sub == "en" && dub == preferredDub:
		return 3
	case sub == "" && dub == preferredDub:
		return 2
	case sub == "en" && dub == altDub:
		return 1
	default:
		return 0
	}
}

func resolutionRank(r int) int {
	return r
}
