# one-pace-map
A CLI that scrapes the [One Pace](https://onepace.net/en/watch) project website and maps every arc to the best downloadable Pixeldrain episode links. One Pace has multiple playlists — dubbed, subbed, various variants and qualities. The tool picks one playlist per arc by ranking:
1. Variant: Extended > Alternate > Standard
2. Language (sub=en, dub=en) > dub=en > (sub=en, dub=ja) — overwritten with `--dub ja`
3. Quality: 1080p > 720p > 480p

The structured JSON output can be used to automate bulk downloads, populate a media server, or build your own tool on top.

## Getting Started
> [!NOTE]
> Pre-parsed JSON files are available in [`data/`](data/) (refreshed daily).

Download a pre-built binary from [Releases](https://github.com/DendyLusus/one-pace-map/releases/latest).

Or build from source (requires Go v1.26+):
```bash
git clone https://github.com/DendyLusus/one-pace-map.git
cd one-pace-map
go build -o op-map .
```

## Usage
List all arcs with their best playlist:
```bash
op-map
```
Resolve episodes for a single arc:
```bash
op-map --resolve --arc water-seven
```
Generate a full episode index as JSON with Japanese dub:
```bash
op-map --resolve --dub ja --json > episodes.json
```

## Flags
| Flag | Default | Description |
|------|---------|-------------|
| `--dub` | `en` | Preferred dub language: `en` or `ja` |
| `--resolve` | `false` | Resolve playlists to individual episode files |
| `--json` | `false` | Output as JSON instead of a table |
| `--arc` | `all` | Filter to a single arc by slug |
| `--no-cache` | `false` | Bypass the 24h local cache and re-scrape |