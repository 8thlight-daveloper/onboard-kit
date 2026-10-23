package reader

import (
	contract "onboard-kit/contracts/catalog/reader"
	"onboard-kit/internal/model"
)

type reader struct{}

type definition struct {
	summary contract.Summary
	content string
}

func New() contract.Reader {
	return reader{}
}

func (reader) List() []contract.Summary {
	definitions := inventory()
	result := make([]contract.Summary, len(definitions))
	for index, definition := range definitions {
		result[index] = definition.summary
	}
	return result
}

func (reader) Read(id string) (contract.Entry, []model.Problem) {
	for _, definition := range inventory() {
		if definition.summary.ID == id {
			return contract.Entry{Summary: definition.summary, Content: []byte(definition.content)}, nil
		}
	}
	return contract.Entry{}, []model.Problem{{
		Code: "catalog.unknown-id", Severity: model.SeverityError,
		Message: "Choose an exact ID from catalog list.",
		Detail:  map[string]string{"field": "id"},
	}}
}
