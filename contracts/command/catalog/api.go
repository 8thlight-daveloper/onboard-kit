package catalog

import "io"

type Command interface {
	Run(args []string, stdout, stderr io.Writer) int
}
