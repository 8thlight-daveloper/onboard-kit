package runner

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

type Scenario struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Source      string `yaml:"source"`
	Steps       []Step `yaml:"steps"`
}

type Step struct {
	Run      []string              `yaml:"run"`
	Files    map[string]FileSource `yaml:"files"`
	Stdin    string                `yaml:"stdin"`
	Talk     []Exchange            `yaml:"talk"`
	Expect   Expectation           `yaml:"expect"`
	Capture  map[string]Capture    `yaml:"capture"`
	Terminal *Terminal             `yaml:"terminal"`
}

func (step Step) IsStdio() bool {
	return step.Talk != nil
}

func (step *Step) UnmarshalYAML(node *yaml.Node) error {
	return decodeFields(node, map[string]func(*yaml.Node) error{
		"run":      func(value *yaml.Node) error { return value.Decode(&step.Run) },
		"files":    func(value *yaml.Node) error { return value.Decode(&step.Files) },
		"stdin":    func(value *yaml.Node) error { return value.Decode(&step.Stdin) },
		"talk":     func(value *yaml.Node) error { return value.Decode(&step.Talk) },
		"expect":   func(value *yaml.Node) error { return value.Decode(&step.Expect) },
		"terminal": func(value *yaml.Node) error { return value.Decode(&step.Terminal) },
		"capture": func(value *yaml.Node) error {
			if value.Kind != yaml.MappingNode || len(value.Content) == 0 {
				return errors.New("Capture schema is invalid.")
			}
			if err := validateYAMLNodes(value); err != nil {
				return errors.New("Capture schema is invalid.")
			}
			return value.Decode(&step.Capture)
		},
	})
}

type FileSource struct {
	Text string
	From string
}

type Exchange struct {
	Send   any
	Expect *Expected
}

type Expected struct {
	Value any
}

type Expectation struct {
	Exit      *int                       `yaml:"exit"`
	Stdout    *OutputExpectation         `yaml:"stdout"`
	Stderr    *OutputExpectation         `yaml:"stderr"`
	Files     map[string]FileExpectation `yaml:"files"`
	Unchanged []string                   `yaml:"unchanged"`
}

type OutputExpectation struct {
	JSON     *Expected
	Contains *string
	Equals   *string
}

type FileExpectation struct {
	SameAs      *string
	JSON        *Expected
	Absent      bool
	ZIP         *zipExpectation
	yamlScalars []yamlScalarExpectation
}

func Load(path string) (Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Scenario{}, err
	}
	scenario, err := Parse(data)
	if err != nil {
		return Scenario{}, fmt.Errorf("%s: %w", path, err)
	}
	fileName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if scenario.Name != fileName {
		return Scenario{}, fmt.Errorf("%s: name %q does not match the file name %q", path, scenario.Name, fileName)
	}
	return scenario, nil
}

func Parse(data []byte) (Scenario, error) {
	scenario, err := parseScenario(data)
	if err != nil && extensionInDocument(data) {
		return Scenario{}, errors.New("Scenario schema is invalid.")
	}
	return scenario, err
}

func parseScenario(data []byte) (Scenario, error) {
	var scenario Scenario
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&scenario); err != nil {
		if errors.Is(err, io.EOF) {
			return Scenario{}, errors.New("the scenario is empty")
		}
		return Scenario{}, err
	}
	if err := validate(scenario); err != nil {
		return Scenario{}, err
	}
	if extensionEnabled(scenario) {
		var document yaml.Node
		parser := yaml.NewDecoder(bytes.NewReader(data))
		if parser.Decode(&document) != nil || validateYAMLNodes(&document) != nil {
			return Scenario{}, errors.New("Scenario schema is invalid.")
		}
		var extra yaml.Node
		if !errors.Is(parser.Decode(&extra), io.EOF) {
			return Scenario{}, errors.New("Scenario schema is invalid.")
		}
	}
	if err := validateCaptures(scenario); err != nil {
		return Scenario{}, err
	}
	return scenario, nil
}

func (expectation *Expectation) UnmarshalYAML(node *yaml.Node) error {
	return decodeFields(node, map[string]func(*yaml.Node) error{
		"exit":      func(value *yaml.Node) error { return value.Decode(&expectation.Exit) },
		"stdout":    func(value *yaml.Node) error { return value.Decode(&expectation.Stdout) },
		"stderr":    func(value *yaml.Node) error { return value.Decode(&expectation.Stderr) },
		"files":     func(value *yaml.Node) error { return value.Decode(&expectation.Files) },
		"unchanged": func(value *yaml.Node) error { return value.Decode(&expectation.Unchanged) },
	})
}

func validate(scenario Scenario) error {
	if scenario.Name == "" {
		return errors.New("name is missing")
	}
	if len(scenario.Steps) == 0 {
		return errors.New("steps is missing")
	}
	for index, step := range scenario.Steps {
		if len(step.Run) == 0 {
			return fmt.Errorf("step %d: run is missing", index+1)
		}
		if step.Terminal != nil {
			if err := validateTerminal(index, step); err != nil {
				return err
			}
		}
		if step.IsStdio() && step.Stdin != "" {
			return fmt.Errorf("step %d: a step with talk cannot also have stdin", index+1)
		}
		for path := range step.Files {
			if !strings.Contains(path, "${") && !filepath.IsLocal(path) {
				return fmt.Errorf("step %d: file %q must stay inside the folder", index+1, path)
			}
		}
	}
	return nil
}

func (source *FileSource) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		source.Text = node.Value
		return nil
	}
	return decodeFields(node, map[string]func(*yaml.Node) error{
		"from": func(value *yaml.Node) error {
			source.From = value.Value
			if source.From == "" {
				return errors.New("from is empty")
			}
			return nil
		},
	})
}

func (exchange *Exchange) UnmarshalYAML(node *yaml.Node) error {
	sent := false
	err := decodeFields(node, map[string]func(*yaml.Node) error{
		"send": func(value *yaml.Node) error {
			sent = true
			converted, err := convertNode(value)
			exchange.Send = converted
			return err
		},
		"expect": func(value *yaml.Node) error {
			expected, err := decodeExpected(value)
			exchange.Expect = expected
			return err
		},
	})
	if err == nil && !sent {
		return fmt.Errorf("line %d: send is missing", node.Line)
	}
	return err
}

func (expectation *OutputExpectation) UnmarshalYAML(node *yaml.Node) error {
	return decodeFields(node, map[string]func(*yaml.Node) error{
		"json": func(value *yaml.Node) error {
			expected, err := decodeExpected(value)
			expectation.JSON = expected
			return err
		},
		"contains": func(value *yaml.Node) error {
			expectation.Contains = &value.Value
			return nil
		},
		"equals": func(value *yaml.Node) error {
			expectation.Equals = &value.Value
			return nil
		},
	})
}

func (expectation *FileExpectation) UnmarshalYAML(node *yaml.Node) error {
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == "yaml-scalars" {
			parsed, err := decodeYAMLScalars(node)
			expectation.yamlScalars = parsed
			return err
		}
	}
	absentHeld := false
	err := decodeFields(node, map[string]func(*yaml.Node) error{
		"same-as": func(value *yaml.Node) error {
			expectation.SameAs = &value.Value
			return nil
		},
		"json": func(value *yaml.Node) error {
			expected, err := decodeExpected(value)
			expectation.JSON = expected
			return err
		},
		"absent": func(value *yaml.Node) error {
			absentHeld = true
			return value.Decode(&expectation.Absent)
		},
		"zip": func(value *yaml.Node) error {
			if expectation.ZIP != nil {
				return errors.New("duplicate zip field")
			}
			parsed, err := decodeZIP(value)
			expectation.ZIP = parsed
			return err
		},
	})
	if err != nil {
		return err
	}
	if expectation.ZIP != nil && (expectation.SameAs != nil || expectation.JSON != nil || absentHeld) {
		return errors.New("zip cannot be combined with same-as, json or absent")
	}
	return nil
}

func decodeFields(node *yaml.Node, handlers map[string]func(*yaml.Node) error) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: expected a mapping", node.Line)
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		key, value := node.Content[index], node.Content[index+1]
		handle, known := handlers[key.Value]
		if !known {
			return fmt.Errorf("line %d: unknown field %q", key.Line, key.Value)
		}
		if err := handle(value); err != nil {
			return fmt.Errorf("line %d: %s: %w", key.Line, key.Value, err)
		}
	}
	return nil
}

func decodeExpected(node *yaml.Node) (*Expected, error) {
	value, err := convertNode(node)
	if err != nil {
		return nil, err
	}
	return &Expected{Value: value}, nil
}

func convertNode(node *yaml.Node) (any, error) {
	switch node.Kind {
	case yaml.AliasNode:
		return convertNode(node.Alias)
	case yaml.SequenceNode:
		items := make([]any, 0, len(node.Content))
		for _, child := range node.Content {
			item, err := convertNode(child)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		return items, nil
	case yaml.MappingNode:
		object := make(map[string]any, len(node.Content)/2)
		for index := 0; index+1 < len(node.Content); index += 2 {
			key := node.Content[index]
			if key.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("line %d: a key must be text", key.Line)
			}
			item, err := convertNode(node.Content[index+1])
			if err != nil {
				return nil, err
			}
			object[key.Value] = item
		}
		return object, nil
	}
	switch node.ShortTag() {
	case "!!null":
		return nil, nil
	case "!!bool":
		var flag bool
		return flag, node.Decode(&flag)
	case "!!int", "!!float":
		var number float64
		err := node.Decode(&number)
		return number, err
	}
	return node.Value, nil
}
