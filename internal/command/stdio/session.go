package stdio

import (
	"bytes"
	"encoding/json"
	"io"

	"onboard-kit/internal/command/planfile"
	"onboard-kit/internal/command/valuejson"
	"onboard-kit/internal/document/assembler"
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/editor"
	"onboard-kit/internal/plan/validator"
	"onboard-kit/internal/question/finder"
)

type session struct {
	plan      model.Plan
	finder    finder.Finder
	editor    editor.Editor
	validator validator.Validator
}

func newSession(plan model.Plan) *session {
	return &session{plan: plan, finder: finder.New(), editor: editor.New(), validator: validator.New()}
}

type request struct {
	Command  string `json:"command"`
	Question string `json:"question"`
	Section  string `json:"section"`
	Enabled  *bool  `json:"enabled"`
	Path     string `json:"path"`
	Value    any    `json:"value"`
}

func (s *session) handle(line []byte) map[string]any {
	var asked request
	decoder := json.NewDecoder(bytes.NewReader(line))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&asked); err != nil {
		return failure("bad_request", err.Error())
	}
	switch asked.Command {
	case "questions":
		return success("questions", valuejson.EncodeQuestions(s.finder.Questions(s.plan)))
	case "blockers":
		return success("blockers", blockerIDs(s.finder.Blockers(s.plan)))
	case "validate":
		return success("problems", valuejson.EncodeProblems(s.validator.Validate(s.plan)))
	case "submit":
		return s.submit(asked)
	case "toggle":
		return s.toggle(asked)
	case "plan":
		return s.encodedPlan()
	case "assemble":
		document, gaps := assembler.New().Assemble(s.plan)
		return map[string]any{"ok": true, "document": document, "gaps": valuejson.EncodeProblems(gaps)}
	case "save":
		return s.save(asked)
	}
	return failure("unknown_command", "unknown command: "+asked.Command)
}

func (s *session) submit(asked request) map[string]any {
	value, err := valuejson.Decode(asked.Value)
	if err != nil {
		return failure("bad_value", err.Error())
	}
	next, problems := s.editor.Submit(s.plan, asked.Question, value)
	return s.applied(next, problems)
}

func (s *session) toggle(asked request) map[string]any {
	if asked.Enabled == nil {
		return failure("bad_request", "toggle needs section and enabled")
	}
	next, problems := s.editor.SetSectionEnabled(s.plan, asked.Section, *asked.Enabled)
	return s.applied(next, problems)
}

func (s *session) applied(next model.Plan, problems []model.Problem) map[string]any {
	refused := hasError(problems)
	if !refused {
		s.plan = next
	}
	return map[string]any{"ok": true, "applied": !refused, "problems": valuejson.EncodeProblems(problems)}
}

func (s *session) encodedPlan() map[string]any {
	data, problems := encodePlan(s.plan)
	if len(problems) > 0 {
		return failure("encode_failed", "the plan could not be written")
	}
	return map[string]any{"ok": true, "yaml": string(data)}
}

func (s *session) save(asked request) map[string]any {
	if asked.Path == "" {
		return failure("bad_request", "save needs path")
	}
	if !planfile.Save(asked.Path, s.plan, io.Discard) {
		return failure("save_failed", "the plan could not be saved")
	}
	return map[string]any{"ok": true}
}

func encodePlan(plan model.Plan) ([]byte, []model.Problem) {
	var failed []model.Problem
	data, ok := planfile.Encode(plan, io.Discard)
	if !ok {
		failed = append(failed, model.Problem{Code: "encode_failed"})
	}
	return data, failed
}

func hasError(problems []model.Problem) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func blockerIDs(blockers []model.Blocker) []string {
	ids := make([]string, 0, len(blockers))
	for _, blocker := range blockers {
		ids = append(ids, blocker.Question)
	}
	return ids
}

func success(key string, value any) map[string]any {
	return map[string]any{"ok": true, key: value}
}

func failure(code string, message string) map[string]any {
	return map[string]any{"ok": false, "error": map[string]any{"code": code, "message": message}}
}
