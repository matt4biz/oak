package oak

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Settings is used to save internal settings.
type Settings struct {
	Base    radix   `json:"base"`
	Digits  uint    `json:"digits"`
	Display display `json:"display_mode"`
	Mode    mode    `json:"trig_mode"`
}

// MachineImage is used to save the machine's state;
// some work must be done to restore the values and
// words from the saved JSON before they can be used.
type MachineImage struct {
	Stack  []*Value           `json:"stack,omitempty"`
	LastX  *Value             `json:"last,omitempty"`
	Vars   map[string]*Symbol `json:"vars,omitempty"`
	Words  map[string]*Word   `json:"words,omitempty"`
	Stats  []*Value           `json:"stats,omitempty"`
	Status Settings           `json:"status"`
}

// SaveToFile copies the necessary parts of the machine state
// into something that can be encoded to JSON, and stores that
// JSON representation in a file given the desired filename.
func (m *Machine) SaveToFile(fn string) error {
	mi := MachineImage{
		Stack: m.stack,
		LastX: m.x,
		Words: m.words,
		Stats: m.stats,
		Status: Settings{
			Digits:  m.digits,
			Display: m.disp,
			Base:    m.base,
			Mode:    m.mode,
		},
	}

	// do not save result vars since their line
	// numbers won't line up with a new session

	mi.Vars = make(map[string]*Symbol, len(m.vars))

	for k, v := range m.vars {
		if !v.result {
			mi.Vars[k] = v
		}
	}

	b, err := json.Marshal(mi)

	if err != nil {
		return fmt.Errorf("save: %w", err)
	}

	err = ioutil.WriteFile(fn, b, 0600)

	if err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

// LoadFromFile decodes a JSON representation of a machine
// image, resets the current machine as needed, and then
// copies the decoded state into the machine. Old machine
// state is overwritten except for result variables.
func (m *Machine) LoadFromFile(fn string) error {
	b, err := ioutil.ReadFile(fn)

	if err != nil {
		return fmt.Errorf("load: %w", err)
	}

	var mi MachineImage

	err = json.Unmarshal(b, &mi)

	if err != nil {
		return fmt.Errorf("load: %w", err)
	}

	m.resetForLoad()

	if err != nil {
		return fmt.Errorf("clear: %w", err)
	}

	for _, v := range mi.Stack {
		v.m = m
		m.Push(*v)
	}

	if x := mi.LastX; x != nil {
		x.m = m
		m.x = x
	}

	for k, s := range mi.Vars {
		s.V.m = m
		m.vars[k] = s
	}

	for _, w := range mi.Words {
		// for now, ignore words that don't compile

		if w.Compile(m) != nil {
			continue
		}

		m.Install(w)
	}

	if len(mi.Stats) != 0 {
		m.stats = mi.Stats

		for _, s := range m.stats {
			s.m = m
		}
	}

	m.base = mi.Status.Base
	m.digits = mi.Status.Digits
	m.disp = mi.Status.Display
	m.mode = mi.Status.Mode

	return nil
}

func (m *Machine) resetForLoad() {
	m.stack = nil
	m.stats = nil
	m.words = make(map[string]*Word, 1024)
	m.x = nil

	for k, v := range m.vars {
		if !resultVar.MatchString(v.S) {
			delete(m.vars, k)
		}
	}
}

var (
	// Save pops a filename and saves the machine to that file.
	Save ExprFunc = func(m *Machine) error {
		if len(m.stack) < 1 {
			return errUnderflow
		}

		x := m.Pop()

		if x.T != stringer {
			return fmt.Errorf("save: invalid operand x=%#v", x)
		}

		fn := x.V.(string)

		return m.SaveToFile(fn)
	}

	// Load pops a filename and loads a machine image from that file.
	Load ExprFunc = func(m *Machine) error {
		if len(m.stack) < 1 {
			return errUnderflow
		}

		x := m.Pop()

		if x.T != stringer {
			return fmt.Errorf("save: invalid operand x=%#v", x)
		}

		fn := x.V.(string)

		return m.LoadFromFile(fn)
	}
)
