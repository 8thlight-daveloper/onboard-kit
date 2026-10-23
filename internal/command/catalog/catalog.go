package catalog

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	readercontract "onboard-kit/contracts/catalog/reader"
	contract "onboard-kit/contracts/command/catalog"
	"onboard-kit/internal/catalog/reader"
)

const help = "Usage: onboard-kit catalog list\n" +
	"       onboard-kit catalog show --id ID\n" +
	"       onboard-kit catalog export --id ID\n\n" +
	"list prints JSON inventory. show displays complete catalog content.\n" +
	"export prints section YAML or preparation/tool-help Markdown; it writes no files.\n"

type command struct{}

type request struct {
	operation string
	id        string
}

type report struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func New() contract.Command {
	return command{}
}

func (command) Run(args []string, stdout, stderr io.Writer) int {
	if helpRequested(args) {
		if !writeComplete(stdout, []byte(help)) {
			return refuse(stderr, 1, report{
				Code:    "catalog.help-output",
				Message: "The catalog help could not be fully displayed.",
			})
		}
		return 0
	}

	parsed, valid := parse(args)
	if !valid {
		return refuse(stderr, 2, report{
			Code:    "catalog.usage",
			Message: "Use catalog list, show --id ID, or export --id ID.",
		})
	}

	catalog := reader.New()
	var result []byte
	if parsed.operation == "list" {
		encoded, err := json.Marshal(catalog.List())
		if err != nil {
			return resultFailure(stderr)
		}
		result = append(encoded, '\n')
	} else {
		entry, problems := catalog.Read(parsed.id)
		if len(problems) != 0 {
			return refuse(stderr, 1, report{
				Code:    "catalog.unknown-id",
				Message: "Choose an exact ID from catalog list.",
			})
		}
		result = entry.Content
		if parsed.operation == "show" {
			result = show(entry)
		}
	}
	if !writeComplete(stdout, result) {
		return resultFailure(stderr)
	}
	return 0
}

func helpRequested(args []string) bool {
	if len(args) == 0 {
		return false
	}
	if args[0] == "--help" {
		return true
	}
	return recognized(args[0]) && len(args) > 1 && args[1] == "--help"
}

func recognized(operation string) bool {
	return operation == "list" || operation == "show" || operation == "export"
}

func parse(args []string) (request, bool) {
	if len(args) == 0 || !recognized(args[0]) {
		return request{}, false
	}
	parsed := request{operation: args[0]}
	if parsed.operation == "list" {
		return parsed, len(args) == 1
	}
	if len(args) == 3 && args[1] == "--id" {
		parsed.id = args[2]
		return parsed, parsed.id != "" && !strings.HasPrefix(parsed.id, "--")
	}
	if len(args) == 2 && strings.HasPrefix(args[1], "--id=") {
		parsed.id = strings.TrimPrefix(args[1], "--id=")
		return parsed, parsed.id != ""
	}
	return request{}, false
}

func show(entry readercontract.Entry) []byte {
	summary := entry.Summary
	result := []byte(fmt.Sprintf(
		"# %s\n\nID: `%s`\nCarrier: `%s`\nCategory: `%s`\nGoal: %s\nResult: %s\nContent: `%s`\n\n## Content\n\n",
		summary.Title, summary.ID, summary.Carrier, summary.Category,
		summary.Goal, summary.Result, summary.ContentPath,
	))
	if summary.Carrier != readercontract.CarrierSectionTemplate {
		return append(result, entry.Content...)
	}
	fence := strings.Repeat("`", fenceLength(entry.Content))
	result = append(result, fence...)
	result = append(result, "yaml\n"...)
	result = append(result, entry.Content...)
	result = append(result, fence...)
	return append(result, '\n')
}

func fenceLength(content []byte) int {
	length := 3
	run := 0
	for _, character := range content {
		if character != '`' {
			run = 0
			continue
		}
		run++
		if run >= length {
			length = run + 1
		}
	}
	return length
}

func resultFailure(stderr io.Writer) int {
	return refuse(stderr, 1, report{
		Code:    "catalog.result-output",
		Message: "The catalog result could not be fully displayed.",
	})
}

func refuse(stderr io.Writer, status int, problem report) int {
	encoded, err := json.Marshal(problem)
	if err != nil || !writeComplete(stderr, append(encoded, '\n')) {
		return 1
	}
	return status
}

func writeComplete(destination io.Writer, content []byte) bool {
	written, err := destination.Write(content)
	return err == nil && written == len(content)
}
