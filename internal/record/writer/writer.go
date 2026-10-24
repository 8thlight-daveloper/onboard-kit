package writer

import (
	contract "onboard-kit/contracts/record/writer"
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/editor"
	"onboard-kit/internal/record/joiner"
)

type Writer struct {
	joiner joiner.Joiner
	editor editor.Editor
}

func New() Writer {
	return Writer{joiner: joiner.New(), editor: editor.New()}
}

func (writer Writer) Write(plan model.Plan, record model.Record, intent contract.Intent) (model.Record, []model.Problem) {
	if problem := checkOperation(intent); problem != nil {
		return model.Record{}, []model.Problem{*problem}
	}
	scope, name, variable, found := findQuestion(plan, intent.Question)
	if !found {
		return refuse(intent.Question, "question", "Choose one declared newcomer question.", "intent.question")
	}
	if variable.Audience != model.AudienceOnboardee {
		return refuse(intent.Question, "audience", "Only newcomer answers can be changed here.", "intent.question")
	}
	if intent.Operation == contract.OperationSubmit && isUnanswered(intent.Value) {
		return refuse(intent.Question, "empty", "Choose an answer, defer, or explicitly follow the plan.", "intent.value")
	}
	joined := writer.joiner.Join(plan, record)
	accepted, problems := writer.editor.Submit(joined, intent.Question, intent.Value)
	problems = copyProblems(problems)
	if hasError(problems) {
		return model.Record{}, problems
	}
	result := copyRecord(record)
	if intent.Operation == contract.OperationFollowPlan {
		removeAnswer(&result.Answers, scope, name)
	} else {
		putAnswer(&result.Answers, scope, name, answerAt(accepted.Answers, scope, name))
	}
	return result, problems
}

func checkOperation(intent contract.Intent) *model.Problem {
	if intent.Operation != contract.OperationSubmit && intent.Operation != contract.OperationFollowPlan {
		problem := ownProblem(intent.Question, "operation", "Unsupported answer operation.", "intent.operation")
		return &problem
	}
	if intent.Operation == contract.OperationFollowPlan && !isZeroValue(intent.Value) {
		problem := ownProblem(intent.Question, "follow-value", "Follow-plan must not carry an answer.", "intent.value")
		return &problem
	}
	return nil
}

func isZeroValue(value model.Value) bool {
	return value.Kind == model.ValueKindNone && value.Text == "" && !value.Flag && value.List == nil
}

func isUnanswered(value model.Value) bool {
	switch value.Kind {
	case model.ValueKindNone:
		return true
	case model.ValueKindText:
		return value.Text == ""
	case model.ValueKindList:
		for _, item := range value.List {
			if item != "" {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func hasError(problems []model.Problem) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func refuse(question, code, message, path string) (model.Record, []model.Problem) {
	return model.Record{}, []model.Problem{ownProblem(question, code, message, path)}
}

func ownProblem(question, code, message, path string) model.Problem {
	return model.Problem{
		Code:     "record.answer." + code,
		Severity: model.SeverityError,
		Message:  message,
		Question: question,
		Detail:   map[string]string{"path": path},
	}
}
