# Changelog

All notable changes to this project will be documented in this file.

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
