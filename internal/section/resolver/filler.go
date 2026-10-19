package resolver

import (
	"strings"

	"onboard-kit/internal/model"
)

type filler struct {
	resolutions map[string]resolution
	used        map[string]bool
}

func newFiller(section model.Section, answers map[string]model.Value) *filler {
	resolutions := make(map[string]resolution, len(section.Variables))
	for _, variable := range section.Variables {
		if _, declared := resolutions[variable.Name]; !declared {
			resolutions[variable.Name] = resolveVariable(variable, answers[variable.Name])
		}
	}
	return &filler{resolutions: resolutions, used: map[string]bool{}}
}

func (f *filler) fill(text string) string {
	var filled strings.Builder
	for index := 0; index < len(text); {
		if text[index] != '$' {
			filled.WriteByte(text[index])
			index++
			continue
		}
		index = f.fillDollar(text, index, &filled)
	}
	return filled.String()
}

func (f *filler) fillDollar(text string, dollar int, filled *strings.Builder) int {
	if dollar+1 < len(text) && text[dollar+1] == '$' {
		filled.WriteByte('$')
		return dollar + 2
	}
	end := nameEnd(text, dollar+1)
	result, declared := f.resolutions[text[dollar+1:end]]
	if end == dollar+1 || !declared {
		filled.WriteByte('$')
		return dollar + 1
	}
	f.recordUse(text[dollar+1:end], result)
	filled.WriteString(result.text)
	return end
}

func (f *filler) recordUse(name string, result resolution) {
	if result.gapCode != "" {
		f.used[name] = true
	}
}

func (f *filler) fillCommands(items []string) []string {
	if items == nil {
		return nil
	}
	filled := make([]string, 0, len(items))
	for _, item := range items {
		filled = append(filled, f.fillCommand(item)...)
	}
	return filled
}

func (f *filler) fillCommand(item string) []string {
	if name, exact := exactName(item); exact {
		if commands := f.resolutions[name].commands; commands != nil {
			return commands
		}
	}
	return []string{f.fill(item)}
}

func (f *filler) fillValues(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	filled := make(map[string]string, len(values))
	for key, value := range values {
		filled[key] = f.fill(value)
	}
	return filled
}
