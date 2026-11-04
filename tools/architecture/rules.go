package main

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

type kindRule struct {
	Imports          []string `yaml:"imports"`
	BannedCategories []string `yaml:"banned-categories"`
	Contract         string   `yaml:"contract"`
}

type componentRule struct {
	Kind    string   `yaml:"kind"`
	Status  string   `yaml:"status"`
	Imports []string `yaml:"imports"`
}

type rules struct {
	Naming struct {
		DepthOneExceptions []string `yaml:"depth-one-exceptions"`
	} `yaml:"naming"`
	Kinds      map[string]kindRule      `yaml:"kinds"`
	Components map[string]componentRule `yaml:"components"`
}

func loadRules(path string) (rules, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return rules{}, err
	}
	var loaded rules
	if err := yaml.Unmarshal(text, &loaded); err != nil {
		return rules{}, fmt.Errorf("%s: %w", path, err)
	}
	return loaded, nil
}
