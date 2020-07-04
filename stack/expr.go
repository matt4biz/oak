package stack

import (
	"fmt"
	"math"
	"math/bits"
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

func BinarySaveOp(op string, f func(float64, float64) float64) Expr {
	return func(m *Machine) error {
		if len(m.stack) < 2 {
			return errUnderflow
		}

		// same as above, but don't pop the y value;
		// used e.g. for the percent/delta percent ops

		// TODO - refactor out the common parts ...

		x := m.PopX()
		y := m.Top()

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

func InverseTrigOp(op string, f func(float64) float64) Expr {
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

			s = f(s)

			if x.M == degrees {
				s *= 180 / math.Pi
			}

			m.Push(m.makeFloatVal(s))
			return nil

		case integer:
			var s = float64(x.V.(uint))

			s = f(s)

			if x.M == degrees {
				s *= 180 / math.Pi
			}

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

func StatsOp(m *Machine) error {
	l := len(m.stack)

	switch {
	case l > 1:
		x := m.stack[l-1]
		y := m.stack[l-2]

		if x.T != floater || y.T != floater {
			return fmt.Errorf("invalid operands y=%#v, x=%#v", y.V, x.V)
		}

		m.x = x
		m.SumXY(x, y)

		n := *m.stats[sumn] // must copy value
		m.stack[l-1] = &n

		return nil

	case l == 1:
		x := m.stack[l-1]

		if x.T != floater {
			return fmt.Errorf("invalid operand x=%#v", x.V)
		}

		m.x = x

		m.SumX(x)

		n := *m.stats[sumn] // must copy value
		m.stack[l-1] = &n

		return nil
	}

	return errUnderflow
}

func Average(m *Machine) error {
	if m.stats == nil || m.stats[sumn].V == nil || m.stats[sumn].V.(float64) == 0 {
		return errNoStats
	}

	n := m.stats[sumn].V.(float64)
	xs := m.stats[xsum].V.(float64)
	ys := m.stats[ysum].V.(float64)

	m.Push(m.makeFloatVal(ys / n))
	m.Push(m.makeFloatVal(xs / n))

	return nil
}

func StdDeviation(m *Machine) error {
	if m.stats == nil || m.stats[sumn].V == nil || m.stats[sumn].V.(float64) == 0 {
		return errNoStats
	}

	n := m.stats[sumn].V.(float64)
	xs := m.stats[xsum].V.(float64)
	ys := m.stats[ysum].V.(float64)
	xsq := m.stats[xsqsum].V.(float64)
	ysq := m.stats[ysqsum].V.(float64)

	sdx := math.Sqrt((n*xsq - xs*xs) / (n * (n - 1)))
	sdy := math.Sqrt((n*ysq - ys*ys) / (n * (n - 1)))

	m.Push(m.makeFloatVal(sdy))
	m.Push(m.makeFloatVal(sdx))

	return nil
}

func LinRegression(m *Machine) error {
	if m.stats == nil || m.stats[sumn].V == nil || m.stats[sumn].V.(float64) == 0 {
		return errNoStats
	}

	n := m.stats[sumn].V.(float64)
	xs := m.stats[xsum].V.(float64)
	ys := m.stats[ysum].V.(float64)
	xsq := m.stats[xsqsum].V.(float64)
	xys := m.stats[xyprod].V.(float64)

	b := (xys - (xs*ys)/n) / (xsq - (xs*xs)/n)
	a := (ys - b*xs) / n

	m.Push(m.makeFloatVal(b))
	m.Push(m.makeFloatVal(a))

	return nil
}

func LinEstimate(m *Machine) error {
	if m.stats == nil || m.stats[sumn].V == nil || m.stats[sumn].V.(float64) == 0 {
		return errNoStats
	}

	if len(m.stack) < 1 {
		return errUnderflow
	}

	x := m.PopX()

	if x.T != floater {
		return fmt.Errorf("invalid operand %#v", x.V)
	}

	xf := x.V.(float64)

	n := m.stats[sumn].V.(float64)
	xs := m.stats[xsum].V.(float64)
	ys := m.stats[ysum].V.(float64)
	xsq := m.stats[xsqsum].V.(float64)
	ysq := m.stats[ysqsum].V.(float64)
	xys := m.stats[xyprod].V.(float64)

	b := (xys - (xs*ys)/n) / (xsq - (xs*xs)/n)
	a := (ys - b*xs) / n
	y := b*xf + a
	r := (xys - (xs*ys)/n) / math.Sqrt((xsq-(xs*xs)/n)*(ysq-(ys*ys)/n))

	m.Push(m.makeFloatVal(r))
	m.Push(m.makeFloatVal(y))

	return nil
}

func Degrees(m *Machine) error {
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

func Radians(m *Machine) error {
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

func ArithmeticShift(y, x uint) uint {
	if x >= 64 {
		return 0
	}

	const high = 1 << 63
	var mask uint

	if y&high == high {
		mask = ^uint(0) << (64 - x)
	}

	return y>>x | mask
}

func MaskLeft(x uint) uint {
	if x >= 64 {
		return ^uint(0)
	}

	return ^uint(0) << (64 - x)
}

func MaskRight(x uint) uint {
	if x >= 64 {
		return ^uint(0)
	}

	return ^uint(0) >> (64 - x)
}

var (
	Add      = BinaryOp("add", func(y, x float64) float64 { return y + x })
	Multiply = BinaryOp("mul", func(y, x float64) float64 { return y * x })
	Subtract = BinaryOp("sub", func(y, x float64) float64 { return y - x })
	Divide   = BinaryOp("div", func(y, x float64) float64 { return y / x })
	Modulo   = BinaryOp("mod", func(y, x float64) float64 { return math.Mod(y, x) })
	Power    = BinaryOp("pow", func(y, x float64) float64 { return math.Pow(y, x) })

	And        = BinaryBitwiseOp("and", func(y, x uint) uint { return y & x })
	Or         = BinaryBitwiseOp("or", func(y, x uint) uint { return y | x })
	Xor        = BinaryBitwiseOp("xor", func(y, x uint) uint { return y ^ x })
	LeftShift  = BinaryBitwiseOp("shl", func(y, x uint) uint { return y << x })
	RightShift = BinaryBitwiseOp("shr", func(y, x uint) uint { return y >> x })
	ArithShift = BinaryBitwiseOp("shr", ArithmeticShift)

	Not = UnaryBitwiseOp("not", func(x uint) uint { return ^x })
)

func Predefined(s string) Expr {
	switch s {
	case "abs":
		return UnaryOp(s, math.Abs)
	case "acos":
		return InverseTrigOp(s, math.Acos)
	case "asin":
		return InverseTrigOp(s, math.Asin)
	case "atan":
		return InverseTrigOp(s, math.Atan)
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
		return BinarySaveOp(s, func(y, x float64) float64 { return (x - y) / y * 100 })
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
	case "maskl":
		return UnaryBitwiseOp(s, MaskLeft)
	case "maskr":
		return UnaryBitwiseOp(s, MaskRight)
	case "max":
		return BinaryOp(s, math.Max)
	case "min":
		return BinaryOp(s, math.Min)
	case "perc":
		return BinarySaveOp(s, func(y, x float64) float64 { return y * x / 100 })
	case "popcnt":
		return UnaryBitwiseOp(s, func(x uint) uint { return uint(bits.OnesCount(x)) })
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
