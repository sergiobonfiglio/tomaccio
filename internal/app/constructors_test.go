package app

import (
	"slices"
	"testing"

	"github.com/sergiobonfiglio/tomaccio/internal/config"
	tomagnetlib "github.com/sergiobonfiglio/tomagnet/pkg/tomagnet"
)

func TestNewSearchProvidersUsesTomagnetDefaultsWhenConfigOmitted(t *testing.T) {
	providers, errs := newSearchProviders(&config.Config{})
	if len(errs) != 0 {
		t.Fatalf("errs=%#v", errs)
	}

	got := make([]string, 0, len(providers))
	for _, provider := range providers {
		named, ok := provider.(interface{ Name() string })
		if !ok {
			t.Fatalf("provider %T does not expose Name()", provider)
		}
		got = append(got, named.Name())
	}

	want := make([]string, 0, len(tomagnetlib.DefaultIndexers()))
	for _, idx := range tomagnetlib.DefaultIndexers() {
		want = append(want, idx.ID)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("provider names=%v, want %v", got, want)
	}
}
