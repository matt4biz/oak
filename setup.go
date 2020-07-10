package oak

import (
	"fmt"
	"math"
	"strconv"
)

var (
	// CONSTANTS

	E ExprFunc = func(m *Machine) error {
		m.Push(Value{floater, m.mode, math.E, m})
		return nil
	}

	Pi ExprFunc = func(m *Machine) error {
		m.Push(Value{floater, m.mode, math.Pi, m})
		return nil
	}

	Phi ExprFunc = func(m *Machine) error {
		m.Push(Value{floater, m.mode, math.Phi, m})
		return nil
	}

	// MISCELLANY

	Bye ExprFunc = func(m *Machine) error {
		if m.inter {
			fmt.Fprintln(m.output, "Goodbye")
		}

		return m.Quit()
	}

	ChangeSign ExprFunc = func(m *Machine) error {
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

	Clear ExprFunc = func(m *Machine) error {
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

	ClearAll ExprFunc = func(m *Machine) error {
		m.stack = nil
		m.vars = make(map[string]*Symbol)
		m.x = nil
		m.clearStats()
		return nil
	}

	ClearStack ExprFunc = func(m *Machine) error {
		m.stack = nil
		m.clearStats()
		return nil
	}

	ClearRegs ExprFunc = func(m *Machine) error {
		m.x = nil
		m.clearStats()
		return nil
	}

	ClearVars ExprFunc = func(m *Machine) error {
		m.vars = make(map[string]*Symbol)
		return nil
	}

	Dump ExprFunc = func(m *Machine) error {
		fmt.Fprintln(m.output, "DUMP ========")

		if m.stats != nil {
			fmt.Printf("STAT: %s\n", m.stats)
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

		if len(m.vars) > 0 {
			fmt.Fprintln(m.output)

			for k, v := range m.vars {
				fmt.Printf("V %s: %s\n", k, *v.V)
			}
		}

		if len(m.words) > 0 {
			fmt.Fprintln(m.output)

			for k, w := range m.words {
				fmt.Printf("W %s: %s\n", k, w.Definition())
			}
		}

		fmt.Fprintln(m.output, "======== DUMP")
		return nil
	}

	Show ExprFunc = func(m *Machine) error {
		fmt.Fprintln(m.output, "base:", m.Base(), "mode:", m.Mode(), "display:", m.Display())
		return nil
	}

	Status ExprFunc = func(m *Machine) error {
		return Show(m)
	}

	// STACK OPERATIONS

	Depth ExprFunc = func(m *Machine) error {
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

	Drop ExprFunc = func(m *Machine) error {
		m.Pop()
		return nil
	}

	Dup ExprFunc = func(m *Machine) error {
		return m.Dup()
	}

	Dup2 ExprFunc = func(m *Machine) error {
		return m.Dup2()
	}

	Nop ExprFunc = func(m *Machine) error {
		return nil
	}

	Over ExprFunc = func(m *Machine) error {
		return m.Over()
	}

	Roll ExprFunc = func(m *Machine) error {
		m.Roll()
		return nil
	}

	Swap ExprFunc = func(m *Machine) error {
		return m.Swap()
	}

	Top ExprFunc = func(m *Machine) error {
		m.Top()
		return nil
	}

	// BASE CONVERSION

	SetBase ExprFunc = func(m *Machine) error {
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

	Binary ExprFunc = func(m *Machine) error {
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

	Decimal ExprFunc = func(m *Machine) error {
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

	Hexadecimal ExprFunc = func(m *Machine) error {
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

	Octal ExprFunc = func(m *Machine) error {
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

	// ANGULAR MODE

	SetMode ExprFunc = func(m *Machine) error {
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

	// DISPLAY

	SetEngineering ExprFunc = func(m *Machine) error {
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

	SetFixed ExprFunc = func(m *Machine) error {
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

	SetScientific ExprFunc = func(m *Machine) error {
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
)

func (m *Machine) setBuiltins() {
	m.builtin = map[string]Expr{
		// CONSTANTS

		"e":   E,
		"pi":  Pi,
		"phi": Phi,

		// MISCELLANY

		"bye":    Bye,
		"chs":    ChangeSign,
		"clr":    Clear,
		"clrall": ClearAll,
		"clrstk": ClearStack,
		"clrreg": ClearRegs,
		"clrvar": ClearVars,
		"dump":   Dump,
		"load":   Load,
		"save":   Save,
		"show":   Show,
		"status": Status,

		// STACK OPERATIONS

		"depth": Depth,
		"drop":  Drop,
		"dup":   Dup,
		"dup2":  Dup2,
		"nop":   Nop,
		"over":  Over,
		"roll":  Roll,
		"swap":  Swap,
		"top":   Top,

		// BASE CONVERSION

		"base": SetBase,
		"bin":  Binary,
		"dec":  Decimal,
		"hex":  Hexadecimal,
		"oct":  Octal,

		// ANGULAR MODE

		"mode": SetMode,

		// DISPLAY

		"fix": SetFixed,
		"eng": SetEngineering,
		"sci": SetScientific,

		// STATS

		"sum":  StatsOpAdd,
		"mean": Average,
		"sdev": StdDeviation,
		"line": LinRegression,
		"estm": LinEstimate,

		// ADVANCED MATH

		"ddx":    RunDDX,
		"integr": RunIntegrate,
		"solve":  RunSolve,

		"gaussl": RunGauss,
		"rombrg": RunRomberg,
	}
}
