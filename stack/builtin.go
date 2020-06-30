package stack

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

var errUnderflow = errors.New("stack underflow")

func Nop(m *Machine) error {
	return nil
}

// Last returns the last top-of-stack value that
// was popped with PopX for use in a calculation.
func (m *Machine) Last() *Value {
	return m.x
}

func (m *Machine) Top() *Value {
	if l := len(m.stack); l > 0 {
		return m.stack[l-1]
	}

	return nil
}

// Pop just removes and returns the top of stack.
func (m *Machine) Pop() *Value {
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
func (m *Machine) PopX() *Value {
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

func (m *Machine) Push(v Value) {
	m.stack = append(m.stack, &v)
}

func (m *Machine) Swap() error {
	l := len(m.stack)

	if l < 2 {
		return errUnderflow
	}

	m.stack = append(m.stack[0:l-2], m.stack[l-1], m.stack[l-2])
	return nil
}

// Dup takes the top-of-stack (x) and duplicates
// it on top of the stack.
func (m *Machine) Dup() error {
	l := len(m.stack)

	if l < 1 {
		return errUnderflow
	}

	m.stack = append(m.stack, m.stack[l-1])
	return nil
}

// Dup2 takes {y,x} and duplicates those two values
// in order on top of the stack.
func (m *Machine) Dup2() error {
	l := len(m.stack)

	if l < 2 {
		return errUnderflow
	}

	m.stack = append(m.stack, m.stack[l-2], m.stack[l-1])
	return nil
}

// Over takes {y,x} and duplicates the second value
// onto the top of stack as {y,x,y}.
func (m *Machine) Over() error {
	l := len(m.stack)

	if l < 2 {
		return errUnderflow
	}

	m.stack = append(m.stack, m.stack[l-2])
	return nil
}

// Roll moves the top-of-stack item onto the bottom,
// exposing the next item (y) as the top.
func (m *Machine) Roll() {
	switch l := len(m.stack); l {
	case 0, 1:
		// nothing to do

	case 2:
		_ = m.Swap()

	default:
		tmp := []*Value{m.stack[l-1]}

		m.stack = append(tmp, m.stack[0:l-1]...)
	}
}

func (m *Machine) SetFixed(d int) {
	m.disp = fixed
	m.digits = d
}

func (m *Machine) SetScientific(d int) {
	m.disp = scientific
	m.digits = d
}

func (m *Machine) SetEngineering(d int) {
	m.disp = engineering
	m.digits = d
}

func (m *Machine) SetRadians() {
	m.mode = radians
}

func (m *Machine) Quit() {
	// TODO - save state if required
	os.Exit(0)
}

func (m *Machine) SaveToFile(fn string) error {
	b, err := m.MarshalJSON()

	if err != nil {
		return fmt.Errorf("dump %w", err)
	}

	err = ioutil.WriteFile(fn, b, 0644)

	if err != nil {
		return fmt.Errorf("save %w", err)
	}

	return nil
}
