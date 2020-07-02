package stack

import (
	"fmt"
	"math"
)

func BinaryOp(op string, f func(float64, float64) float64) Expr {
	return func(m *Machine) error {
		if len(m.stack) < 2 {
			return errUnderflow
		}

		x := m.PopX()
		y := m.Pop()

		if x == nil || y == nil {
			return fmt.Errorf("%s: empty stack", op)
		}

		switch x.T {
		case floater:
			switch y.T {
			case floater:
				s := f(y.V.(float64), x.V.(float64))
				m.Push(m.makeFloatVal(s))
				return nil

			case integer:
				s := f(float64(y.V.(uint)), x.V.(float64))
				m.Push(m.makeFloatVal(s))
				return nil
			}

		case integer:
			switch y.T {
			case floater:
				s := f(y.V.(float64), float64(x.V.(uint)))
				m.Push(m.makeFloatVal(s))
				return nil

			case integer:
				s := f(float64(y.V.(uint)), float64(x.V.(uint)))
				m.Push(m.makeFloatVal(s))
				return nil
			}
		}

		return fmt.Errorf("%s: mismatched operands y=%#v, x=%#v", op, y.V, x.V)
	}
}

func UnaryOp(op string, f func(float64) float64) Expr {
	return func(m *Machine) error {
		if len(m.stack) < 1 {
			return errUnderflow
		}

		x := m.PopX()

		if x == nil {
			return fmt.Errorf("%s: empty stack", op)
		}

		switch x.T {
		case floater:
			s := f(x.V.(float64))
			m.Push(m.makeFloatVal(s))
			return nil

		case integer:
			s := f(float64(x.V.(uint)))
			m.Push(m.makeFloatVal(s))
			return nil
		}

		return fmt.Errorf("%s: invalid operand x=%#v", op, x.V)
	}
}

func TrigonometryOp(op string, f func(float64) float64) Expr {
	return func(m *Machine) error {
		if len(m.stack) < 1 {
			return errUnderflow
		}

		x := m.PopX()

		if x == nil {
			return fmt.Errorf("%s: empty stack", op)
		}

		switch x.T {
		case floater:
			var s = x.V.(float64)

			if x.M == degrees {
				s *= math.Pi / 180
			}

			s = f(s)
			m.Push(m.makeFloatVal(s))
			return nil

		case integer:
			var s = float64(x.V.(uint))

			if x.M == degrees {
				s *= math.Pi / 180
			}

			s = f(s)
			m.Push(m.makeFloatVal(s))
			return nil
		}

		return fmt.Errorf("%s: invalid operand x=%#v", op, x.V)
	}
}

func BinaryBitwiseOp(op string, f func(y, x uint) uint) Expr {
	return func(m *Machine) error {
		if len(m.stack) < 2 {
			return errUnderflow
		}

		x := m.PopX()
		y := m.Pop()

		if x == nil || y == nil {
			return fmt.Errorf("%s: empty stack", op)
		}

		if x.T == integer && y.T == integer {
			r := f(y.V.(uint), x.V.(uint))

			m.Push(m.makeIntVal(r))
			return nil
		}

		return fmt.Errorf("%s: mismatched operands y=%#v, x=%#v", op, y.V, x.V)
	}
}

func UnaryBitwiseOp(op string, f func(x uint) uint) Expr {
	return func(m *Machine) error {
		if len(m.stack) < 1 {
			return errUnderflow
		}

		x := m.PopX()

		if x == nil {
			return fmt.Errorf("%s: empty stack", op)
		}

		if x.T != integer {
			return fmt.Errorf("%s: invalid operand %#v", op, x)
		}

		r := f(x.V.(uint))

		m.Push(m.makeIntVal(r))
		return nil
	}
}

var Degrees = func(m *Machine) error {
	m.mode = degrees

	if t := m.Top(); t != nil {
		switch t.T {
		case floater:
			t.V = t.V.(float64) * 180 / math.Pi
			t.M = degrees
			return nil

		case integer:
			t.V = int(float64(t.V.(uint)) * 180 / math.Pi)
			t.M = degrees
			return nil
		}

		return fmt.Errorf("hex: invalid operand x=%#v", t.V)
	}

	return nil
}

var Radians = func(m *Machine) error {
	m.mode = radians

	if t := m.Top(); t != nil {
		switch t.T {
		case floater:
			t.V = t.V.(float64) * math.Pi / 180
			t.M = radians
			return nil

		case integer:
			t.V = int(float64(t.V.(uint)) * math.Pi / 180)
			t.M = radians
			return nil
		}

		return fmt.Errorf("hex: invalid operand x=%#v", t.V)
	}

	return nil
}

var (
	Add      = BinaryOp("add", func(y, x float64) float64 { return y + x })
	Multiply = BinaryOp("mul", func(y, x float64) float64 { return y * x })
	Subtract = BinaryOp("sub", func(y, x float64) float64 { return y - x })
	Divide   = BinaryOp("div", func(y, x float64) float64 { return y / x })
	Modulo   = BinaryOp("mod", func(y, x float64) float64 { return math.Mod(y, x) })
	Power    = BinaryOp("pow", func(y, x float64) float64 { return math.Pow(y, x) })

	Not = UnaryBitwiseOp("not", func(x uint) uint { return ^x })
	And = BinaryBitwiseOp("and", func(y, x uint) uint { return y & x })
	Or  = BinaryBitwiseOp("and", func(y, x uint) uint { return y | x })
	Xor = BinaryBitwiseOp("and", func(y, x uint) uint { return y ^ x })
)

func Predefined(s string) Expr {
	switch s {
	case "abs":
		return UnaryOp(s, math.Abs)
	case "cbrt":
		return UnaryOp(s, math.Cbrt)
	case "ceil":
		return UnaryOp(s, math.Ceil)
	case "cos":
		return TrigonometryOp(s, math.Cos)
	case "cube":
		return UnaryOp(s, func(x float64) float64 { return x * x * x })
	case "deg":
		return Degrees
	case "dist":
		return BinaryOp(s, math.Hypot)
	case "dperc":
		return BinaryOp(s, func(y, x float64) float64 { return (x - y) / y * 100 })
	case "exp":
		return UnaryOp(s, math.Exp)
	case "fact":
		return UnaryOp(s, func(x float64) float64 { return math.Gamma(x + 1) })
	case "floor":
		return UnaryOp(s, math.Floor)
	case "frac":
		return UnaryOp(s, func(x float64) float64 { return x - math.Trunc(x) })
	case "ln":
		return UnaryOp(s, math.Log)
	case "log":
		return UnaryOp(s, math.Log10)
	case "max":
		return BinaryOp(s, math.Max)
	case "min":
		return BinaryOp(s, math.Min)
	case "perc":
		return BinaryOp(s, func(y, x float64) float64 { return y * x / 100 })
	case "pow":
		return UnaryOp(s, func(x float64) float64 { return math.Pow(10, x) })
	case "rad":
		return Radians
	case "recp":
		return UnaryOp(s, func(x float64) float64 { return 1 / x })
	case "sin":
		return TrigonometryOp(s, math.Sin)
	case "sqr":
		return UnaryOp(s, func(x float64) float64 { return x * x })
	case "sqrt":
		return UnaryOp(s, math.Sqrt)
	case "tan":
		return TrigonometryOp(s, math.Tan)
	case "trunc":
		return UnaryOp(s, math.Trunc)
	}

	return nil
}
