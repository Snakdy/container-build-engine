package v1

import "errors"

type Pipeline struct {
	Base       string      `json:"base"`
	Statements []Statement `json:"statements"`
	Config     Config      `json:"config"`
}

type Config struct {
	OverwriteEntrypoint bool     `json:"overwrite-entrypoint"`
	Entrypoint          []string `json:"entrypoint"`
	Command             []string `json:"command"`
}

type Statement struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Options   Options  `json:"options"`
	DependsOn []string `json:"depends-on"`
}

type Options map[string]any

type OptionsList []Options

var (
	ErrNoValue   = errors.New("no value found for option")
	ErrWrongType = errors.New("wrong type for option")
)
