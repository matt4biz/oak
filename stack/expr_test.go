package stack

import (
	"math"
	"testing"
)

func TestBinaryOp(t *testing.T) {
	m := &Machine{vars: make(map[string]*Symbol)} // we need a place for $0, etc.

	table := []struct {
		name   string
		xt, yt tag
		x, y   interface{}
		ops    []Expr
		base   radix
		want   string
	}{
		{name: "add-f-f-dec", xt: floater, yt: floater, x: 1.0, y: 2.0, ops: []Expr{Add}, want: "3"},
		{name: "add-i-f-dec", xt: integer, yt: floater, x: 1, y: 2.0, ops: []Expr{Add}, want: "3"},
		{name: "add-f-i-dec", xt: floater, yt: integer, x: 1.0, y: 2, ops: []Expr{Add}, want: "3"},
		{name: "add-i-i-dec", xt: integer, yt: integer, x: 1, y: 2, ops: []Expr{Add}, want: "3"},

		{name: "add-f-f-oct", xt: floater, yt: floater, x: 1.0, y: 2.0, ops: []Expr{Add}, base: base08, want: "003"},
		{name: "add-i-f-oct", xt: integer, yt: floater, x: 1, y: 2.0, ops: []Expr{Add}, base: base08, want: "003"},
		{name: "add-f-i-oct", xt: floater, yt: integer, x: 1.0, y: 2, ops: []Expr{Add}, base: base08, want: "003"},
		{name: "add-i-i-oct", xt: integer, yt: integer, x: 1, y: 2, ops: []Expr{Add}, base: base08, want: "003"},

		{name: "sub-f-f-dec", xt: floater, yt: floater, x: 1.0, y: 2.0, ops: []Expr{Subtract}, want: "1"},
		{name: "sub-i-f-dec", xt: integer, yt: floater, x: 1, y: 2.0, ops: []Expr{Subtract}, want: "1"},
		{name: "sub-f-i-dec", xt: floater, yt: integer, x: 1.0, y: 2, ops: []Expr{Subtract}, want: "1"},
		{name: "sub-i-i-dec", xt: integer, yt: integer, x: 1, y: 2, ops: []Expr{Subtract}, want: "1"},

		{name: "sub-f-f-dec-neg", xt: floater, yt: floater, x: 2.0, y: 1.0, ops: []Expr{Subtract}, want: "-1"},
		{name: "sub-i-f-dec-neg", xt: integer, yt: floater, x: 2, y: 1.0, ops: []Expr{Subtract}, want: "-1"},
		{name: "sub-f-i-dec-neg", xt: floater, yt: integer, x: 2.0, y: 1, ops: []Expr{Subtract}, want: "-1"},
		{name: "sub-i-i-dec-neg", xt: integer, yt: integer, x: 2, y: 1, ops: []Expr{Subtract}, want: "-1"},

		{name: "sub-f-f-oct", xt: floater, yt: floater, x: 2.0, y: 3.0, ops: []Expr{Subtract}, base: base08, want: "001"},
		{name: "sub-i-f-oct", xt: integer, yt: floater, x: 2, y: 3.0, ops: []Expr{Subtract}, base: base08, want: "001"},
		{name: "sub-f-i-oct", xt: floater, yt: integer, x: 2.0, y: 3, ops: []Expr{Subtract}, base: base08, want: "001"},
		{name: "sub-i-i-oct", xt: integer, yt: integer, x: 2, y: 3, ops: []Expr{Subtract}, base: base08, want: "001"},

		{name: "sub-f-f-oct-neg", xt: floater, yt: floater, x: 3.0, y: 2.0, ops: []Expr{Subtract}, base: base08, want: "-001"},
		{name: "sub-i-f-oct-neg", xt: integer, yt: floater, x: 3, y: 2.0, ops: []Expr{Subtract}, base: base08, want: "-001"},
		{name: "sub-f-i-oct-neg", xt: floater, yt: integer, x: 3.0, y: 2, ops: []Expr{Subtract}, base: base08, want: "-001"},
		{name: "sub-i-i-oct-neg", xt: integer, yt: integer, x: 3, y: 2, ops: []Expr{Subtract}, base: base08, want: "-001"},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			y := Value{T: tt.yt, V: tt.y, m: m}
			x := Value{T: tt.xt, V: tt.x, m: m}

			m.base = tt.base

			m.Push(y)
			m.Push(x)

			r, err := m.Eval(0, tt.ops)

			if err != nil {
				t.Errorf("add: %v", err)
			}

			if got, ok := r.(string); ok {
				if got != tt.want {
					t.Errorf("add: wanted %v, got %v", tt.want, got)
				}
			} else {
				t.Errorf("add: invalid result %v %[1]T", r)
			}
		})
	}
}

func TestUnaryOp(t *testing.T) {
	m := &Machine{vars: make(map[string]*Symbol)} // we need a place for $0, etc.

	table := []struct {
		name string
		xt   tag
		x    interface{}
		ops  []Expr
		base radix
		want string
	}{
		{name: "log-f-dec", xt: floater, x: 100.0, ops: []Expr{Predefined("log")}, want: "2"},
		{name: "pow-f-dec", xt: floater, x: 3.0, ops: []Expr{Predefined("pow")}, want: "1000"},

		{name: "log-i-hex", xt: integer, x: 100, ops: []Expr{Predefined("log")}, base: base16, want: "0x0002"},
		{name: "pow-i-hex", xt: integer, x: 3, ops: []Expr{Predefined("pow")}, base: base16, want: "0x03e8"},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			x := Value{T: tt.xt, V: tt.x, m: m}

			m.base = tt.base
			m.Push(x)

			r, err := m.Eval(0, tt.ops)

			if err != nil {
				t.Errorf("add: %v", err)
			}

			if got, ok := r.(string); ok {
				if got != tt.want {
					t.Errorf("add: wanted %v, got %v", tt.want, got)
				}
			} else {
				t.Errorf("add: invalid result %v %[1]T", r)
			}
		})
	}
}

func TestTrigonometryOp(t *testing.T) {
	m := &Machine{
		vars:   make(map[string]*Symbol), // we need a place for $0, etc.
		disp:   fixed,
		digits: 3,
	}

	table := []struct {
		name string
		xt   tag
		x    interface{}
		ops  []Expr
		mode mode
		base radix
		want string
	}{
		{name: "sin-f-deg", xt: floater, x: 30.0, ops: []Expr{Predefined("sin")}, want: "0.500"},
		{name: "cos-f-deg", xt: floater, x: 30.0, ops: []Expr{Predefined("cos")}, want: "0.866"},

		{name: "tan-i-oct", xt: integer, x: 45, ops: []Expr{Predefined("tan")}, base: base08, want: "001"},

		{name: "sin-f-rad", xt: floater, x: math.Pi / 6, ops: []Expr{Predefined("sin")}, mode: radians, want: "0.500"},
		{name: "cos-f-rad", xt: floater, x: math.Pi / 6, ops: []Expr{Predefined("cos")}, mode: radians, want: "0.866"},

		{name: "rad-f-deg", xt: floater, x: math.Pi / 6, ops: []Expr{Degrees}, want: "30.000"},
		{name: "deg-f-rad", xt: floater, x: 30.0, ops: []Expr{Radians}, want: "0.524"},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			x := Value{T: tt.xt, V: tt.x, M: tt.mode, m: m}

			m.base = tt.base
			m.mode = tt.mode

			m.Push(x)

			r, err := m.Eval(0, tt.ops)

			if err != nil {
				t.Errorf("add: %v", err)
			}

			if got, ok := r.(string); ok {
				if got != tt.want {
					t.Errorf("add: wanted %v, got %v", tt.want, got)
				}
			} else {
				t.Errorf("add: invalid result %v %[1]T", r)
			}
		})
	}
}
