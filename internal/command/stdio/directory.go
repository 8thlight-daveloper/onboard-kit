package stdio

import (
	"onboard-kit/internal/command/valuejson"
	"onboard-kit/internal/model"
)

func (s *session) setEntry(asked request) map[string]any {
	if len(asked.Entry) == 0 {
		return failure("bad_request", asked.Command+" needs entry")
	}
	var next model.Plan
	var problems []model.Problem
	switch asked.Command {
	case "set-contact":
		entry, err := valuejson.DecodeContact(asked.Entry)
		if err != nil {
			return failure("bad_entry", err.Error())
		}
		next, problems = s.editor.SetContact(s.plan, entry)
	case "set-role":
		entry, err := valuejson.DecodeRole(asked.Entry)
		if err != nil {
			return failure("bad_entry", err.Error())
		}
		next, problems = s.editor.SetRole(s.plan, entry)
	default:
		entry, err := valuejson.DecodeTeam(asked.Entry)
		if err != nil {
			return failure("bad_entry", err.Error())
		}
		next, problems = s.editor.SetTeam(s.plan, entry)
	}
	return s.applied(next, problems)
}

func (s *session) remove(asked request) map[string]any {
	var next model.Plan
	var cleared model.Cleared
	var problems []model.Problem
	switch asked.Kind {
	case "contact":
		next, cleared, problems = s.editor.RemoveContact(s.plan, asked.ID)
	case "role":
		next, cleared, problems = s.editor.RemoveRole(s.plan, asked.ID)
	case "team":
		next, cleared, problems = s.editor.RemoveTeam(s.plan, asked.ID)
	default:
		return failure("bad_request", "remove needs kind contact, role, or team, and id")
	}
	result := s.applied(next, problems)
	result["cleared"] = valuejson.EncodeCleared(cleared)
	return result
}

func (s *session) prune(asked request) map[string]any {
	next, cleared, problems := s.editor.PruneHiddenAnswers(s.plan)
	if !asked.Apply {
		return map[string]any{"ok": true, "applied": false, "problems": valuejson.EncodeProblems(problems), "cleared": valuejson.EncodeCleared(cleared)}
	}
	result := s.applied(next, problems)
	result["cleared"] = valuejson.EncodeCleared(cleared)
	return result
}
