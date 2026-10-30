package editguide

import (
	"fmt"
	"strconv"

	editorcontract "onboard-kit/contracts/catalog/editor"
	"onboard-kit/internal/model"
)

func canonical(text string, low, high int) (int, bool) {
	number, err := strconv.Atoi(text)
	if err != nil || number < low || number > high || strconv.Itoa(number) != text {
		return 0, false
	}
	return number, true
}

func (s *session) stage(action string) {
	if len(s.operations) >= maxOperations || action == "add" && s.adds() >= maxAdds {
		s.write(limitReached)
		return
	}
	var operation editorcontract.Operation
	var ok bool
	switch action {
	case "add":
		operation, ok = s.addOperation()
	case "submit":
		operation, ok = s.submitOperation()
	case "unset":
		operation, ok = s.unsetOperation()
	case "retire":
		operation, ok = s.instanceOperation(editorcontract.Retire)
	case "reactivate":
		operation, ok = s.instanceOperation(editorcontract.Reactivate)
	}
	if !ok {
		return
	}
	s.operations = append(s.operations, operation)
	s.write("Operation staged; validation occurs at review.\n")
}

func (s *session) instanceOperation(kind editorcontract.Kind) (editorcontract.Operation, bool) {
	id, ok := s.ask("Instance ID:")
	return editorcontract.Operation{Kind: kind, SectionID: id}, ok
}

func (s *session) unsetOperation() (editorcontract.Operation, bool) {
	id, ok := s.ask("Instance ID:")
	if !ok {
		return editorcontract.Operation{}, false
	}
	name, ok := s.ask("Variable name:")
	return editorcontract.Operation{Kind: editorcontract.Unset, SectionID: id, Name: name}, ok
}

func (s *session) submitOperation() (editorcontract.Operation, bool) {
	id, ok := s.ask("Instance ID:")
	if !ok {
		return editorcontract.Operation{}, false
	}
	name, ok := s.ask("Variable name:")
	if !ok {
		return editorcontract.Operation{}, false
	}
	value, ok := s.value("Text value (empty text clears an existing answer):")
	return editorcontract.Operation{Kind: editorcontract.Submit, SectionID: id, Name: name, Value: value}, ok
}

func (s *session) addOperation() (editorcontract.Operation, bool) {
	source, ok := s.ask("Source ID:")
	if !ok {
		return editorcontract.Operation{}, false
	}
	id, ok := s.ask("Permanent section ID:")
	if !ok {
		return editorcontract.Operation{}, false
	}
	count, ok := s.count("Initial binding count: 0-64:", "Use a canonical count from 0 through 64.", 64)
	if !ok {
		return editorcontract.Operation{}, false
	}
	bindings := make([]editorcontract.Binding, 0, count)
	for i := 0; i < count; i++ {
		name, ok := s.ask("Binding name:")
		if !ok {
			return editorcontract.Operation{}, false
		}
		value, ok := s.value("Text value (empty text is an exact new binding):")
		if !ok {
			return editorcontract.Operation{}, false
		}
		bindings = append(bindings, editorcontract.Binding{Name: name, Value: value})
	}
	return editorcontract.Operation{Kind: editorcontract.Add, SectionID: id, SourceID: source, Bindings: bindings}, true
}

func (s *session) count(prompt, retry string, high int) (int, bool) {
	for {
		text, ok := s.ask(prompt)
		if !ok {
			return 0, false
		}
		if number, valid := canonical(text, 0, high); valid {
			return number, true
		}
		if !s.write(retry + "\n") {
			return 0, false
		}
	}
}

func (s *session) value(textPrompt string) (model.Value, bool) {
	for {
		kind, ok := s.ask("Value kind: text, flag, list, or deferred:")
		if !ok {
			return model.Value{}, false
		}
		switch kind {
		case "text":
			text, ok := s.ask(textPrompt)
			return model.Value{Kind: model.ValueKindText, Text: text}, ok
		case "deferred":
			return model.Value{Kind: model.ValueKindDeferred}, true
		case "flag":
			return s.flag()
		case "list":
			return s.list()
		}
		if !s.write(invalidAction) {
			return model.Value{}, false
		}
	}
}

func (s *session) flag() (model.Value, bool) {
	for {
		text, ok := s.ask("Flag value: true or false:")
		if !ok {
			return model.Value{}, false
		}
		if text == "true" || text == "false" {
			return model.Value{Kind: model.ValueKindFlag, Flag: text == "true"}, true
		}
		if !s.write("Use true or false.\n") {
			return model.Value{}, false
		}
	}
}

func (s *session) list() (model.Value, bool) {
	count, ok := s.count("List length: 0-4096:", "Use a canonical count from 0 through 4096.", 4096)
	if !ok {
		return model.Value{}, false
	}
	items := make([]string, 0, count)
	for i := 1; i <= count; i++ {
		item, ok := s.ask(fmt.Sprintf("List item %d:", i))
		if !ok {
			return model.Value{}, false
		}
		items = append(items, item)
	}
	return model.Value{Kind: model.ValueKindList, List: items}, true
}
