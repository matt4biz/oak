package oak

import (
	"os"
	"testing"
)

func TestConstants(t *testing.T) {
	m := New(os.Stdout)

	m.SetFixed(4)

	// PI

	pi, err := m.Builtin("pi")

	if err != nil {
		t.Errorf("invalid pi function: %s", err)
	}

	sPi, err := m.Eval(0, []Expr{pi})

	if err != nil {
		t.Errorf("invalid pi result: %s", err)
	}

	if sPi != "3.1416" {
		t.Errorf("pi is wrong: %s", sPi)
	}

	// PHI

	m.SetScientific(3)

	phi, err := m.Builtin("phi")

	if err != nil {
		t.Errorf("invalid phi function: %s", err)
	}

	sPhi, err := m.Eval(0, []Expr{phi})

	if err != nil {
		t.Errorf("invalid phi result: %s", err)
	}

	if sPhi != "1.618e+00" {
		t.Errorf("phi is wrong: %s", sPhi)
	}

	// EPSILON

	m.SetEngineering(3)

	eps, err := m.Builtin("e")

	if err != nil {
		t.Errorf("invalid e function: %s", err)
	}

	sEps, err := m.Eval(0, []Expr{eps, Number(1000.0), Multiply})

	if err != nil {
		t.Errorf("invalid e result: %s", err)
	}

	if sEps != "2.718e+03" {
		t.Errorf("e is wrong: %s", sEps)
	}
}
