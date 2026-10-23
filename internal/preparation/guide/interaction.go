package guide

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"onboard-kit/internal/model"
)

type lineReader struct {
	reader   io.Reader
	consumed int
	ended    bool
}

var errInput = errors.New("unsupported input")

func (r *lineReader) line() (string, error) {
	data := make([]byte, 0, 128)
	emptyReads := 0
	for {
		if r.consumed == byteLimit {
			return "", errInput
		}
		if r.ended {
			return incomplete(data)
		}
		var next [1]byte
		n, err := r.reader.Read(next[:])
		if n < 0 || n > 1 {
			return "", errInput
		}
		if n == 1 {
			r.consumed++
			emptyReads = 0
			if err != nil && err != io.EOF {
				return "", errInput
			}
			if err == io.EOF {
				r.ended = true
			}
			if next[0] == '\n' {
				if len(data) > 0 && data[len(data)-1] == '\r' {
					data = data[:len(data)-1]
				}
				if len(data) > 65536 || !utf8.Valid(data) {
					return "", errInput
				}
				return string(data), nil
			}
			data = append(data, next[0])
			if len(data) > 65537 || len(data) == 65537 && next[0] != '\r' {
				return "", errInput
			}
			if r.consumed == byteLimit {
				return "", errInput
			}
			if err == io.EOF {
				return incomplete(data)
			}
		} else {
			if err == io.EOF {
				r.ended = true
				return incomplete(data)
			}
			if err != nil {
				return "", errInput
			}
			emptyReads++
			if emptyReads == 100 {
				return "", errInput
			}
		}
	}
}

func incomplete(data []byte) (string, error) {
	if len(data) > 65536 || !utf8.Valid(data) {
		return "", errInput
	}
	return "", io.EOF
}

func writeText(writer io.Writer, value string) bool {
	n, err := io.WriteString(writer, value)
	return err == nil && n == len(value)
}

func (s *session) line(value string) {
	if s.problem == "" && !writeText(s.output, value+"\n") {
		s.problem = "guide.output"
	}
}

func (s *session) ask(prompt string) (string, bool) {
	s.line(prompt)
	if !s.active() {
		return "", false
	}
	value, err := s.input.line()
	if err == io.EOF {
		s.cancelled = true
		return "", false
	}
	if err != nil {
		s.problem = "guide.input"
		return "", false
	}
	return value, true
}

func (s *session) invalid() { s.line("Choose one of the displayed actions.") }

func quote(value string) string {
	var result strings.Builder
	result.WriteByte('"')
	for _, char := range value {
		switch char {
		case '"':
			result.WriteString("\\\"")
		case '\\':
			result.WriteString("\\\\")
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		default:
			if unicode.IsControl(char) || unicode.Is(unicode.Cf, char) || char == '\u2028' || char == '\u2029' {
				if char > 65535 {
					high, low := utf16.EncodeRune(char)
					escapeRune(&result, high)
					escapeRune(&result, low)
				} else {
					escapeRune(&result, char)
				}
			} else {
				result.WriteRune(char)
			}
		}
	}
	result.WriteByte('"')
	return result.String()
}

func escapeRune(result *strings.Builder, char rune) {
	hex := strconv.FormatInt(int64(char), 16)
	result.WriteString("\\u" + strings.Repeat("0", 4-len(hex)) + hex)
}

func (s *session) problems(problems []model.Problem) {
	if len(problems) == 0 {
		s.line("None.")
	}
	for _, problem := range problems {
		s.line("Severity: " + quote(string(problem.Severity)) + " Code: " + quote(problem.Code) + " Path: " + quote(problem.Detail["path"]))
	}
}

func ordinal(value string, count int) (int, bool) {
	if value == "" || value[0] < '1' || value[0] > '9' {
		return 0, false
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return 0, false
		}
	}
	number, err := strconv.Atoi(value)
	return number - 1, err == nil && number >= 1 && number <= count
}

func (s *session) selection(prompt string, count int) (int, bool) {
	for s.active() {
		value, ok := s.ask(prompt)
		if !ok || value == "back" {
			return 0, false
		}
		if index, valid := ordinal(value, count); valid {
			return index, true
		}
		s.invalid()
	}
	return 0, false
}

func (s *session) value(prompt string) (string, bool) {
	for s.active() {
		value, ok := s.ask(prompt)
		if !ok {
			return "", false
		}
		if value != "" {
			return value, true
		}
	}
	return "", false
}

func (s *session) scalar(label, held string, choices func()) string {
	s.line("Field " + label + ": " + quote(held))
	for s.active() {
		action, ok := s.ask("Choose keep, change, clear, or back:")
		if !ok {
			return held
		}
		switch action {
		case "keep", "back":
			return held
		case "clear":
			return ""
		case "change":
			if choices != nil {
				choices()
			}
			if value, valid := s.value("Enter the new single-line value (empty input retries):"); valid {
				return value
			}
			return held
		default:
			s.invalid()
		}
	}
	return held
}

func (s *session) contacts() {
	s.line("Available contacts:")
	if len(s.plan.Directory.Contacts) == 0 {
		s.line("No declared contacts.")
	}
	for _, contact := range s.plan.Directory.Contacts {
		s.line("ID: " + quote(contact.ID) + " Name: " + quote(contact.Name))
	}
}

func (s *session) teams() {
	s.line("Available teams:")
	if len(s.plan.Directory.Teams) == 0 {
		s.line("No declared teams.")
	}
	for _, team := range s.plan.Directory.Teams {
		s.line("ID: " + quote(team.ID) + " Name: " + quote(team.Name))
	}
}

func (s *session) sections() {
	s.line("Available sections:")
	if len(s.plan.Sections) == 0 {
		s.line("No declared sections.")
	}
	for _, section := range s.plan.Sections {
		s.line("ID: " + quote(section.ID) + " Title: " + quote(section.Title))
	}
}

func (s *session) kinds() {
	s.line("Supported kinds:")
	for _, kind := range []string{"support", "equipment", "office", "remote", "access", "meeting", "reading", "pairing", "ticket", "pr-review", "pr-post", "pr-merge", "deploy-shadow", "deploy-lead"} {
		s.line(kind)
	}
}

func (s *session) statuses() {
	s.line("Supported preparation statuses:")
	for _, status := range []string{"pending", "requested", "arranged", "deferred"} {
		s.line(status)
	}
}
