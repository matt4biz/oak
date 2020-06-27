package stack

import (
	"fmt"
	"math"
)

// Machine represents a stack-based computational engine
// where all operations take items from the stack and/or
// push items onto the stack. It has a "last x" side
// register as well as a map of variables that aren't
// on the stack.
type Machine struct {
	stack  []interface{}
	x      interface{}
	vars   map[string]*Symbol
	words  map[string]*Word
	built  map[string]Expr
	format string
}

// Symbol represents a variable
type Symbol struct {
	S string
	V interface{}
}

// Word represents a stack-based function (macro).
type Word struct {
	// 	n string
	// 	// TODO
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

	m.Load()
	return &m
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
	}

	s := fmt.Sprintf("$%d", line)
	t := m.Top()

	m.vars[s] = &Symbol{s, t}

	if m.format != "" {
		switch v := t.(type) {
		case float64:
			return fmt.Sprintf(m.format, v), nil
		}
	}

	return t, nil
}

// Last returns the last top-of-stack value that
// was popped with PopX for use in a calculation.
func (m *Machine) Last() interface{} {
	return m.x
}

func (m *Machine) Top() interface{} {
	if l := len(m.stack); l > 0 {
		return m.stack[l-1]
	}

	return nil
}

// Pop just removes and returns the top of stack.
func (m *Machine) Pop() interface{} {
	switch l := len(m.stack); l {
	case 0:
		return nil
	case 1:
		i := m.stack[0]
		m.stack = nil
		return i
	default:
		i := m.stack[l-1]
		m.stack = m.stack[:l-1]
		return i
	}
}

// PopX is used when we're going to push back some
// result involving the TOS, so Last() can return it.
func (m *Machine) PopX() interface{} {
	switch l := len(m.stack); l {
	case 0:
		return nil
	case 1:
		i := m.stack[0]
		m.stack = nil
		m.x = i
		return i
	default:
		i := m.stack[l-1]
		m.stack = m.stack[:l-1]
		m.x = i
		return i
	}
}

func (m *Machine) Push(i interface{}) {
	m.stack = append(m.stack, i)
}

func (m *Machine) Swap() {
	if l := len(m.stack); l > 1 {
		m.stack = append(m.stack[0:l-2], m.stack[l-1], m.stack[l-2])
	}
}

// Dup takes the top-of-stack (x) and duplicates
// it on top of the stack.
func (m *Machine) Dup() {
	if l := len(m.stack); l > 0 {
		m.stack = append(m.stack, m.stack[l-1])
	}
}

// Dup2 takes {y,x} and duplicates those two values
// in order on top of the stack.
func (m *Machine) Dup2() {
	if l := len(m.stack); l > 1 {
		m.stack = append(m.stack, m.stack[l-2], m.stack[l-1])
	}
}

// Roll moves the top-of-stack item onto the bottom,
// exposing the next item (y) as the top.
func (m *Machine) Roll() {
	if l := len(m.stack); l > 2 {
		tmp := []interface{}{m.stack[l-1]}

		m.stack = append(tmp, m.stack[0:l-2])
	}
}

func (m *Machine) SetFixed(d int) {
	m.format = fmt.Sprintf("%%.%df", d)
}

func (m *Machine) SetScientific(d int) {
	m.format = fmt.Sprintf("%%.%de", d)
}

// Lookup takes a variable name and returns the symbol
// which includes its value.
func (m *Machine) Lookup(s string) *Symbol {
	if s == "$0" {
		return &Symbol{s, m.x}
	}

	return m.vars[s]
}

// AddSymbol adds an empty variable to the machine for parsing.
func (m *Machine) AddSymbol(s string) {
	m.vars[s] = &Symbol{s, nil}
}

// Store writes (or overwrites) a given variable name with
// a new value.
func (m *Machine) Store(s *Symbol, v interface{}) {
	m.vars[s.S] = &Symbol{s.S, v}
}

func (m *Machine) Known(s string) bool {
	_, ok := m.words[s]
	return ok
}

func (m *Machine) Word(s string) (Expr, error) {
	return nil, nil
}

func (m *Machine) Builtin(s string) (Expr, error) {
	b, ok := m.built[s]

	if !ok {
		return nil, fmt.Errorf("%s unkown", s)
	}

	return b, nil
}

func (m *Machine) Load() {
	m.built = make(map[string]Expr, 1024)

	// CONSTANTS

	m.built["e"] = func(m *Machine) error {
		m.Push(math.E)
		return nil
	}

	m.built["pi"] = func(m *Machine) error {
		m.Push(math.Pi)
		return nil
	}

	m.built["phi"] = func(m *Machine) error {
		m.Push(math.Phi)
		return nil
	}

	// MISCELLANY

	m.built["chs"] = func(m *Machine) error {
		x := m.Pop()

		if xf, ok := x.(float64); ok {
			m.Push(-xf)
			return nil
		}

		return fmt.Errorf("chs: invalid operand x=%#v", x)
	}

	m.built["clr"] = func(m *Machine) error {
		m.Pop()
		m.Push(0.0)
		return nil
	}

	// STACK OPERATIONS

	m.built["drop"] = func(m *Machine) error {
		m.Pop()
		return nil
	}

	m.built["dup"] = func(m *Machine) error {
		m.Dup()
		return nil
	}

	m.built["dup2"] = func(m *Machine) error {
		m.Dup2()
		return nil
	}

	m.built["roll"] = func(m *Machine) error {
		m.Roll()
		return nil
	}

	m.built["show"] = func(m *Machine) error {
		m.Dup()
		m.Pop()
		return nil
	}

	m.built["swap"] = func(m *Machine) error {
		m.Swap()
		return nil
	}

	m.built["depth"] = func(m *Machine) error {
		m.Push(float64(len(m.stack)))
		return nil
	}

	// BASE CONVERSION

	// ANGULAR MODE

	// FORMATTING

	m.built["fix"] = func(m *Machine) error {
		f := m.Pop()
		if d, ok := f.(float64); ok {
			m.SetFixed(int(d))
		}
		return nil
	}

	m.built["sci"] = func(m *Machine) error {
		f := m.Pop()
		if d, ok := f.(float64); ok {
			m.SetScientific(int(d))
		}
		return nil
	}

	// TODO - eng, which will require custom formatting ...
}
