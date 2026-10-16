package model

type Step struct {
	ID        string
	Name      string
	Notes     string
	Owner     string
	Timing    Timing
	Run       []string
	Open      string
	Directory string
	Env       map[string]string
	Checks    []Check
}

type Check struct {
	Run []string
	Ask string
	Env string
}
