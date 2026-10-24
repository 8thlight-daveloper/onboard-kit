package editor

import "io"

type Driver interface {
	Run(arguments []string, streams Streams) Result
}

type Streams struct {
	Input  io.Reader
	Output io.Writer
	Error  io.Writer
}

type Result struct {
	Exit  int
	Saved bool
}

type Clock interface {
	Now() string
}
