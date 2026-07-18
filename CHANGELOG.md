# Changelog

All notable changes to this project will be documented in this file.

## [v0.1.8] - 2026-06-09

### Changed
- Updated the `tomagnet` integration to use `github.com/sergiobonfiglio/tomagnet v0.3.9`.

## [v0.1.7] - 2026-06-08

### Changed
- Updated the `tomagnet` integration to use `github.com/sergiobonfiglio/tomagnet v0.3.8`.
- `tomaccio definitions sync` now writes tomagnet-managed definitions to `.tomaccio/definitions`.
- Tomagnet-backed search now reads definitions from `.tomaccio/definitions`.

## [v0.1.6] - 2026-06-08

### Removed
- Removed the repository-local `definitions/btdig.yml`; indexer definitions are owned and synced by `tomagnet`.

## [v0.1.5] - 2026-06-08

### Changed
- Updated the `tomagnet` integration to use `github.com/sergiobonfiglio/tomagnet v0.3.7`.

### Fixed
- `tomaccio definitions sync` now picks up tomagnet-bundled definitions such as `btdig` from the tomagnet release.

## [v0.1.4] - 2026-06-06

### Changed
- Updated the `tomagnet` integration to use `github.com/sergiobonfiglio/tomagnet v0.3.6`.
- Picked up tomagnet's new default public indexer set, including `btdig`.
- Kept `go.mod` pinned to the released `tomagnet` version.

### Fixed
- Searches against `btdig` now work through the released tomagnet transport and definition updates.
- Explicit `--indexer` selection now works even when the chosen indexer is not present in `tomagnet.yaml`.

## [v0.1.3] - 2026-06-06

### Changed
- Updated the `tomagnet` integration to use `github.com/sergiobonfiglio/tomagnet v0.3.5`.
- Switched the tomagnet-backed search adapter to the new intent-based search API.
- Kept `go.mod` pinned to the released `tomagnet` version.

### Fixed
- Search requests now rely on tomagnet's centralized request planning and optional-parameter omission behavior.
- Tomagnet-backed searches now pick up the request rendering fixes included in `tomagnet v0.3.5`.

## Historical releases

This changelog was introduced after the releases below. Detailed per-release notes were not recorded here at the time.

- `v0.1.2`
- `v0.1.1`
- `v0.1.0`
