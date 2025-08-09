package formatters

import "github.com/thediveo/enumflag/v2"

type Formatter enumflag.Flag

const (
	AutoDetect Formatter = iota
	Term8
	Term16
	Term256
	Term16m
	HTML
	SVG
)

var Formatters = map[Formatter]string{
	AutoDetect: "autodetect",
	Term8:      "terminal",
	Term16:     "terminal16",
	Term256:    "terminal256",
	Term16m:    "terminal16m",
	HTML:       "html",
	SVG:        "svg",
}

var FormatterIds = map[Formatter][]string{
	AutoDetect: {"autodetect", "auto"},
	Term8:      {"terminal", "8"},
	Term16:     {"terminal16", "16"},
	Term256:    {"terminal256", "256"},
	Term16m:    {"terminal16m", "truecolor"},
	HTML:       {"html"},
	SVG:        {"svg"},
}

func (f Formatter) String() string {
	return Formatters[f]
}

func Help() enumflag.Help[Formatter] {
	return enumflag.Help[Formatter]{
		AutoDetect: "choose formatter automatically",
		Term8:      "8‑colour output",
		Term16:     "16‑colour output",
		Term256:    "256‑colour output",
		Term16m:    "true‑colour (24‑bit) output",
		HTML:       "HTML formatter",
		SVG:        "SVG formatter",
	}
}
