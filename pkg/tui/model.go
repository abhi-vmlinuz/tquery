package tui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tquery/tquery/pkg/parser"
)

type ViewMode int

const (
	ViewTable ViewMode = iota
	ViewTree
	ViewJSON
)

type Model struct {
	RawJSON         []byte
	RawData         any
	CurrentData     any
	DataStruct      *parser.DataStructure
	AutoUnwrap      bool
	
	// Components
	TextInput       textinput.Model
	Table           table.Model
	Viewport        viewport.Model
	
	// State
	Query           string
	QueryErr        error
	ViewMode        ViewMode
	ShowInspect     bool
	InspectContent  string
	
	// Dimensions
	Width           int
	Height          int
}

func NewModel(rawJSON []byte, initialQuery string, autoUnwrap bool) (Model, error) {
	parsed, err := parser.Parse(rawJSON, autoUnwrap)
	if err != nil {
		return Model{}, err
	}

	ti := textinput.New()
	ti.Placeholder = "type jq query (e.g. .data[] | {id, owned_by})..."
	ti.SetValue(initialQuery)
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60

	vp := viewport.New(80, 20)

	m := Model{
		RawJSON:     rawJSON,
		RawData:     parsed.Raw,
		CurrentData: parsed.Unwrapped,
		DataStruct:  parsed,
		AutoUnwrap:  autoUnwrap,
		TextInput:   ti,
		Viewport:    vp,
		Query:       initialQuery,
		ViewMode:    ViewTable,
		Width:       80,
		Height:      24,
	}

	m.updateDataStructure()
	return m, nil
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}
