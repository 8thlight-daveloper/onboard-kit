package assembler_test

import (
	"reflect"
	"slices"
	"testing"
)

func TestReferenceDocumentsShowExplicitPrerequisites(t *testing.T) {
	for _, c := range bootstrapCases(t, "requires-context.json", 3) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			plan := bootstrapPlan(t, c)
			before := bootstrapPlan(t, c)
			shared := newAssembler()
			for range 2 {
				document, problems := shared.Assemble(plan)
				assertBootstrap(t, c, document, problems)
				if !reflect.DeepEqual(plan, before) {
					t.Fatal("prerequisite display changed held definitions")
				}
			}
			slices.Reverse(plan.Sections)
			document, problems := shared.Assemble(plan)
			assertBootstrap(t, c, document, problems)
		})
	}
}
