package editguide

import (
	"io"
	"unicode/utf8"
)

const (
	maxInputBytes = 1048576
	maxLineBytes  = 65536
)

func (s *session) line() (string, bool) {
	if s.failure != "" || s.ended {
		return "", false
	}
	data := make([]byte, 0, 128)
	for {
		if s.consumed == maxInputBytes {
			s.failure = "input"
			return "", false
		}
		var one [1]byte
		n, err := s.input.Read(one[:])
		if n < 0 || n > 1 || (err != nil && err != io.EOF) || (n == 0 && err == nil) {
			s.failure = "input"
			return "", false
		}
		s.consumed += n
		if n == 1 && one[0] == '\n' {
			if len(data) > 0 && data[len(data)-1] == '\r' {
				data = data[:len(data)-1]
			}
			if len(data) > maxLineBytes || !utf8.Valid(data) {
				s.failure = "input"
				return "", false
			}
			return string(data), true
		}
		if n == 1 {
			data = append(data, one[0])
			if len(data) > maxLineBytes+1 || (len(data) == maxLineBytes+1 && one[0] != '\r') {
				s.failure = "input"
				return "", false
			}
		}
		if err == io.EOF {
			if len(data) > maxLineBytes || !utf8.Valid(data) {
				s.failure = "input"
			} else {
				s.ended = true
			}
			return "", false
		}
	}
}
