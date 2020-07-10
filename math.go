package oak

import (
	"fmt"
	"math"
)

const eps = 1e-15

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

// gprep returns a function suitable for Gauss-Legendre integration on [-1, 1].
func gprep(f func(float64) (float64, error), a, b float64) func(float64) (float64, error) {
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
	g := gprep(f, a, b)

	if improper(f, g, a, b) {
		return 0, errImproper
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

// improper makes a simple attempt to find an integral we can't handle, such
// as 1/x or ln x, etc. It's pretty simple and will be easily fooled, but it's
// all that we have for now. It checks both the original and mapped functions.
func improper(f, g func(float64) (float64, error), a, b float64) bool {
	if x, _ := f(a); math.IsNaN(x) || math.IsInf(x, -1) || math.IsInf(x, 1) {
		return true
	}

	if x, _ := f(b); math.IsNaN(x) || math.IsInf(x, -1) || math.IsInf(x, 1) {
		return true
	}

	if x, _ := g(0); math.IsNaN(x) || math.IsInf(x, -1) || math.IsInf(x, 1) {
		return true
	}

	return false
}

// romberg performs Romberg integration over the interval, possibly shifting
// the endpoints slightly to avoid improper integrals at those points.
func romberg(f func(float64) (float64, error), a, b float64) (float64, error) {
	const steps = 24

	R1 := make([]float64, steps)
	R2 := make([]float64, steps)

	// we will switch these two back and forth later

	Rp := R1
	Rc := R2

	// we're going to evaluate the function at both
	// endpoints and adjust them a little if we find
	// the integral to be improper
	//
	// this by no means guarantees the result will be
	// valid, but at least we gave the old college try

	xa, _ := f(a)

	if math.IsNaN(xa) || math.IsInf(xa, -1) || math.IsInf(xa, 1) {
		a += eps
		xa, _ = f(a)
	}

	xb, _ := f(b)

	if math.IsNaN(xb) || math.IsInf(xb, -1) || math.IsInf(xb, 1) {
		b -= eps
		xb, _ = f(b)
	}

	h := b - a

	// calculate the initial value over the entire
	// interval (it's a "prior" row of one element)

	Rp[0] = (xa + xb) * h / 2

	for i := 1; i < steps; i++ {
		h /= 2 // cut the interval in half each time

		c := 0.0
		n := 1 << (i - 1)

		for j := 1; j <= n; j++ {
			x, err := f(a + (2*float64(j)-1)*h)

			if err != nil {
				return 0, err
			}

			c += x
		}

		Rc[0] = h*c + 0.5*Rp[0]

		for j := 1; j <= i; j++ {
			nk := math.Pow(4, float64(j))

			Rc[j] = (nk*Rc[j-1] - Rp[j-1]) / (nk - 1)
		}

		// terminating condition; not quite as strict as the
		// HP-15c which checks two intervals, but then we'd
		// need to keep a third row Rpp

		if i > 1 && math.Abs(Rp[i-1]-Rc[i]) < eps {
			return Rc[i-1], nil
		}

		Rc, Rp = Rp, Rc
	}

	return Rp[steps-1], nil
}

// integrate tries one method and if that fails, tries, tries yet again Mr Kidd.
func integrate(f func(float64) (float64, error), a, b float64) (float64, error) {
	i, err := gauss(f, a, b)

	if err == errImproper {
		i, err = romberg(f, a, b)
	}

	return i, err
}

// ddx uses the three-point approximation to finding the derivative.
func ddx(f func(float64) (float64, error), x float64) (float64, error) {
	const h = 1e-5

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

		if math.Abs(x0-x) < eps {
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

		if math.Abs(x1-x) < eps {
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
	RunDDX = UnaryMathFunc("ddx", ddx)

	RunIntegrate = BinaryMathFunc("integr", integrate)
	RunSolve     = BinaryMathFunc("solve", solve)

	RunGauss   = BinaryMathFunc("gaussl", gauss)
	RunRomberg = BinaryMathFunc("rombrg", romberg)
)

// BinaryMathFunc creates a function from a word by pushing
// and popping from the machine stack, so the math routines
// above don't know about the stack, etc. It expects two
// float values to define the interval, plus the word.
func BinaryMathFunc(name string, mf math2Func) ExprFunc {
	return func(m *Machine) error {
		lastX := m.Last()

		if len(m.stack) < 3 {
			return errUnderflow
		}

		w := m.Pop()
		b := m.Pop()
		a := m.Pop()

		if a.T != floater {
			return fmt.Errorf("%s: invalid operand z=%#v", name, a.V)
		}

		if b.T != floater {
			return fmt.Errorf("%s: invalid operand y=%#v", name, b.V)
		}

		if w.T != word {
			return fmt.Errorf("%s: invalid operand x=%#v", name, w)
		}

		f := func(x float64) (r float64, err error) {
			m.Push(m.makeFloatVal(x))

			if err = w.V.(*Word).Eval(m); err != nil {
				return 0, fmt.Errorf("%s: %s", name, err)
			}

			v := m.Pop()

			if v.T != floater {
				return 0, fmt.Errorf("%s: invalid result %#v", name, v)
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
func UnaryMathFunc(name string, mf math1Func) ExprFunc {
	return func(m *Machine) error {
		lastX := m.Last()

		if len(m.stack) < 2 {
			return errUnderflow
		}

		w := m.Pop()
		a := m.Pop()

		if a.T != floater {
			return fmt.Errorf("%s: invalid operand y=%#v", name, a.V)
		}

		if w.T != word {
			return fmt.Errorf("%s: invalid operand x=%#v", name, w)
		}

		f := func(x float64) (r float64, err error) {
			m.Push(m.makeFloatVal(x))

			if err = w.V.(*Word).Eval(m); err != nil {
				return 0, fmt.Errorf("%s: %s", name, err)
			}

			v := m.Pop()

			if v.T != floater {
				return 0, fmt.Errorf("%s: invalid result %#v", name, v)
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
