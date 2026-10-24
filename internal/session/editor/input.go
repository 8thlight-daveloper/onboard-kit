package editor

import (
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"

	reviewcontract "onboard-kit/contracts/session/reviewer"
	"onboard-kit/internal/model"
)

type lineReader struct {
	reader   io.Reader
	consumed int
}

func (r *lineReader) line() (string, bool, bool) {
	if r.reader == nil {
		return "", false, false
	}
	data := make([]byte, 0, 128)
	var one [1]byte
	for {
		if r.consumed >= 1048576 {
			return "", false, false
		}
		n, err := r.reader.Read(one[:])
		if n < 0 || n > 1 {
			return "", false, false
		}
		if n == 1 {
			r.consumed++
			if err != nil && err != io.EOF {
				return "", false, false
			}
			if one[0] == '\n' {
				if len(data) > 0 && data[len(data)-1] == '\r' {
					data = data[:len(data)-1]
				}
				return string(data), true, len(data) <= 65536 && utf8.Valid(data)
			}
			data = append(data, one[0])
			if len(data) > 65537 || len(data) == 65537 && one[0] != '\r' {
				return "", false, false
			}
		}
		if err == io.EOF {
			return "", false, len(data) <= 65536 && utf8.Valid(data)
		}
		if err != nil || n == 0 {
			return "", false, false
		}
	}
}

func parseIntent(line string) (reviewcontract.Intent, bool) {
	result := reviewcontract.Intent{}
	if strings.HasPrefix(line, "follow-plan ") {
		result.Operation = reviewcontract.OperationFollowPlan
		result.Question = strings.TrimPrefix(line, "follow-plan ")
		return result, token(result.Question)
	}
	if strings.HasPrefix(line, "answer ") {
		rest := line[len("answer "):]
		question, tail, ok := strings.Cut(rest, " ")
		if !ok || !token(question) {
			return result, false
		}
		kind, payload, split := strings.Cut(tail, " ")
		result.Operation = reviewcontract.OperationAnswer
		result.Question = question
		switch kind {
		case "deferred":
			if split {
				return result, false
			}
			result.Value.Kind = model.ValueKindDeferred
		case "text":
			if !split || !jsonString(payload, &result.Value.Text) {
				return result, false
			}
			result.Value.Kind = model.ValueKindText
		case "flag":
			if !split || payload != "true" && payload != "false" {
				return result, false
			}
			result.Value = model.Value{Kind: model.ValueKindFlag, Flag: payload == "true"}
		case "list":
			if !split || !validJSONUnicode(payload) || len(payload) < 2 || payload[0] != '[' || payload[len(payload)-1] != ']' {
				return result, false
			}
			var items []json.RawMessage
			if json.Unmarshal([]byte(payload), &items) != nil {
				return result, false
			}
			result.Value.Kind = model.ValueKindList
			result.Value.List = make([]string, len(items))
			for i, item := range items {
				if !jsonString(string(item), &result.Value.List[i]) {
					return reviewcontract.Intent{}, false
				}
			}
		default:
			return result, false
		}
		return result, true
	}
	if strings.HasPrefix(line, "outcome ") {
		section, tail, ok := strings.Cut(line[len("outcome "):], " ")
		if !ok || !token(section) {
			return result, false
		}
		kind, tail, ok := strings.Cut(tail, " ")
		if !ok || kind != "verified" && kind != "failed" && kind != "deferred" {
			return result, false
		}
		end := jsonStringEnd(tail)
		if end < 0 || end+1 >= len(tail) || tail[end] != ' ' {
			return result, false
		}
		result = reviewcontract.Intent{Operation: reviewcontract.OperationOutcome, Section: section, Kind: model.OutcomeKind(kind)}
		if !jsonString(tail[:end], &result.Note) || !jsonString(tail[end+1:], &result.Differently) {
			return reviewcontract.Intent{}, false
		}
		return result, true
	}
	return result, false
}
func token(s string) bool {
	return s != "" && utf8.ValidString(s) && !strings.ContainsAny(s, " \t\r\n\x00")
}
func jsonString(s string, out *string) bool {
	return len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' && validJSONUnicode(s) && json.Unmarshal([]byte(s), out) == nil
}
func jsonStringEnd(s string) int {
	if len(s) == 0 || s[0] != '"' {
		return -1
	}
	escaped := false
	for i := 1; i < len(s); i++ {
		if escaped {
			escaped = false
			continue
		}
		if s[i] == '\\' {
			escaped = true
			continue
		}
		if s[i] == '"' {
			return i + 1
		}
	}
	return -1
}
func validJSONUnicode(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}
	inString, escaped := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !inString {
			if c == '"' {
				inString = true
			}
			continue
		}
		if escaped {
			escaped = false
			if c != 'u' {
				continue
			}
			if i+4 >= len(s) {
				return false
			}
			n, ok := hexFour(s[i+1 : i+5])
			if !ok {
				return false
			}
			i += 4
			if n >= 0xdc00 && n <= 0xdfff {
				return false
			}
			if n >= 0xd800 && n <= 0xdbff {
				if i+6 >= len(s) || s[i+1:i+3] != "\\u" {
					return false
				}
				next, ok := hexFour(s[i+3 : i+7])
				if !ok || next < 0xdc00 || next > 0xdfff {
					return false
				}
				i += 6
			}
			continue
		}
		if c == '\\' {
			escaped = true
		} else if c == '"' {
			inString = false
		}
	}
	return !inString && !escaped
}
func hexFour(s string) (int, bool) {
	n := 0
	for _, c := range s {
		n *= 16
		switch {
		case c >= '0' && c <= '9':
			n += int(c - '0')
		case c >= 'a' && c <= 'f':
			n += int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			n += int(c-'A') + 10
		default:
			return 0, false
		}
	}
	return n, true
}

func writeBlock(w io.Writer, s string) bool {
	if w == nil {
		return false
	}
	n, err := w.Write([]byte(s))
	if err != nil || n != len(s) {
		return false
	}
	if f, ok := w.(interface{ Flush() error }); ok {
		return f.Flush() == nil
	}
	return true
}
