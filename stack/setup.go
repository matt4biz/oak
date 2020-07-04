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
				t.V = uint(-int(t.V.(uint)))
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
		// TODO - clear other non-stack registers
		//   when are defined, e.g. for statistics

		m.stack = nil
		m.vars = make(map[string]*Symbol)
		m.x = nil
		m.stats = nil
		return nil
	}

	m.builtin["clrstk"] = func(m *Machine) error {
		m.stack = nil
		m.stats = nil
		return nil
	}

	m.builtin["clrreg"] = func(m *Machine) error {
		// TODO - clear other non-stack registers
		//   when are defined, e.g. for statistics

		m.x = nil
		m.stats = nil
		return nil
	}

	m.builtin["clrvar"] = func(m *Machine) error {
		m.vars = make(map[string]*Symbol)
		return nil
	}

	m.builtin["status"] = func(m *Machine) error {
		return Show(m)
	}

	// STACK OPERATIONS

	m.builtin["drop"] = func(m *Machine) error {
		m.Pop()
		return nil
	}

	m.builtin["dup"] = func(m *Machine) error {
		return m.Dup()
	}

	m.builtin["dup2"] = func(m *Machine) error {
		return m.Dup2()
	}

	m.builtin["over"] = func(m *Machine) error {
		return m.Over()
	}

	m.builtin["roll"] = func(m *Machine) error {
		m.Roll()
		return nil
	}

	m.builtin["top"] = func(m *Machine) error {
		m.Top()
		return nil
	}

	m.builtin["swap"] = func(m *Machine) error {
		return m.Swap()
	}

	m.builtin["depth"] = func(m *Machine) error {
		var v interface{}
		var t tag

		s := float64(len(m.stack))

		if m.base == base10 {
			v = s
			t = floater
		} else {
			v = uint(s)
			t = integer
		}

		m.Push(Value{t, m.mode, v, m})
		return nil
	}

	// BASE CONVERSION

	m.builtin["base"] = func(m *Machine) error {
		var b uint

		x := m.Pop()

		if x == nil {
			return fmt.Errorf("mode: empty stack")
		}

		switch x.T {
		case floater:
			b = uint(x.V.(float64))
		case integer:
			b = x.V.(uint)
		case stringer:
			i, _ := strconv.Atoi(x.V.(string))
			b = uint(i)
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
				t.V = uint(t.V.(float64))
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
				t.V = uint(t.V.(float64))
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
				t.V = uint(t.V.(float64))
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
				t.V = float64(t.V.(uint))
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
			m.setMode(x.V.(string))
			return nil
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
			m.SetFixed(uint(x.V.(float64)))
		case integer:
			m.SetFixed(x.V.(uint))
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
			m.SetScientific(uint(x.V.(float64)))
		case integer:
			m.SetScientific(x.V.(uint))
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
			m.SetEngineering(uint(x.V.(float64)))
		case integer:
			m.SetEngineering(x.V.(uint))
		}

		return nil
	}

	m.builtin["save"] = Save
	m.builtin["load"] = Load

	m.builtin["sum"] = StatsOp
	m.builtin["mean"] = Average
	m.builtin["sdev"] = StdDeviation
	m.builtin["line"] = LinRegression
	m.builtin["estm"] = LinEstimate

	m.builtin["dump"] = func(m *Machine) error {
		fmt.Println("DUMP ========")

		if m.stats != nil {
			fmt.Printf("STAT: %s\n", m.stats)
		} else {
			fmt.Printf("STAT: <nil>\n")

		}
		if m.x != nil {
			fmt.Printf("LAST: %s\n", *m.x)
		} else {
			fmt.Printf("LAST: <nil>\n")
		}

		for i, l := 0, len(m.stack); i < len(m.stack); i++ {
			l--
			fmt.Printf("ST %d: %s\n", l, *m.stack[i])
		}

		fmt.Println()

		for k, v := range m.vars {
			fmt.Printf("V %s: %s\n", k, *v.V)
		}

		fmt.Println("======== DUMP")
		return nil
	}

	m.builtin["bye"] = func(m *Machine) error {
		if m.inter {
			fmt.Println("Goodbye")
		}

		m.Quit()
		return nil
	}
}
