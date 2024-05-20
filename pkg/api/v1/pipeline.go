package v1

type Pipeline struct {
	Base       string      `json:"base"`
	Statements []Statement `json:"statements"`
}

type Statement struct {
	Name    string  `json:"name"`
	Options Options `json:"options"`
}

type Options map[string]any
