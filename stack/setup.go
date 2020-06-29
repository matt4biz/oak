package stack

import (
	"fmt"
	"math"
	"strconv"
)

func (m *Machine) SetBuiltins() {
	m.builtin = make(map[string]Expr, 1024)

	// CONSTANTS

	m.builtin["e"] = func(m *Machine) error {
		m.Push(Value{floater, m.mode, math.E, m})
		return nil
	}

	m.builtin["pi"] = func(m *Machine) error {
		m.Push(Value{floater, m.mode, math.Pi, m})
		return nil
	}

	m.builtin["phi"] = func(m *Machine) error {
		m.Push(Value{floater, m.mode, math.Phi, m})
		return nil
	}

	// MISCELLANY

	m.builtin["chs"] = func(m *Machine) error {
		if t := m.Top(); t != nil {
			switch t.T {
			case floater:
				t.V = -t.V.(float64)
				return nil

			case integer:
				t.V = -t.V.(int)
				return nil
			}

			return fmt.Errorf("chs: invalid operand x=%#v", *t)
		}

		return nil
	}

	m.builtin["clr"] = func(m *Machine) error {
		if t := m.Top(); t != nil {
			switch t.T {
			case floater:
				t.V = 0.0
				return nil

			case integer:
				t.V = 0
				return nil

			case stringer:
				t.V = ""
				return nil
			}

			return fmt.Errorf("chs: invalid operand x=%#v", *t)
		}

		return nil
	}

	m.builtin["clrall"] = func(m *Machine) error {
		m.stack = nil
		return nil
	}

	m.builtin["show"] = func(m *Machine) error {
		m.Show()
		return nil
	}

	// STACK OPERATIONS

	m.builtin["drop"] = func(m *Machine) error {
		m.Pop()
		return nil
	}

	m.builtin["dup"] = func(m *Machine) error {
		m.Dup()
		return nil
	}

	m.builtin["dup2"] = func(m *Machine) error {
		m.Dup2()
		return nil
	}

	m.builtin["roll"] = func(m *Machine) error {
		m.Roll()
		return nil
	}

	m.builtin["top"] = func(m *Machine) error {
		m.Dup()
		m.Pop()
		return nil
	}

	m.builtin["swap"] = func(m *Machine) error {
		m.Swap()
		return nil
	}

	m.builtin["depth"] = func(m *Machine) error {
		var v interface{}
		var t tag

		s := float64(len(m.stack))

		if m.base == base10 {
			v = s
			t = floater
		} else {
			v = int(s)
			t = integer
		}

		m.Push(Value{t, m.mode, v, m})
		return nil
	}

	// BASE CONVERSION

	m.builtin["base"] = func(m *Machine) error {
		var b int

		x := m.Pop()

		if x == nil {
			return fmt.Errorf("mode: empty stack")
		}

		switch x.T {
		case floater:
			b = int(x.V.(float64))
		case integer:
			b = x.V.(int)
		case stringer:
			b, _ = strconv.Atoi(x.V.(string))
		}

		switch b {
		case 2:
			m.base = base02
			return nil
		case 8:
			m.base = base08
			return nil
		case 10:
			m.base = base10
			return nil
		case 16:
			m.base = base16
			return nil
		}

		return fmt.Errorf("mode: invalid operand %#v", x.V)
	}

	m.builtin["hex"] = func(m *Machine) error {
		m.base = base16

		if t := m.Top(); t != nil {
			switch t.T {
			case floater:
				t.V = int(t.V.(float64))
				t.T = integer
				return nil

			case integer:
				return nil
			}

			return fmt.Errorf("hex: invalid operand x=%#v", t.V)
		}

		return nil
	}

	m.builtin["bin"] = func(m *Machine) error {
		m.base = base02

		if t := m.Top(); t != nil {
			switch t.T {
			case floater:
				t.V = int(t.V.(float64))
				t.T = integer
				return nil

			case integer:
				return nil
			}

			return fmt.Errorf("bin: invalid operand x=%#v", t.V)
		}

		return nil
	}

	m.builtin["oct"] = func(m *Machine) error {
		m.base = base08

		if t := m.Top(); t != nil {
			switch t.T {
			case floater:
				t.V = int(t.V.(float64))
				t.T = integer
				return nil

			case integer:
				return nil
			}

			return fmt.Errorf("oct: invalid operand x=%#v", t.V)
		}

		return nil
	}

	m.builtin["dec"] = func(m *Machine) error {
		m.base = base10

		if t := m.Top(); t != nil {
			switch t.T {
			case floater:
				return nil

			case integer:
				t.V = float64(t.V.(int))
				t.T = floater
				return nil
			}

			return fmt.Errorf("dec: invalid operand x=%#v", t.V)
		}

		return nil
	}

	// ANGULAR MODE

	m.builtin["mode"] = func(m *Machine) error {
		x := m.Pop()

		if x == nil {
			return fmt.Errorf("mode: empty stack")
		}

		if x.T == stringer {
			switch x.V.(string) {
			case "deg":
				m.mode = degrees
				return nil
			case "rad":
				m.mode = radians
				return nil
			}
		}

		return fmt.Errorf("mode: invalid operand %#v", x.V)
	}

	// FORMATTING

	m.builtin["fix"] = func(m *Machine) error {
		x := m.Pop()

		if x == nil {
			return fmt.Errorf("fix: empty stack")
		}

		switch x.T {
		case floater:
			m.SetFixed(int(x.V.(float64)))
		case integer:
			m.SetFixed(x.V.(int))
		}

		return nil
	}

	m.builtin["sci"] = func(m *Machine) error {
		x := m.Pop()

		if x == nil {
			return fmt.Errorf("sci: empty stack")
		}

		switch x.T {
		case floater:
			m.SetScientific(int(x.V.(float64)))
		case integer:
			m.SetScientific(x.V.(int))
		}

		return nil
	}

	m.builtin["eng"] = func(m *Machine) error {
		x := m.Pop()

		if x == nil {
			return fmt.Errorf("eng: empty stack")
		}

		switch x.T {
		case floater:
			m.SetEngineering(int(x.V.(float64)))
		case integer:
			m.SetEngineering(x.V.(int))
		}

		return nil
	}
}
