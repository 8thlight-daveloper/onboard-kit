package editguide

import (
	"io"
	"strconv"
	"unicode/utf8"

	editing "onboard-kit/contracts/catalog/editor"
	"onboard-kit/internal/model"
)

func (screen *terminal) line() (string, bool) {
	var payload []byte
	for {
		if screen.consumed >= 1048576 || screen.input == nil {
			screen.failure = "input"
			return "", false
		}
		var single [1]byte
		n, err := screen.input.Read(single[:])
		if n < 0 || n > 1 || err != nil && err != io.EOF || n == 0 && err == nil {
			screen.failure = "input"
			return "", false
		}
		screen.consumed += n
		if n == 1 {
			if single[0] == '\n' {
				if len(payload) > 0 && payload[len(payload)-1] == '\r' {
					payload = payload[:len(payload)-1]
				}
				if len(payload) > 65536 || !utf8.Valid(payload) {
					screen.failure = "input"
					return "", false
				}
				return string(payload), true
			}
			payload = append(payload, single[0])
			if len(payload) > 65537 || len(payload) == 65537 && payload[len(payload)-1] != '\r' {
				screen.failure = "input"
				return "", false
			}
		}
		if err == io.EOF {
			if !utf8.Valid(payload) || len(payload) > 65536 {
				screen.failure = "input"
			}
			return "", false
		}
	}
}

func (screen *terminal) finish(stderr io.Writer) int {
	if screen.failure != "" {
		return report(stderr, screen.failure)
	}
	return report(stderr, "cancelled")
}

func (screen *terminal) field(prompt string) (string, bool) {
	if !screen.say(prompt + "\n") {
		return "", false
	}
	return screen.line()
}

func count(text string, maximum int) (int, bool) {
	if text == "" || len(text) > 1 && text[0] == '0' {
		return 0, false
	}
	value := 0
	for _, digit := range text {
		if digit < '0' || digit > '9' {
			return 0, false
		}
		value = value*10 + int(digit-'0')
		if value > maximum {
			return 0, false
		}
	}
	return value, true
}

func (screen *terminal) number(prompt, retry string, maximum int) (int, bool) {
	for {
		text, ok := screen.field(prompt)
		if !ok {
			return 0, false
		}
		if number, valid := count(text, maximum); valid {
			return number, true
		}
		if !screen.say(retry + "\n") {
			return 0, false
		}
	}
}

func (screen *terminal) value(initial bool) (model.Value, bool) {
	for {
		kind, ok := screen.field("Value kind: text, flag, list, or deferred:")
		if !ok {
			return model.Value{}, false
		}
		switch kind {
		case "deferred":
			return model.Value{Kind: model.ValueKindDeferred}, true
		case "text":
			prompt := "Text value (empty text clears an existing answer):"
			if initial {
				prompt = "Text value (empty text is an exact new binding):"
			}
			text, read := screen.field(prompt)
			return model.Value{Kind: model.ValueKindText, Text: text}, read
		case "flag":
			for {
				text, read := screen.field("Flag value: true or false:")
				if !read {
					return model.Value{}, false
				}
				if text == "true" || text == "false" {
					return model.Value{Kind: model.ValueKindFlag, Flag: text == "true"}, true
				}
				if !screen.say("Use true or false.\n") {
					return model.Value{}, false
				}
			}
		case "list":
			length, read := screen.number("List length: 0-4096:", "Use a canonical count from 0 through 4096.", 4096)
			if !read {
				return model.Value{}, false
			}
			items := make([]string, 0, length)
			for i := 1; i <= length; i++ {
				item, complete := screen.field("List item " + strconv.Itoa(i) + ":")
				if !complete {
					return model.Value{}, false
				}
				items = append(items, item)
			}
			return model.Value{Kind: model.ValueKindList, List: items}, true
		default:
			if !screen.say(invalidAction) {
				return model.Value{}, false
			}
		}
	}
}

func (screen *terminal) operation(action string) (editing.Operation, bool) {
	operation := editing.Operation{Kind: editing.Kind(action)}
	var ok bool
	if action == "add" {
		operation.SourceID, ok = screen.field("Source ID:")
		if !ok {
			return operation, false
		}
		operation.SectionID, ok = screen.field("Permanent section ID:")
		if !ok {
			return operation, false
		}
		length, read := screen.number("Initial binding count: 0-64:", "Use a canonical count from 0 through 64.", 64)
		if !read {
			return operation, false
		}
		for i := 0; i < length; i++ {
			name, read := screen.field("Binding name:")
			if !read {
				return operation, false
			}
			value, read := screen.value(true)
			if !read {
				return operation, false
			}
			operation.Bindings = append(operation.Bindings, editing.Binding{Name: name, Value: value})
		}
		return operation, true
	}
	operation.SectionID, ok = screen.field("Instance ID:")
	if !ok {
		return operation, false
	}
	if action == "submit" || action == "unset" {
		operation.Name, ok = screen.field("Variable name:")
		if !ok {
			return operation, false
		}
	}
	if action == "submit" {
		operation.Value, ok = screen.value(false)
	}
	return operation, ok
}
