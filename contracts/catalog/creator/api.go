package creator

import "io"

type Command interface {
	Run(args []string, input io.Reader, stdout, stderr io.Writer) int
}
