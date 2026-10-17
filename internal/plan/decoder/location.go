package decoder

import (
	"strconv"

	"go.yaml.in/yaml/v3"
)

type location struct {
	path   string
	line   int
	column int
}

func rootLocation() location {
	return location{}
}

func (parent location) field(name string, key *yaml.Node) location {
	path := name
	if parent.path != "" {
		path = parent.path + "." + name
	}
	return location{path: path, line: key.Line, column: key.Column}
}

func (parent location) item(index int, item *yaml.Node) location {
	return location{
		path:   parent.path + "[" + strconv.Itoa(index) + "]",
		line:   item.Line,
		column: item.Column,
	}
}
