package stack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Settings is used to save internal settings.
type Settings struct {
	Base    radix   `json:"base"`
	Digits  int     `json:"digits"`
	Display display `json:"display_mode"`
	Mode    mode    `json:"trig_mode"`
}

// MachineImage is used to save the machine's state;
// some work must be done to restore the values and
// words from the saved JSON before they can be used.
type MachineImage struct {
	Stack  []*Value           `json:"stack"`
	LastX  *Value             `json:"last"`
	Vars   map[string]*Symbol `json:"vars"`
	Words  map[string]*Word   `json:"words"`
	Status Settings           `json:"status"`
}

func (m *Machine) SaveToFile(fn string) error {
	mi := MachineImage{
		Stack: m.stack,
		LastX: m.x,
		Vars:  m.vars,
		Words: m.words,
		Status: Settings{
			Digits:  m.digits,
			Display: m.disp,
			Base:    m.base,
			Mode:    m.mode,
		},
	}

	b, err := json.Marshal(mi)

	if err != nil {
		return fmt.Errorf("save: %w", err)
	}

	err = ioutil.WriteFile(fn, b, 0644)

	if err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

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

	for _, v := range mi.Stack {
		v.m = m
		m.Push(*v)
	}

	x := mi.LastX
	x.m = m
	m.x = x

	for k, s := range mi.Vars {
		s.V.m = m
		m.vars[k] = s
	}

	// TODO - read & compile words

	m.base = mi.Status.Base
	m.digits = mi.Status.Digits
	m.disp = mi.Status.Display
	m.mode = mi.Status.Mode

	return nil
}
