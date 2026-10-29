package store_test

import (
	"encoding/json"
	"os"
	"testing"
)

func TestEveryFixtureCitesItsSource(t *testing.T) {
	t.Parallel()
	var sources map[string]string
	if err := json.Unmarshal(fixture(t, "sources.json"), &sources); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if sources[entry.Name()] == "" {
			t.Errorf("fixture %s has no source", entry.Name())
		}
	}
	for name := range sources {
		if _, err := os.Stat("testdata/" + name); err != nil {
			t.Errorf("source for missing fixture %s", name)
		}
	}
}
