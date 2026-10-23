package reader

import "onboard-kit/internal/model"

type Carrier string

const (
	CarrierSectionTemplate     Carrier = "section-template"
	CarrierPreparationGuidance Carrier = "preparation-guidance"
	CarrierToolHelp            Carrier = "tool-help"
)

type Summary struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Category    string  `json:"category"`
	Goal        string  `json:"goal"`
	Result      string  `json:"result"`
	Carrier     Carrier `json:"carrier"`
	ContentPath string  `json:"contentPath"`
}

type Entry struct {
	Summary Summary
	Content []byte
}

type Reader interface {
	List() []Summary
	Read(id string) (Entry, []model.Problem)
}
