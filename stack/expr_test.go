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
		{name: "add-i-f-dec", xt: integer, yt: floater, x: uint(1), y: 2.0, ops: []Expr{Add}, want: "3"},
		{name: "add-f-i-dec", xt: floater, yt: integer, x: 1.0, y: uint(2), ops: []Expr{Add}, want: "3"},
		{name: "add-i-i-dec", xt: integer, yt: integer, x: uint(1), y: uint(2), ops: []Expr{Add}, want: "3"},

		{name: "add-f-f-oct", xt: floater, yt: floater, x: 1.0, y: 2.0, ops: []Expr{Add}, base: base08, want: "003"},
		{name: "add-i-f-oct", xt: integer, yt: floater, x: uint(1), y: 2.0, ops: []Expr{Add}, base: base08, want: "003"},
		{name: "add-f-i-oct", xt: floater, yt: integer, x: 1.0, y: uint(2), ops: []Expr{Add}, base: base08, want: "003"},
		{name: "add-i-i-oct", xt: integer, yt: integer, x: uint(1), y: uint(2), ops: []Expr{Add}, base: base08, want: "003"},

		{name: "sub-f-f-dec", xt: floater, yt: floater, x: 1.0, y: 2.0, ops: []Expr{Subtract}, want: "1"},
		{name: "sub-i-f-dec", xt: integer, yt: floater, x: uint(1), y: 2.0, ops: []Expr{Subtract}, want: "1"},
		{name: "sub-f-i-dec", xt: floater, yt: integer, x: 1.0, y: uint(2), ops: []Expr{Subtract}, want: "1"},
		{name: "sub-i-i-dec", xt: integer, yt: integer, x: uint(1), y: uint(2), ops: []Expr{Subtract}, want: "1"},

		{name: "sub-f-f-dec-neg", xt: floater, yt: floater, x: 2.0, y: 1.0, ops: []Expr{Subtract}, want: "-1"},
		{name: "sub-i-f-dec-neg", xt: integer, yt: floater, x: uint(2), y: 1.0, ops: []Expr{Subtract}, want: "-1"},
		{name: "sub-f-i-dec-neg", xt: floater, yt: integer, x: 2.0, y: uint(1), ops: []Expr{Subtract}, want: "-1"},
		{name: "sub-i-i-dec-neg", xt: integer, yt: integer, x: uint(2), y: uint(1), ops: []Expr{Subtract}, want: "-1"},

		{name: "sub-f-f-oct", xt: floater, yt: floater, x: 2.0, y: 3.0, ops: []Expr{Subtract}, base: base08, want: "001"},
		{name: "sub-i-f-oct", xt: integer, yt: floater, x: uint(2), y: 3.0, ops: []Expr{Subtract}, base: base08, want: "001"},
		{name: "sub-f-i-oct", xt: floater, yt: integer, x: 2.0, y: uint(3), ops: []Expr{Subtract}, base: base08, want: "001"},
		{name: "sub-i-i-oct", xt: integer, yt: integer, x: uint(2), y: uint(3), ops: []Expr{Subtract}, base: base08, want: "001"},

		{name: "sub-f-f-oct-neg", xt: floater, yt: floater, x: 3.0, y: 2.0, ops: []Expr{Subtract}, base: base08, want: "01777777777777777777777"},
		{name: "sub-i-f-oct-neg", xt: integer, yt: floater, x: uint(3), y: 2.0, ops: []Expr{Subtract}, base: base08, want: "01777777777777777777777"},
		{name: "sub-f-i-oct-neg", xt: floater, yt: integer, x: 3.0, y: uint(2), ops: []Expr{Subtract}, base: base08, want: "01777777777777777777777"},
		{name: "sub-i-i-oct-neg", xt: integer, yt: integer, x: uint(3), y: uint(2), ops: []Expr{Subtract}, base: base08, want: "01777777777777777777777"},
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

		{name: "log-i-hex", xt: integer, x: uint(100), ops: []Expr{Predefined("log")}, base: base16, want: "0x0002"},
		{name: "pow-i-hex", xt: integer, x: uint(3), ops: []Expr{Predefined("pow")}, base: base16, want: "0x03e8"},

		{name: "not-i-bin", xt: integer, x: uint(0b01101001), ops: []Expr{Not}, base: base02, want: "0b1111111111111111111111111111111111111111111111111111111110010110"},
		{name: "maskl", xt: integer, x: uint(3), ops: []Expr{Predefined("maskl")}, base: base16, want: "0xe000000000000000"},
		{name: "maskr", xt: integer, x: uint(6), ops: []Expr{Predefined("maskr")}, base: base16, want: "0x003f"},
		{name: "mask65", xt: integer, x: uint(65), ops: []Expr{Predefined("maskr")}, base: base16, want: "0xffffffffffffffff"},
		{name: "popcnt", xt: integer, x: uint(65535), ops: []Expr{Predefined("popcnt")}, base: base16, want: "0x0010"},
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

		{name: "tan-i-oct", xt: integer, x: uint(45), ops: []Expr{Predefined("tan")}, base: base08, want: "001"},

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

func TestBitwiseOp(t *testing.T) {
	m := &Machine{
		vars: make(map[string]*Symbol), // we need a place for $0, etc.
		base: base02,
	}

	table := []struct {
		name string
		x, y uint
		ops  []Expr
		want string
	}{
		{name: "and", x: 0b01101001, y: 0b01010101, ops: []Expr{And}, want: "0b01000001"},
		{name: "or", x: 0b01101001, y: 0b01010101, ops: []Expr{Or}, want: "0b01111101"},
		{name: "xor", x: 0b01101001, y: 0b01010101, ops: []Expr{Xor}, want: "0b00111100"},

		{name: "shl", x: 3, y: 0b01101001, ops: []Expr{LeftShift}, want: "0b1101001000"},
		{name: "shr", x: 3, y: 0b01101001, ops: []Expr{RightShift}, want: "0b00001101"},
		{name: "asr", x: 3, y: 0xf000000001101001, ops: []Expr{ArithShift}, want: "0b1111111000000000000000000000000000000000001000100000001000000000"},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			y := Value{T: integer, V: tt.y, m: m}
			x := Value{T: integer, V: tt.x, m: m}

			m.Push(y)
			m.Push(x)

			r, err := m.Eval(0, tt.ops)

			if err != nil {
				t.Errorf("%s: %v", tt.name, err)
			}

			if got, ok := r.(string); ok {
				if got != tt.want {
					t.Errorf("%s: wanted %v, got %v", tt.name, tt.want, got)
				}
			} else {
				t.Errorf("%s: invalid result %v %[1]T", tt.name, r)
			}
		})
	}
}
