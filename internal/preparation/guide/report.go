package guide

import (
	"encoding/json"
	"io"
)

type outcome struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func terminal(code string) (outcome, int) {
	result := outcome{Status: "refused", Code: code}
	exit := 1
	switch code {
	case "guide.cancelled":
		result.Status = "cancelled"
		result.Message = "Cancelled. No files changed."
		exit = 0
	case "guide.saved":
		result.Status = "saved"
		result.Message = "Preparation saved."
		exit = 0
	case "guide.unchanged":
		result.Status = "unchanged"
		result.Message = "Preparation unchanged. Original bytes preserved."
		exit = 0
	case "guide.usage":
		result.Message = "Use prepare guide --plan FILE --account ID [--root DIR]."
		exit = 2
	case "guide.source":
		result.Message = "Cannot safely load the selected plan or preparation."
	case "guide.plan":
		result.Message = "The selected plan is invalid."
	case "guide.identity":
		result.Message = "Preparation identity or arrangement IDs are invalid."
	case "guide.path":
		result.Message = "Cannot safely select the preparation destination."
	case "guide.input":
		result.Message = "Cannot read a complete supported input line."
	case "guide.output":
		result.Message = "Cannot display the current draft or review. No files changed."
	case "guide.encode":
		result.Message = "Cannot encode the reviewed preparation within the size limit."
	case "guide.stale-review":
		result.Message = "Source files changed. Reload and review before saving."
	case "guide.commit":
		result.Message = "Preparation was not saved. A commit attempt may have left empty preparation directories."
	case "guide.cleanup-refused":
		result.Message = "Preparation was not saved. Owned temporary cleanup needs repair; empty preparation directories may remain."
	case "guide.cleanup-saved":
		result.Status = "saved"
		result.Message = "Preparation saved. Owned temporary cleanup needs repair."
	case "guide.report-saved":
		result.Status = "saved"
		result.Message = "Preparation saved, but its completion report could not be written."
	}
	return result, exit
}

func report(writer io.Writer, code string) int {
	result, exit := terminal(code)
	data, _ := json.Marshal(result)
	if writeText(writer, string(data)+"\n") {
		return exit
	}
	if code == "guide.saved" {
		fallback, _ := terminal("guide.report-saved")
		data, _ = json.Marshal(fallback)
		writeText(writer, string(data)+"\n")
	}
	return 1
}
