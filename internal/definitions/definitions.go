package definitions

import tomagnetlib "github.com/sergiobonfiglio/tomagnet/pkg/tomagnet"

// CacheDir is tomaccio's local cache for tomagnet-managed indexer definitions.
const CacheDir = ".tomaccio/definitions"

// Sync refreshes tomaccio's local cache of tomagnet-managed public definitions.
func Sync() (tomagnetlib.DefinitionsMetadata, error) {
	return tomagnetlib.SyncDefinitionsTo(CacheDir)
}

// LoadByID loads a definition by ID from tomaccio's local definitions cache.
func LoadByID(id string) (*tomagnetlib.Definition, error) {
	return tomagnetlib.LoadDefinitionByIDFrom(id, CacheDir)
}
