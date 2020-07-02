package stack

import (
	"fmt"
)

type (
	tag     int
	mode    int
	radix   int
	display int
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

// Machine represents a stack-based computational engine
// where all operations take items from the stack and/or
// push items onto the stack. It has a "last x" side
// register as well as a map of variables that aren't
// on the stack.
type Machine struct {
	stack   []*Value
	x       *Value
	vars    map[string]*Symbol
	words   map[string]*Word
	builtin map[string]Expr
	digits  int
	disp    display
	base    radix
	mode    mode
	debug   bool
	inter   bool
}

// Word represents a stack-based function (macro).
type Word struct {
	N string
	P []Expr
	// TODO - how do we keep def'n for recompile?
}

// Expr represents an expression (operation) that runs
// against the stack.
type Expr func(m *Machine) error

// New returns a new stack-based machine with an empty
// stack.
func New() *Machine {
	m := Machine{
		vars:  make(map[string]*Symbol, 1024),
		words: make(map[string]*Word),
	}

	m.SetBuiltins()
	return &m
}

func (m *Machine) SetInteractive() {
	m.inter = true
}

func (m *Machine) SetDebugging() {
	m.debug = true
}

// Eval takes a list of expressions (from the parser)
// and applies them; it stops and discards the list if
// any expression results in a failure. Note that the
// operations the stack exports on itself don't return
// errors, only values (possibly nil, stack unchanged)
func (m *Machine) Eval(line int, exprs []Expr) (interface{}, error) {
	for _, e := range exprs {
		if err := e(m); err != nil {
			return nil, err
		}

		if m.debug {
			showStack(m.stack)
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

func (m *Machine) Mode() string {
	if m.mode == degrees {
		return "deg"
	}
	return "rad"
}

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
func (m *Machine) Show() {
	fmt.Println("base:", m.Base(), "mode:", m.Mode(), "display:", m.Display())
}

// Put a numerical value onto the stack.
func Number(f float64) Expr {
	return func(m *Machine) error {
		m.Push(m.makeVal(f))
		return nil
	}
}

// Put a numerical value onto the stack.
func Integer(n int) Expr {
	return func(m *Machine) error {
		m.Push(m.makeVal(float64(n)))
		return nil
	}
}

// Put a string value onto the stack.
func String(s string) Expr {
	return func(m *Machine) error {
		m.Push(Value{stringer, m.mode, trimQuotes(s), m})
		return nil
	}
}

func (m *Machine) makeVal(s float64) Value {
	var v interface{}
	var t tag

	if m.base == base10 {
		v = s
		t = floater
	} else {
		v = int(s)
		t = integer
	}

	return Value{t, m.mode, v, m}
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func showStack(s []*Value) {
	l := len(s)

	fmt.Print("[")

	for i := l - 1; i >= 0; i-- {
		fmt.Print(*s[i], ", ")
	}

	fmt.Println("]")
}
