package oak

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type (
	tag     uint
	mode    uint
	radix   uint
	display uint
	sreg    uint
)

const (
	floater tag = iota
	integer
	stringer
	symbol
	word
)

const (
	degrees mode = iota
	radians
)

const (
	base10 radix = iota
	base02
	base08
	base16
)

const (
	free display = iota
	fixed
	scientific
	engineering
)

const (
	sumn sreg = iota
	xsum
	ysum
	xsqsum
	ysqsum
	xyprod
	nsreg // total number
)

// Machine represents a stack-based computational engine
// where all operations take items from the stack and/or
// push items onto the stack. It has a "last x" side
// register as well as a map of variables that aren't
// on the stack.
type Machine struct {
	stack   []*Value
	x       *Value
	stats   []*Value
	vars    map[string]*Symbol
	words   map[string]*Word
	builtin map[string]Expr
	output  io.Writer
	autos   string
	digits  uint
	disp    display
	base    radix
	mode    mode
	debug   bool
	inter   bool
}

// Expr represents an expression (operation) that runs
// against the stack.
type Expr interface {
	Eval(m *Machine) error
}

// ExprFunc is a function on the machine state
// (mainly a function from stack -> stack).
type ExprFunc func(m *Machine) error

// Eval allows an ExprFunc to be an Expr without
// needing any other object (e.g., a struct).
func (e ExprFunc) Eval(m *Machine) error {
	return e(m)
}

// New returns a new stack-based machine
// with an empty stack.
func New(o io.Writer) *Machine {
	m := Machine{
		digits: 2,
		vars:   make(map[string]*Symbol, 1024),
		words:  make(map[string]*Word, 1024),
		output: o,
	}

	m.setBuiltins()
	return &m
}

// SetInteractive indicates we're running with
// readline input.
func (m *Machine) SetInteractive() {
	m.inter = true
}

// SetDebugging indicates the machine should
// generate debugging output while evaluating.
func (m *Machine) SetDebugging() {
	m.debug = true
}

// AutoReload loads the machine state from a file if
// the autosave option is set.
func (m *Machine) AutoReload() {
	if m.autos != "" {
		if err := m.LoadFromFile(m.autos + "/.oakimg"); err != nil {
			fmt.Println("unable to reload autosaved state")
		}
	}
}

// AutoSave saves the machine state to a file if the
// option is set.
func (m *Machine) AutoSave() {
	if m.autos != "" {
		if err := m.SaveToFile(m.autos + "/.oakimg"); err != nil {
			fmt.Println("unable to autosave state")
		}
	}
}

// Eval takes a list of expressions (from the parser)
// and applies them; it stops and discards the list if
// any expression results in a failure. Note that the
// operations the stack exports on itself don't return
// errors, only values (possibly nil, stack unchanged)
func (m *Machine) Eval(line int, exprs []Expr) (interface{}, error) {
	for _, e := range exprs {
		if e == nil {
			return nil, fmt.Errorf("found nil expression")
		}

		if err := e.Eval(m); err != nil {
			return nil, err
		}

		if m.debug {
			showStack(m.output, m.stack)
		}
	}

	s := fmt.Sprintf("$%d", line)
	t := m.Top()

	if t == nil {
		return nil, nil
	}

	m.vars[s] = &Symbol{S: s, V: t, result: true}

	return t.String(), nil
}

// Base returns the current base as an int.
func (m *Machine) Base() int {
	switch m.base {
	case base02:
		return 2
	case base08:
		return 8
	case base16:
		return 16
	}

	return 10
}

// Mode returns the current angular mode as a string.
func (m *Machine) Mode() string {
	if m.mode == degrees {
		return deg
	}

	return rad
}

// Display returns the display mode and precision (digits).
func (m *Machine) Display() string {
	switch m.disp {
	case fixed:
		return fmt.Sprintf("fix/%d", m.digits)
	case scientific:
		return fmt.Sprintf("sci/%d", m.digits)
	case engineering:
		return fmt.Sprintf("eng/%d", m.digits)
	}

	return "free"
}

// SetOptions takes an options map (e.g., from a config
// file) and adjusts the machines settings.
func (m *Machine) SetOptions(home string, opts map[string]string) {
	if mode, ok := opts["trig_mode"]; ok {
		m.setMode(mode)
	}

	if base, ok := opts["base"]; ok {
		m.setBase(base)
	}

	if digits, ok := opts["digits"]; ok {
		if d, err := strconv.Atoi(digits); err == nil {
			m.digits = uint(d)
		}
	}

	if display, ok := opts["display_mode"]; ok {
		m.setDisplay(display)
	}

	if auto, ok := opts["autosave"]; ok {
		if strings.ToLower(auto) == "true" {
			m.autos = home
		}
	}
}

func (m *Machine) initStats() {
	zero := m.makeFloatVal(0.0)
	m.stats = make([]*Value, nsreg)

	for i := 0; i < int(nsreg); i++ {
		z := zero
		m.stats[i] = &z

		s := m.makeSymbol(fmt.Sprintf("$r_%d", i+2), m.stats[i])
		m.vars[s.S] = s
	}
}

func (m *Machine) clearStats() {
	m.stats = nil

	for i := 0; i < int(nsreg); i++ {
		delete(m.vars, fmt.Sprintf("r_%d", i+2))
	}
}

func (m *Machine) makeFloatVal(s float64) Value {
	var v interface{}
	var t tag

	if m.base == base10 {
		v = s
		t = floater
	} else {
		v = uint(s)
		t = integer
	}

	return Value{t, m.mode, v, m}
}

func (m *Machine) makeIntVal(i uint) Value {
	return Value{T: integer, M: m.mode, V: i, m: m}
}

func (m *Machine) makeStringVal(s string) Value {
	return Value{T: stringer, V: trimQuotes(s), m: m}
}

func (m *Machine) makeSymbol(s string, v *Value) *Symbol {
	return &Symbol{S: s, V: v, readonly: true}
}

func (m *Machine) makeWord(w *Word) Value {
	return Value{T: word, V: w}
}

func (m *Machine) setMode(s string) {
	switch s {
	case deg:
		m.mode = degrees
	case rad:
		m.mode = radians
	}
}

func (m *Machine) setBase(s string) {
	if base, err := strconv.Atoi(s); err == nil {
		switch base {
		case 2:
			m.base = base02
		case 8:
			m.base = base08
		case 16:
			m.base = base16
		default:
			m.base = base10
		}
	}
}

func (m *Machine) setDisplay(s string) {
	switch s {
	case "fix":
		m.disp = fixed
	case "sci":
		m.disp = scientific
	case "eng":
		m.disp = engineering
	default:
		m.disp = free
	}
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func showStack(w io.Writer, s []*Value) {
	l := len(s)

	fmt.Fprint(w, "[")

	for i := l - 1; i >= 0; i-- {
		fmt.Fprint(w, *s[i], ", ")
	}

	fmt.Fprintln(w, "]")
}
