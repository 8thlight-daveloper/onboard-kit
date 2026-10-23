package splitter_test

import (
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestFirstOpenedAtSurvivesAnswerEdits(t *testing.T) {
	for _, held := range []model.Timestamp{"", "2026-10-23T12:34:56Z", " held\nnot a time\u2028\u2029 "} {
		t.Run(string(held), func(t *testing.T) {
			plan := model.Plan{Variables: []model.Variable{{Name: "q", Audience: model.AudienceOnboardee}}}
			joined := plan
			joined.Answers = model.Answers{Plan: map[string]model.Value{"q": {Kind: model.ValueKindText, Text: "edited"}}}
			previous := model.Record{SchemaVersion: 1, StartDate: "legacy date", FirstOpenedAt: held}
			want := previous
			want.Answers = model.Answers{Plan: map[string]model.Value{"q": {Kind: model.ValueKindText, Text: "edited"}}}
			got := newSplitter().Split(plan, joined, previous)
			if !reflect.DeepEqual(got, want) || previous.FirstOpenedAt != held {
				t.Fatalf("got %#v, want %#v; previous %#v", got, want, previous)
			}
		})
	}
}
