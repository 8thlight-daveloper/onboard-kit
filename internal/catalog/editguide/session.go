package editguide

import (
	"io"

	editorcontract "onboard-kit/contracts/catalog/editor"
	"onboard-kit/internal/model"
)

const (
	maxOperations = 1024
	maxAdds       = 256
)

type session struct {
	command    command
	input      io.Reader
	output     io.Writer
	consumed   int
	failure    string
	ended      bool
	selection  *selection
	original   model.Plan
	operations []editorcontract.Operation
}

func (s *session) write(text string) bool {
	if s.failure != "" {
		return false
	}
	if !writeAll(s.output, text) {
		s.failure = "output"
		return false
	}
	return true
}

func (s *session) ask(prompt string) (string, bool) {
	if !s.write(prompt + "\n") {
		return "", false
	}
	return s.line()
}

func (s *session) stopped() (string, bool) {
	if s.failure != "" {
		return s.failure, true
	}
	if s.ended {
		return "cancelled", true
	}
	return "", false
}

func (s *session) run() string {
	if s.failure != "" {
		return s.failure
	}
	for {
		action, ok := s.line()
		if !ok {
			code, _ := s.stopped()
			return code
		}
		switch action {
		case "cancel":
			return "cancelled"
		case "add", "submit", "unset", "retire", "reactivate":
			s.stage(action)
		case "drop":
			s.drop()
		case "review":
			if code, done := s.consent(); done {
				return code
			}
		default:
			s.write(invalidAction)
		}
		if code, done := s.stopped(); done {
			return code
		}
		if !s.write(draftPrompt) {
			return s.failure
		}
	}
}

func (s *session) consent() (string, bool) {
	frozen, ok := s.review()
	if !ok {
		return s.stopped()
	}
	for {
		choice, ok := s.line()
		if !ok {
			return s.stopped()
		}
		switch choice {
		case "cancel":
			return "cancelled", true
		case "save":
			return s.selection.publish(frozen.save, frozen.unchanged), true
		case "back":
			return "", false
		}
		if !s.write(invalidAction + reviewPrompt) {
			return s.failure, true
		}
	}
}

func (s *session) adds() int {
	count := 0
	for _, operation := range s.operations {
		if operation.Kind == editorcontract.Add {
			count++
		}
	}
	return count
}

func (s *session) drop() {
	text, ok := s.ask("Staged operation number:")
	if !ok {
		return
	}
	number, valid := canonical(text, 1, len(s.operations))
	if !valid {
		s.write("Select an existing staged operation number.\n")
		return
	}
	s.operations = append(append([]editorcontract.Operation(nil), s.operations[:number-1]...), s.operations[number:]...)
	s.write("Operation dropped.\n")
}
