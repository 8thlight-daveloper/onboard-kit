package jsonform_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
)

func TestJSONReadsBackAsTheSamePlan(t *testing.T) {
	var paths []string
	for _, pattern := range []string{"../../plan/testdata/*.plan.json", "../../plan/*/testdata/*.plan.json"} {
		found, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, found...)
	}
	if len(paths) == 0 {
		t.Fatal("no fixtures found")
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			plan, err := jsonform.Plan(data)
			if err != nil {
				t.Fatal(err)
			}
			written, err := jsonform.JSON(plan)
			if err != nil {
				t.Fatal(err)
			}
			again, err := jsonform.Plan(written)
			if err != nil {
				t.Fatalf("written JSON does not read back: %v\n%s", err, written)
			}
			if !reflect.DeepEqual(plan, again) {
				t.Errorf("plan changed on the way through JSON\n%s", written)
			}
		})
	}
}
