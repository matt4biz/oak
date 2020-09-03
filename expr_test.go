package oak

import (
	"math"
	"os"
	"strconv"
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

		{name: "sub-f-f-oct-neg", xt: floater, yt: floater, x: 3.0, y: 2.0, ops: []Expr{Subtract}, base: base08,
			want: "01777777777777777777777"},
		{name: "sub-i-f-oct-neg", xt: integer, yt: floater, x: uint(3), y: 2.0, ops: []Expr{Subtract}, base: base08,
			want: "01777777777777777777777"},
		{name: "sub-f-i-oct-neg", xt: floater, yt: integer, x: 3.0, y: uint(2), ops: []Expr{Subtract}, base: base08,
			want: "01777777777777777777777"},
		{name: "sub-i-i-oct-neg", xt: integer, yt: integer, x: uint(3), y: uint(2), ops: []Expr{Subtract}, base: base08,
			want: "01777777777777777777777"},

		{name: "dist", xt: floater, yt: floater, x: 3.0, y: 4.0, ops: []Expr{Predefined("dist")}, want: "5"},
		{name: "perc", xt: floater, yt: floater, x: 30.0, y: 4.0, ops: []Expr{Predefined("perc")}, want: "1.2"},
		{name: "dperc", xt: floater, yt: floater, x: 5.0, y: 4.0, ops: []Expr{Predefined("dperc")}, want: "25"},

		{name: "perm", xt: floater, yt: floater, x: 3.0, y: 5.0, ops: []Expr{Predefined("perm")}, want: "60"},
		{name: "comb", xt: floater, yt: floater, x: 3.0, y: 5.0, ops: []Expr{Predefined("comb")}, want: "10"},
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
		{name: "abs-f-dec", xt: floater, x: -3.0, ops: []Expr{Predefined("abs")}, want: "3.000"},
		{name: "ceil-f-dec", xt: floater, x: 4.4, ops: []Expr{Predefined("ceil")}, want: "5.000"},
		{name: "frac-f-dec", xt: floater, x: 4.4, ops: []Expr{Predefined("frac")}, want: "0.400"},
		{name: "trunc-f-dec", xt: floater, x: 4.4, ops: []Expr{Predefined("trunc")}, want: "4.000"},
		{name: "recp-f-dec", xt: floater, x: 4.0, ops: []Expr{Predefined("recp")}, want: "0.250"},
		{name: "fact-f-dec", xt: floater, x: 4.0, ops: []Expr{Predefined("fact")}, want: "24.000"},

		{name: "cbrt-f-dec", xt: floater, x: 27.0, ops: []Expr{Predefined("cbrt")}, want: "3.000"},
		{name: "cube-f-dec", xt: floater, x: 2.0, ops: []Expr{Predefined("cube")}, want: "8.000"},
		{name: "sqr-f-dec", xt: floater, x: 5.0, ops: []Expr{Predefined("sqr")}, want: "25.000"},
		{name: "sqrt-f-dec", xt: floater, x: 4.0, ops: []Expr{Predefined("sqrt")}, want: "2.000"},

		{name: "log-f-dec", xt: floater, x: 100.0, ops: []Expr{Predefined("log")}, want: "2.000"},
		{name: "alog-f-dec", xt: floater, x: 3.0, ops: []Expr{Predefined("alog")}, want: "1000.000"},
		{name: "ln-f-dec", xt: floater, x: 100.0, ops: []Expr{Predefined("ln")}, want: "4.605"},
		{name: "ln1-f-dec", xt: floater, x: math.E, ops: []Expr{Predefined("ln")}, want: "1.000"},
		{name: "exp-f-dec", xt: floater, x: 1.0, ops: []Expr{Predefined("exp")}, want: "2.718"},

		{name: "log-i-hex", xt: integer, x: uint(100), ops: []Expr{Predefined("log")}, base: base16, want: "0x0002"},
		{name: "alog-i-hex", xt: integer, x: uint(3), ops: []Expr{Predefined("alog")}, base: base16, want: "0x03e8"},

		{name: "not-i-bin", xt: integer, x: uint(0b01101001), ops: []Expr{Not}, base: base02,
			want: "0b1111111111111111111111111111111111111111111111111111111110010110"},
		{name: "maskl", xt: integer, x: uint(3), ops: []Expr{Predefined("maskl")}, base: base16, want: "0xe000000000000000"},
		{name: "maskr", xt: integer, x: uint(6), ops: []Expr{Predefined("maskr")}, base: base16, want: "0x003f"},
		{name: "mask65", xt: integer, x: uint(65), ops: []Expr{Predefined("maskr")}, base: base16, want: "0xffffffffffffffff"},
		{name: "popcnt", xt: integer, x: uint(65535), ops: []Expr{Predefined("popcnt")}, base: base16, want: "0x0010"},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			x := Value{T: tt.xt, V: tt.x, m: m}

			m.base = tt.base
			m.digits = 3
			m.disp = fixed

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
		{name: "tan-f-deg", xt: floater, x: 45.0, ops: []Expr{Predefined("tan")}, want: "1.000"},

		{name: "asin-f-deg", xt: floater, x: 0.5, ops: []Expr{Predefined("asin")}, want: "30.000"},
		{name: "acos-f-deg", xt: floater, x: 0.5, ops: []Expr{Predefined("acos")}, want: "60.000"},
		{name: "atan-f-deg", xt: floater, x: 1.0, ops: []Expr{Predefined("atan")}, want: "45.000"},

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
		{name: "asr", x: 3, y: 0xf000000001101001, ops: []Expr{ArithShift},
			want: "0b1111111000000000000000000000000000000000001000100000001000000000"},
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

func TestStatsOp(t *testing.T) {
	data := []struct {
		op   Expr
		x, y float64
		n    int
	}{
		{StatsOpAdd, 0, 4.63, 1},
		{StatsOpAdd, 20, 4.78, 2},
		{StatsOpAdd, 40, 6.61, 3},
		{StatsOpAdd, 60, 7.21, 4},
		{StatsOpAdd, 80, 7.78, 5},
		{StatsOpRm, 20, 4.78, 4},
		{StatsOpAdd, 20, 5.78, 5},
	}

	m := New(os.Stdout)

	m.setDisplay("fix")
	m.digits = 2

	for _, d := range data {
		s, err := m.Eval(0, []Expr{Number(d.y), Number(d.x), d.op})

		if err != nil {
			t.Fatalf("entering stats: %s", err)
		}

		if j, _ := strconv.ParseFloat(s.(string), 64); int(j) != d.n {
			t.Fatalf("invalid data point %d, want %d", int(j), d.n)
		}
	}

	s, err := m.Eval(0, []Expr{Average})

	if err != nil {
		t.Fatal(err)
	}

	if ave, ok := s.(string); !ok || ave != "40.00" {
		t.Errorf("incorrect ave: %v", s)
	}

	s, err = m.Eval(0, []Expr{StdDeviation})

	if err != nil {
		t.Fatal(err)
	}

	if sd, ok := s.(string); !ok || sd != "31.62" {
		t.Errorf("incorrect ave: %v", s)
	}

	s, err = m.Eval(0, []Expr{LinRegression})

	if err != nil {
		t.Fatal(err)
	}

	if lr, ok := s.(string); !ok || lr != "4.86" {
		t.Errorf("incorrect ave: %v", s)
	}

	s, err = m.Eval(0, []Expr{Number(70.0), LinEstimate})

	if err != nil {
		t.Fatal(err)
	}

	if ey, ok := s.(string); !ok || ey != "7.56" {
		t.Errorf("incorrect ave: %v", s)
	}

	clr, _ := m.Builtin("clrall")

	if _, err = m.Eval(0, []Expr{clr}); err != nil {
		t.Fatalf("can't clear stats: %s", err)
	}

	if _, err = m.Eval(0, []Expr{Average}); err != errNoStats {
		t.Fatalf("invalid err %v, expected no stats", err)
	}
}
