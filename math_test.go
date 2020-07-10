package oak

import (
	"fmt"
	"math"
	"testing"
)

func TestGauss(t *testing.T) {
	var probs = []struct {
		a, b float64
		r    string
		f    func(float64) (float64, error)
	}{
		{0, 1.5, "1.085853318e+00", func(x float64) (float64, error) { return math.Exp(x * x / -2), nil }},
		{1, 3, "3.173442467e+02", func(x float64) (float64, error) { x2, x4 := x*x, x*x*x*x; return x4*x2 - x2*math.Sin(2*x), nil }},
		{4, 10, "1.669620000e+05", func(x float64) (float64, error) { return x + 3*x*x + x*x*x*x*x, nil }},
		{-1, 1, "2.000000000e+00", func(x float64) (float64, error) { return x + 3*x*x + x*x*x*x*x, nil }},
	}

	for _, p := range probs {
		i, _ := gauss(p.f, p.a, p.b)

		if f := fmt.Sprintf("%.9e", i); p.r != f {
			t.Errorf("wanted %s, got %s", p.r, f)
		}
	}
}

func TestDerivative(t *testing.T) {
	var probs = []struct {
		x float64
		r string
		f func(float64) (float64, error)
	}{
		{2, "-2.500000000e-01", func(x float64) (float64, error) { return 1 / x, nil }},
		{0, "1.000000000e+00", func(x float64) (float64, error) { return math.Exp(x), nil }},
	}

	for _, p := range probs {
		i, _ := ddx(p.f, p.x)

		if f := fmt.Sprintf("%.9e", i); p.r != f {
			t.Errorf("wanted %s, got %s", p.r, f)
		}
	}
}

func TestSolve(t *testing.T) {
	var probs = []struct {
		a, b float64
		r    string
		f    func(float64) (float64, error)
	}{
		{-1, 1, "6.823278038e-01", func(x float64) (float64, error) { return x*x*x + x - 1, nil }},
		{-1, 1, "9.313225746e-10", func(x float64) (float64, error) { return x * x, nil }},
		{1, 2, "1.334457345e+00", func(x float64) (float64, error) { return 4*x*x*x*x - 6*x*x - 11/4, nil }},
	}

	for _, p := range probs {
		i, _ := solve(p.f, p.a, p.b)

		if f := fmt.Sprintf("%.9e", i); p.r != f {
			t.Errorf("wanted %s, got %s", p.r, f)
		}
	}
}
