# tomaccio

Taste Oriented Media Assistant Accio.

v0 is a local media downloader CLI in Go: search configured tomagnet indexers, enqueue magnet/torrent URLs in Transmission, and read download status directly from Transmission. No recommendation engine or database in v0.

## Quick start

```bash
cp config.example.yaml config.yaml
just build

./tomaccio download check
./tomaccio search "The Matrix 1999"
./tomaccio watched
./tomaccio download add "magnet:?xt=urn:btih:..."
./tomaccio download list
```

Secrets may be referenced with `${ENV_VAR}` in YAML.

## Configuration

See `config.example.yaml`. Main sections:
- `app`
- `download`
- `search`
- `watched`

Search providers use [`tomagnet`](https://github.com/sergiobonfiglio/tomagnet): configure `name`, `indexer_id`, optional `base_url`, and optional `timeout_seconds`. Definitions resolve from `./definitions/<indexer_id>.yml|yaml`, then `./.tomagnet/definitions/<indexer_id>.yml|yaml`. Provider failures are reported per provider while successful providers still return results.

Watched movies use Plex: configure `watched.plex.url` and `watched.plex.token`. `tomaccio watched` emits JSON by default for agent consumption; use `--format text` for tab-separated text.

## Commands

- `tomaccio download check`
- `tomaccio download list`
- `tomaccio download add URL` or `tomaccio download add --url URL`
- `tomaccio search "Title Year"`
- `tomaccio watched [--format json|text]`
