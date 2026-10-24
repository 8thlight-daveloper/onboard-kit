package materializer_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"

	api "onboard-kit/contracts/catalog/materializer"
	"onboard-kit/internal/model"
)

func holderBoundaryRequest(count int) api.Request {
	r := base()
	holders := make([]string, count)
	for i := range holders {
		holders[i] = "helper"
	}
	r.Directory = model.Directory{
		Contacts: []model.Contact{{ID: "helper", Name: "Helper", Email: "helper@example.test"}},
		Roles:    []model.Role{{ID: "helpers", Name: "Helpers", Holders: []string{"helper"}}},
		Teams:    []model.Team{{ID: "team-a", Name: "Team A", Holders: map[string][]string{"helpers": holders}}},
	}
	return r
}

func TestTeamHolderListCapIsInclusiveBeforeDirectoryValidation(t *testing.T) {
	for _, count := range []int{1024, 1025} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			r := holderBoundaryRequest(count)
			got, ds := newMaterializer().Create(r)
			if count == 1025 {
				own(t, got, ds, "request", "materialize.limit", "request", "Reduce the selection or held creation data to the supported limits.")
			} else {
				noErrors(t, ds)
				oracle := holderBoundaryRequest(count)
				want := expectedPlan(oracle)
				want.Directory = oracle.Directory
				if !reflect.DeepEqual(got, api.Result{Plan: want, Guidance: []api.Guidance{}}) || len(ds) != 0 {
					t.Error("1024 valid repeated team holders must materialize exactly without directory errors or deduplication")
				}
			}
			if !reflect.DeepEqual(r, holderBoundaryRequest(count)) {
				t.Error("holder admission changed the caller's repeated valid IDs")
			}
		})
	}
}

func TestInactiveValueListItemsStillRequireUTF8(t *testing.T) {
	for _, tc := range []struct {
		name    string
		request func() api.Request
	}{
		{"default-none", func() api.Request {
			r := base()
			r.PlanVariables = []model.Variable{{Name: "shared_fact", Default: model.Value{Kind: model.ValueKindNone, List: []string{string([]byte{255})}}}}
			return r
		}},
		{"binding-none", func() api.Request {
			r := base()
			r.PlanVariables = []model.Variable{{Name: "shared_fact"}}
			r.PlanBindings = []api.Binding{{Name: "shared_fact", Value: model.Value{Kind: model.ValueKindNone, List: []string{string([]byte{255})}}}}
			return r
		}},
		{"binding-text", func() api.Request {
			r := base()
			r.PlanVariables = []model.Variable{{Name: "shared_fact"}}
			r.PlanBindings = []api.Binding{{Name: "shared_fact", Value: model.Value{Kind: model.ValueKindText, Text: "accepted fact", List: []string{string([]byte{255})}}}}
			return r
		}},
		{"instance-binding-text", func() api.Request {
			r := base()
			r.Instances = []api.Instance{{SourceID: "repository-access", PermanentSectionID: "repo-a", Bindings: []api.Binding{{Name: "repository_scope", Value: model.Value{Kind: model.ValueKindText, Text: "approved capability", List: []string{string([]byte{255})}}}}}}
			return r
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.request()
			got, ds := newMaterializer().Create(r)
			own(t, got, ds, "request", "materialize.text", "request", "Use valid UTF8 creation data.")
			if !reflect.DeepEqual(r, tc.request()) {
				t.Error("invalid inactive UTF8 bytes or held request shape were changed")
			}
		})
	}
}

func invalidHolderKeyRequest() api.Request {
	r := holderBoundaryRequest(0)
	r.Directory.Teams[0].Holders = map[string][]string{string([]byte{255}): []string{}}
	return r
}

func TestTeamHolderMapKeysRequireUTF8BeforeDirectoryLookup(t *testing.T) {
	r := invalidHolderKeyRequest()
	got, ds := newMaterializer().Create(r)
	own(t, got, ds, "request", "materialize.text", "request", "Use valid UTF8 creation data.")
	if !reflect.DeepEqual(r, invalidHolderKeyRequest()) {
		t.Error("invalid map-key bytes or allocated-empty held value were normalized")
	}
}

func TestSourceIDHasItsOwnInclusiveByteCap(t *testing.T) {
	for _, size := range []int{4096, 4097} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			r := base()
			r.Instances = []api.Instance{{SourceID: strings.Repeat("x", size), PermanentSectionID: "activity"}}
			before := snapshot(t, r)
			got, ds := newMaterializer().Create(r)
			if size == 4097 {
				own(t, got, ds, "request", "materialize.limit", "request", "Reduce the selection or held creation data to the supported limits.")
			} else {
				want := []api.Diagnostic{{Stage: "source", Path: "instances[0].sourceId", Problem: model.Problem{Code: "catalog.unknown-id", Severity: model.SeverityError, Message: "Choose an exact ID from catalog list.", Detail: map[string]string{"field": "id"}}}}
				if !reflect.DeepEqual(got, api.Result{}) || !reflect.DeepEqual(ds, want) {
					t.Error("4096 source bytes must pass admission and reach the exact unknown-source refusal")
				}
			}
			if !bytes.Equal(snapshot(t, r), before) {
				t.Error("source admission changed or truncated the held ID")
			}
		})
	}
}

func TestPermanentSectionIDHasItsOwnInclusiveByteCap(t *testing.T) {
	for _, size := range []int{4096, 4097} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			r := base()
			id := strings.Repeat("x", size)
			r.Instances = []api.Instance{{SourceID: "repository-access", PermanentSectionID: id}}
			before := snapshot(t, r)
			got, ds := newMaterializer().Create(r)
			if size == 4097 {
				own(t, got, ds, "request", "materialize.limit", "request", "Reduce the selection or held creation data to the supported limits.")
			} else {
				noErrors(t, ds)
				canonical := fixture(t, "repository-a.canonical")
				if bytes.Count(canonical, []byte(`"id":"repo-a"`)) != 1 || len(canonical) == 0 || canonical[len(canonical)-1] != '\n' {
					t.Fatal("independent canonical fixture no longer isolates the permanent ID")
				}
				canonical = bytes.Replace(canonical[:len(canonical)-1], []byte(`"id":"repo-a"`), []byte(`"id":"`+id+`"`), 1)
				sum := sha256.Sum256(canonical)
				section := expectedRepository(t, "repo-a")
				section.ID, section.Source.Hash = id, hex.EncodeToString(sum[:])
				want := expectedPlan(r)
				want.Sections = []model.Section{section}
				if !reflect.DeepEqual(got.Plan, want) {
					t.Error("4096 permanent-ID bytes were refused, truncated or omitted from the exact unresolved baseline")
				}
				guides(t, got.Guidance, []guideOracle{{"invitation-prepare", []string{id}}, {"resource-access-arrange", []string{id}}})
			}
			if !bytes.Equal(snapshot(t, r), before) {
				t.Error("permanent-ID admission changed the request")
			}
		})
	}
}

func TestUnsupportedValueKindRefusesAtTheFirstTraversalPath(t *testing.T) {
	for _, tc := range []struct {
		name, path string
		request    func() api.Request
	}{
		{"defaults-before-later-values", "planVariables[0].default", func() api.Request {
			r := base()
			r.PlanVariables = []model.Variable{{Name: "first_fact", Default: model.Value{Kind: model.ValueKind(99)}}, {Name: "second_fact", Default: model.Value{Kind: model.ValueKind(98)}}}
			r.PlanBindings = []api.Binding{{Name: "first_fact", Value: model.Value{Kind: model.ValueKind(97)}}}
			return r
		}},
		{"condition-values-before-later-declarations", "planVariables[0].askIf.all[0].answer.is", func() api.Request {
			r := base()
			r.PlanVariables = []model.Variable{
				{Name: "first_fact", AskIf: &model.Condition{All: []model.Clause{{Question: "plan.first_fact", Answer: model.AnswerTest{Is: model.Value{Kind: model.ValueKind(99)}, IsNot: model.Value{Kind: model.ValueKind(98)}}}}}},
				{Name: "second_fact", Default: model.Value{Kind: model.ValueKind(97)}},
			}
			return r
		}},
		{"plan-bindings-before-instance-bindings", "planBindings[0].value", func() api.Request {
			r := base()
			r.PlanVariables = []model.Variable{{Name: "first_fact"}, {Name: "second_fact"}}
			r.PlanBindings = []api.Binding{{Name: "first_fact", Value: model.Value{Kind: model.ValueKind(99)}}, {Name: "second_fact", Value: model.Value{Kind: model.ValueKind(98)}}}
			r.Instances = []api.Instance{{SourceID: "repository-access", PermanentSectionID: "repo-a", Bindings: []api.Binding{{Name: "repository_scope", Value: model.Value{Kind: model.ValueKind(97)}}}}}
			return r
		}},
		{"instance-held-binding-order", "instances[0].bindings[0].value", func() api.Request {
			r := base()
			r.Instances = []api.Instance{
				{SourceID: "repository-access", PermanentSectionID: "repo-a", Bindings: []api.Binding{{Name: "repository_scope", Value: model.Value{Kind: model.ValueKind(99)}}, {Name: "repository_grant_route", Value: model.Value{Kind: model.ValueKind(98)}}}},
				{SourceID: "repository-access", PermanentSectionID: "repo-b", Bindings: []api.Binding{{Name: "repository_scope", Value: model.Value{Kind: model.ValueKind(97)}}}},
			}
			return r
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.request()
			got, ds := newMaterializer().Create(r)
			own(t, got, ds, "request", "materialize.value-kind", tc.path, "Choose a supported value kind.")
			if !reflect.DeepEqual(r, tc.request()) {
				t.Error("shape refusal changed the held unsupported values or request")
			}
		})
	}
}
