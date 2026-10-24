package starter

import (
	"io"

	"onboard-kit/internal/model"
)

type Command interface {
	Run(args []string, input io.Reader, stdout, stderr io.Writer) int
}

type Runtime struct {
	Clock               func() model.Timestamp
	LookupEnvironment   func(string) string
	WorkingDirectory    func() (string, error)
	UserHomeDirectory   func() (string, error)
	UserConfigDirectory func() (string, error)
}
