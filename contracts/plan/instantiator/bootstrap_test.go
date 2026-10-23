package instantiator_test

import (
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestBootstrapInstantiationPreservesDefinition(t *testing.T) {
	t.Parallel()
	for _, marker := range []bool{false, true} {
		template := model.Section{ID: "access", Title: "Access", Summary: "Request access", Parent: "welcome", Bootstrap: marker}
		templates := []model.Section{template}
		got, problems := newInstantiator().Instantiate(templates, nil, "set", "2026-01-02")
		wantHash := "7d499ca40c6beb131cbd5fe60630e18d6f328eea6194a37e785dcf0ba0956d99"
		if marker {
			wantHash = "f535f7f5bd75a85be203c32c910e72e5372fe09ad2679527b1e325c418c19ec2"
		}
		wantSection := template
		wantSection.Source = &model.Source{Template: "access", Hash: wantHash}
		want := model.Plan{SchemaVersion: 1, Provenance: model.Provenance{TemplateSet: "set", GeneratedOn: "2026-01-02"}, Sections: []model.Section{wantSection}}
		if len(problems) != 0 || !reflect.DeepEqual(got, want) {
			t.Fatalf("plan %#v, problems %#v; want %#v", got, problems, want)
		}
		if !reflect.DeepEqual(templates, []model.Section{template}) {
			t.Fatal("template changed")
		}
		got.Sections[0].Bootstrap = !marker
		if templates[0].Bootstrap != marker {
			t.Fatal("returned section aliases template")
		}
	}
}
