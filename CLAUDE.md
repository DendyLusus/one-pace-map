# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

op-map is a Go CLI that scrapes onepace.net for arc metadata and resolves Pixeldrain playlists to downloadable episode links. Binary name: `op-map`, module name: `op-map`. Zero external dependencies (stdlib only).

## Build & Run

```bash
go build -o op-map .        # build
./op-map                     # list arcs as table
./op-map --resolve --json    # full episode index as JSON
./op-map --dub ja --json     # Japanese dub arcs
```

No tests, no linter config, no Makefile.

## Architecture

Single `main` package, flat structure. Data flows in a pipeline:

1. **onepace.go** — Scrapes onepace.net watch page, parses Next.js RSC payload to extract arc JSON. Retry with backoff. Exports `FetchArcs()`.
2. **cache.go** — 24h disk cache at `~/.cache/one-pace-map/arcs.json`. Atomic writes via temp file + rename.
3. **selector.go** — `SelectBestPlaylist()` picks the best playlist per arc. Ranking: variant (extended > alternate > standard) > language (preferred dub + en sub highest) > resolution.
4. **pixeldrain.go** — `ResolvePlaylist()` fetches file lists from Pixeldrain API. Retry with backoff on transient errors (429/5xx).
5. **episode.go** — `ParseEpisodeNum()` extracts episode numbers from One Pace filenames via regex. `DeduplicateFiles()` keeps first per episode.
6. **main.go** — CLI entry point. Flags: `--dub`, `--resolve`, `--json`, `--arc`, `--no-cache`. Concurrent playlist resolution (max 5 goroutines, semaphore pattern).
7. **table.go** — Terminal table output (arc table or episode table depending on `--resolve`).
8. **types.go** — All data structures: `Arc`, `PlayGroup`, `Playlist`, `PixeldrainList`, `PixeldrainFile`, `ResolvedArc`, `ResolvedFile`.

## CI/CD

- **`.github/workflows/release.yml`** — Triggered on `v*` tags. GoReleaser builds cross-platform binaries (Linux amd64, macOS universal, Windows amd64).
- **`.github/workflows/update-data.yml`** — Runs daily at 01:23 UTC. Regenerates pre-parsed JSON in `data/`, commits only if data changed.
- **`.goreleaser.yml`** — GoReleaser v2 config. Custom `name_template` produces `op-map-linux`, `op-map-macos`, `op-map-windows.exe`.
