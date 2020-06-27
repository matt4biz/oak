package expr

import (
	"fmt"
	"math"

	"oak/stack"
)

func BinaryOp(op string, f func(float64, float64) float64) stack.Expr {
	return func(m *stack.Machine) error {
		x := m.PopX()
		y := m.Pop()

		if x == nil || y == nil {
			return fmt.Errorf("%s: empty stack", op)
		}

		if xf, ok := x.(float64); ok {
			if yf, ok := y.(float64); ok {
				s := f(yf, xf)
				m.Push(s)
				return nil
			}
		}

		return fmt.Errorf("%s: mismatched operands y=%#v, x=%#v", op, y, x)
	}
}

func UnaryOp(op string, f func(float64) float64) stack.Expr {
	return func(m *stack.Machine) error {
		x := m.PopX()

		if x == nil {
			return fmt.Errorf("%s: empty stack", op)
		}

		if xf, ok := x.(float64); ok {
			s := f(xf)
			m.Push(s)
			return nil
		}

		return fmt.Errorf("%s: invalid operand x=%#v", op, x)
	}
}

func Predefined(s string) stack.Expr {
	switch s {
	case "abs":
		return UnaryOp(s, math.Abs)
	case "cbrt":
		return UnaryOp(s, math.Cbrt)
	case "ceil":
		return UnaryOp(s, math.Ceil)
	case "cos":
		return UnaryOp(s, math.Cos) // radians
	case "cube":
		return UnaryOp(s, func(x float64) float64 { return x * x * x })
	case "deg":
		return UnaryOp(s, func(x float64) float64 { return x * 180 / math.Pi })
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
		return UnaryOp(s, func(x float64) float64 { return x * math.Pi / 180 })
	case "recp":
		return UnaryOp(s, func(x float64) float64 { return 1 / x })
	case "sin":
		return UnaryOp(s, math.Sin) // radians
	case "sqr":
		return UnaryOp(s, func(x float64) float64 { return x * x })
	case "sqrt":
		return UnaryOp(s, math.Sqrt)
	case "tan":
		return UnaryOp(s, math.Tan) // radians
	case "trunc":
		return UnaryOp(s, math.Trunc)
	}

	return nil
}

func String(s string) stack.Expr {
	return func(m *stack.Machine) error {
		v := m.Lookup(s)

		if v == nil {
			return fmt.Errorf("can't find %s", s)
		}

		m.Push(v.V)
		return nil
	}
}

func GetSymbol(s string) stack.Expr {
	return func(m *stack.Machine) error {
		v := m.Lookup(s)

		if v == nil {
			return fmt.Errorf("can't find %s", s)
		}

		m.Push(v.V)
		return nil
	}
}
