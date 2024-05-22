package v1

type Pipeline struct {
	Base       string      `json:"base"`
	Statements []Statement `json:"statements"`
}

type Statement struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Options   Options  `json:"options"`
	DependsOn []string `json:"depends-on"`
}

type Options map[string]any
