package oak

import (
	"fmt"
	"math"
)

const (
	e = 1e-9
	h = 1e-5
)

type (
	math2Func = func(f func(float64) (float64, error), a, b float64) (y float64, err error)
	math1Func = func(f func(float64) (float64, error), x float64) (y float64, err error)
)

// See Table of the zeros of the Legendre polynomials of order 1-16 and the weight coefficients for
// Gauss’ mechanical quadrature formula, Arnold N. Lowan, Norman Davids, Arthur Levenson,
// Bulletin Amer Math Soc 48 (1942), pp. 739--743.
// https://www.ams.org/journals/bull/1942-48-10/S0002-9904-1942-07771-8/S0002-9904-1942-07771-8.pdf
var legendre = []struct {
	x, a float64
}{
	{-0.949107912342759, 0.129484966168870},
	{-0.741531185599394, 0.279705391489277},
	{-0.405845151377397, 0.381830050505119},
	{0.000000000000000, 0.417959183673469},
	{0.405845151377397, 0.381830050505119},
	{0.741531185599394, 0.279705391489277},
	{0.949107912342759, 0.129484966168870},
}

// prep returns a function suitable for integration on [-1, 1].
func prep(f func(float64) (float64, error), a, b float64) func(float64) (float64, error) {
	if a == -1 && b == 1 {
		return f
	}

	r := b - a
	s := b + a

	return func(t float64) (float64, error) {
		y, err := f((r*t + s) / 2)

		if err != nil {
			return 0, err
		}

		return y * r / 2, nil
	}
}

// gauss runs Gaussian quadrature using a 7th-order Legendre polynomial
// on the interval [-1, 1] using a function modified from the original.
func gauss(f func(float64) (float64, error), a, b float64) (s float64, err error) {
	g := prep(f, a, b)

	if improper(f, g, a, b) {
		return 0, errImproper
	}

	if x, _ := g(0); math.IsNaN(x) || math.IsInf(x, -1) || math.IsInf(x, 1) {
		fmt.Println("g=", x)
		return 0, fmt.Errorf("improper integral (conv)")
	}

	for _, l := range legendre {
		y, err := g(l.x)

		if err != nil {
			return 0, err
		}

		s += l.a * y
	}

	return
}

// improper makes a simple attempt to find an integral we can't handle, such as
// 1/x or 1/(1-x), etc. It's pretty simple and will be easily fooled, but it's
// all we have for now. It checks both the original and mapped functions.
func improper(f, g func(float64) (float64, error), _, _ float64) bool {
	if x, _ := f(0); math.IsNaN(x) || math.IsInf(x, -1) || math.IsInf(x, 1) {
		return true
	}

	if x, _ := g(0); math.IsNaN(x) || math.IsInf(x, -1) || math.IsInf(x, 1) {
		return true
	}

	return false
}

// ddx uses the three-point approximation to finding the derivative.
func ddx(f func(float64) (float64, error), x float64) (float64, error) {
	y1, err := f(x + h)

	if err != nil {
		return 0, err
	}

	y2, err := f(x - h)

	if err != nil {
		return 0, err
	}

	return (y1 - y2) / (2 * h), nil
}

// newton uses the Newton-Raphson method to find a root.
func newton(f func(float64) (float64, error), a, b float64) (x float64, err error) {
	x0 := (b - a) / 2

	for i := 0; i < 100; i++ {
		y, err := f(x0)

		if err != nil {
			return 0, err
		}

		d, err := ddx(f, x0)

		if err != nil {
			return 0, err
		}

		x = x0 - (y / d)

		if math.Abs(x0-x) < e {
			return x, nil
		}

		x0 = x
	}

	return
}

// secant finds a root using the secant method.
func secant(f func(float64) (float64, error), a, b float64) (x float64, err error) {
	x0 := math.Min(a, b)
	x1 := math.Max(a, b)

	for i := 0; i < 100; i++ {
		y0, err := f(x0)

		if err != nil {
			return 0, err
		}

		y1, err := f(x1)

		if err != nil {
			return 0, err
		}

		x = x1 - ((y1 * (x1 - x0)) / (y1 - y0))

		if math.Abs(x1-x) < e {
			return x, nil
		}

		x0, x1 = x1, x
	}

	return
}

// solve uses the secant method by preference because sometimes newton
// will oscillate between two roots, but we need newton as a backup
// because secant will fail for some cases too (hopefully not the
// same ones :-); we always check that the root lies in the interval.
func solve(f func(float64) (float64, error), a, b float64) (x float64, err error) {
	x, err = secant(f, a, b)

	if err != nil {
		return 0, err
	}

	if !math.IsNaN(x) && a <= x && x <= b {
		return x, nil
	}

	x, err = newton(f, a, b)

	if err != nil {
		return 0, err
	}

	y, err := f(x)

	if err != nil {
		return 0, err
	}

	if math.Abs(y) > 1e-9 || x < a || x > b {
		return 0, errNoSolution
	}

	return x, nil
}

var (
	RunDDX = UnaryMathFunc(ddx)

	RunGauss = BinaryMathFunc(gauss)
	RunSolve = BinaryMathFunc(solve)
)

// BinaryMathFunc creates a function from a word by pushing
// and popping from the machine stack, so the math routines
// above don't know about the stack, etc. It expects two
// float values to define the interval, plus the word.
func BinaryMathFunc(mf math2Func) ExprFunc {
	return func(m *Machine) error {
		lastX := m.Last()

		if len(m.stack) < 3 {
			return errUnderflow
		}

		w := m.Pop()
		b := m.Pop()
		a := m.Pop()

		if a.T != floater {
			return fmt.Errorf("solve: invalid operand z=%#v", a.V)
		}

		if b.T != floater {
			return fmt.Errorf("solve: invalid operand y=%#v", b.V)
		}

		if w.T != word {
			return fmt.Errorf("solve: invalid operand x=%#v", w)
		}

		f := func(x float64) (r float64, err error) {
			m.Push(m.makeFloatVal(x))

			if err = w.V.(*Word).Eval(m); err != nil {
				return 0, fmt.Errorf("solve: %s", err)
			}

			v := m.Pop()

			if v.T != floater {
				return 0, fmt.Errorf("solve: invalid result %#v", v)
			}

			r = v.V.(float64)
			return
		}

		s, err := mf(f, a.V.(float64), b.V.(float64))

		if err != nil {
			return err
		}

		m.Push(m.makeFloatVal(s))

		if lastX != nil {
			m.x = lastX
		}

		return nil
	}
}

// UnaryMathFunc creates a function from a word by pushing
// and popping from the machine stack, so the math routines
// above don't know about the stack, etc. It expects one
// float value for the point along with the word itself.
func UnaryMathFunc(mf math1Func) ExprFunc {
	return func(m *Machine) error {
		lastX := m.Last()

		if len(m.stack) < 2 {
			return errUnderflow
		}

		w := m.Pop()
		a := m.Pop()

		if a.T != floater {
			return fmt.Errorf("solve: invalid operand y=%#v", a.V)
		}

		if w.T != word {
			return fmt.Errorf("solve: invalid operand x=%#v", w)
		}

		f := func(x float64) (r float64, err error) {
			m.Push(m.makeFloatVal(x))

			if err = w.V.(*Word).Eval(m); err != nil {
				return 0, fmt.Errorf("solve: %s", err)
			}

			v := m.Pop()

			if v.T != floater {
				return 0, fmt.Errorf("solve: invalid result %#v", v)
			}

			r = v.V.(float64)
			return
		}

		s, err := mf(f, a.V.(float64))

		if err != nil {
			return err
		}

		m.Push(m.makeFloatVal(s))

		if lastX != nil {
			m.x = lastX
		}

		return nil
	}
}
