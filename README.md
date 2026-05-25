# tomaccio

Taste Oriented Media Assistant Accio: a small, local-first CLI for searching media releases, sending downloads to Transmission, and exporting watched-movie history from Plex.

Tomaccio is intentionally lightweight. It has no database, no background daemon, and no recommendation engine in v0; it focuses on exposing simple commands that humans or automation agents can compose.

## Features

- Search configured release providers through [`tomagnet`](https://github.com/sergiobonfiglio/tomagnet).
- Add magnet or torrent URLs to Transmission.
- List current Transmission downloads and progress.
- Sync public indexer definitions for first-run setup.
- Read watched movie history from Plex.
- Emit watched movies as JSON by default for easy scripting.
- Keep secrets out of config files with `${ENV_VAR}` expansion.

## Requirements

- Go 1.26 or newer
- A Transmission RPC endpoint for download commands
- A Plex server and token for `tomaccio watched`
- Synced tomagnet indexer definitions for search commands (`tomaccio definitions sync`)

## Installation

Build from source:

```bash
git clone https://github.com/sergiobonfiglio/tomaccio.git
cd tomaccio
go build -o tomaccio ./cmd/tomaccio
```

If you use [`just`](https://github.com/casey/just), you can also run:

```bash
just build
```

## Quick start

Create a local config file:

```bash
cp config.example.yaml config.yaml
```

Set any secrets referenced by the config:

```bash
export TRANSMISSION_USERNAME='...'
export TRANSMISSION_PASSWORD='...'
export PLEX_TOKEN='...'
```

Run commands:

```bash
./tomaccio definitions sync
./tomaccio download check
./tomaccio search "The Matrix 1999"
./tomaccio watched
./tomaccio download add "magnet:?xt=urn:btih:..."
./tomaccio download list
```

Use `--config` to load a different YAML file:

```bash
./tomaccio --config ./config.local.yaml search "Alien 1979"
```

## Configuration

See [`config.example.yaml`](config.example.yaml) for a complete example.

```yaml
app:
  log_level: info

download:
  label: "tomaccio"
  dir_aliases:
    movies: "/media/usb-drive/movies"
    series: "/media/usb-drive/series"
  transmission:
    url: "https://transmission.example.com/transmission/rpc"
    username: "${TRANSMISSION_USERNAME}"
    password: "${TRANSMISSION_PASSWORD}"
    download_dir: "/media/usb-drive/movies"

# search.providers is optional. If omitted, tomaccio uses default public indexers.
watched:
  plex:
    url: "http://plex.example.com:32400"
    token: "${PLEX_TOKEN}"
```

Environment variables in YAML are expanded when the config is loaded, so secrets can stay out of files committed to git. If `download.label` is omitted, tomaccio defaults to the Transmission label `tomaccio`; set `download.label: ""` to disable labels entirely.
### Search providers

Search is powered by [`tomagnet`](https://github.com/sergiobonfiglio/tomagnet). If `search.providers` is omitted, tomaccio uses the default public indexers: `yts`, `1337x`, and `thepiratebay`.

Before the first public-indexer search, sync definitions:

```bash
tomaccio definitions sync
```

Definition storage and loading are managed by tomagnet.

Custom providers can be configured with:

- `name`: display name used in results and warnings
- `indexer_id`: tomagnet indexer definition id
- `base_url`: optional indexer base URL override
- `timeout_seconds`: optional per-provider timeout

Provider failures are printed as warnings while successful providers still return results.

## Commands

### `definitions sync`

Sync public indexer definitions managed by tomagnet.

```bash
tomaccio definitions sync
```

### `download check`

Check that the configured Transmission endpoint is reachable.

```bash
tomaccio download check
```

### `download list`

List current Transmission downloads.

```bash
tomaccio download list
tomaccio download list --label tomaccio
```

Output format:

```text
<ID>    <status>    <progress%>    <labels>    <title>
```

### `download dirs`

List configured download directory aliases.

```bash
tomaccio download dirs
```

Output format:

```text
movies	/media/usb-drive/movies
series	/media/usb-drive/series
```

### `download add`

Add a magnet or torrent URL to Transmission.

```bash
tomaccio download add "magnet:?xt=urn:btih:..."
# or
tomaccio download add --url "magnet:?xt=urn:btih:..."
# override download directory for this item with a path
tomaccio download add --dir /path/to/movies "magnet:?xt=urn:btih:..."
# or use a configured alias
tomaccio download add --dir movies "magnet:?xt=urn:btih:..."
```

Tomaccio applies the configured `download.label` to added torrents. If `download.label` is omitted, it defaults to `tomaccio`. If it is set to an empty string, no label is sent.

### `search`

Search configured providers for a movie release.

```bash
tomaccio search "The Matrix 1999"
tomaccio search "The Matrix (1999)"
```

If the query ends with a year, tomaccio passes the title and year separately to the search provider.

### `watched`

List watched movies from Plex.

```bash
tomaccio watched
```

JSON is the default output for scripting:

```bash
tomaccio watched --format json
```

For tab-separated text:

```bash
tomaccio watched --format text
```

## Development

Common commands:

```bash
just fmt
just test
just build
just check
```

Equivalent Go commands:

```bash
gofmt -w cmd internal
go test ./...
go build -o tomaccio ./cmd/tomaccio
```

## Notes

Tomaccio only coordinates tools you configure. Use it with services and content you are authorized to access.

## License

[MIT](LICENSE)
