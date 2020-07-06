package oak

import "testing"

func TestValueFreeFloat(t *testing.T) {
	m := &Machine{}

	table := []struct {
		v float64
		s string
	}{
		{1, "1"},
		{1.1, "1.1"},
	}

	for _, tt := range table {
		v := Value{T: floater, V: tt.v, m: m}
		s := v.String()

		if s != tt.s {
			t.Errorf("%v: wanted %s, got %s", tt.v, tt.s, s)
		}
	}
}

func TestValueFixedFloat(t *testing.T) {
	m := &Machine{disp: fixed}

	table := []struct {
		v float64
		d uint
		s string
	}{
		{0.1, 1, "0.1"},
		{1, 2, "1.00"},
		{1.1, 3, "1.100"},
	}

	for _, tt := range table {
		m.digits = tt.d

		v := Value{T: floater, V: tt.v, m: m}
		s := v.String()

		if s != tt.s {
			t.Errorf("%v: wanted %s, got %s", tt.v, tt.s, s)
		}
	}
}

func TestValueSciFloat(t *testing.T) {
	m := &Machine{disp: scientific}

	table := []struct {
		v float64
		d uint
		s string
	}{
		{0.1, 1, "1.0e-01"},
		{1, 2, "1.00e+00"},
		{1001, 3, "1.001e+03"},
	}

	for _, tt := range table {
		m.digits = tt.d

		v := Value{T: floater, V: tt.v, m: m}
		s := v.String()

		if s != tt.s {
			t.Errorf("%v: wanted %s, got %s", tt.v, tt.s, s)
		}
	}
}

func TestValueEngFloat(t *testing.T) {
	m := &Machine{disp: engineering}

	table := []struct {
		v float64
		d uint
		s string
	}{
		{0.0001, 1, "100e-06"},
		{0.0001, 2, "100e-06"},
		{0.0001, 3, "100.0e-06"},
		{0.001, 1, "1.0e-03"},
		{0.001, 2, "1.00e-03"},
		{0.01, 1, "10e-03"},
		{0.01, 2, "10.0e-03"},
		{0.1, 1, "100e-03"},
		{0.1, 2, "100e-03"},
		{0.1, 3, "100.0e-03"},
		{1, 1, "1.0e+00"},
		{1, 2, "1.00e+00"},
		{10, 1, "10e+00"},
		{10, 2, "10.0e+00"},
		{10, 3, "10.00e+00"},
		{201, 2, "201e+00"},
		{201, 3, "201.0e+00"},
		{1001, 2, "1.00e+03"},
		{1001, 3, "1.001e+03"},
		{10201, 2, "10.2e+03"},
		{10201, 3, "10.20e+03"},
		{10201, 4, "10.201e+03"},
		{3022201, 2, "3.02e+06"},
		{3022201, 3, "3.022e+06"},
		{30000201, 3, "30.00e+06"},

		{-0.01, 2, "-10.0e-03"},
		{-0.1, 2, "-100e-03"},
		{-201, 2, "-201e+00"},
		{-201, 3, "-201.0e+00"},
	}

	for _, tt := range table {
		m.digits = tt.d

		v := Value{T: floater, V: tt.v, m: m}
		s := v.String()

		if s != tt.s {
			t.Errorf("%v: wanted %s, got %s", tt.v, tt.s, s)
		}
	}
}

func TestPlacesBinary(t *testing.T) {
	table := []struct {
		i uint
		r int
	}{
		{7, 8},
		{257, 9},
		{65537, 17},
	}

	for _, tt := range table {
		r := places(tt.i, 1, 8, 64)

		if r != tt.r {
			t.Errorf("%d: wanted %d, got %d", tt.i, tt.r, r)
		}
	}
}

func TestPlacesOctal(t *testing.T) {
	table := []struct {
		i uint
		r int
	}{
		{7, 3},
		{257, 3},
		{65537, 6},
	}

	for _, tt := range table {
		r := places(tt.i, 3, 3, 24)

		if r != tt.r {
			t.Errorf("%d: wanted %d, got %d", tt.i, tt.r, r)
		}
	}
}

func TestPlacesHexadecimal(t *testing.T) {
	table := []struct {
		i uint
		r int
	}{
		{7, 4},
		{257, 4},
		{65537, 5},
	}

	for _, tt := range table {
		r := places(tt.i, 4, 4, 16)

		if r != tt.r {
			t.Errorf("%d: wanted %d, got %d", tt.i, tt.r, r)
		}
	}
}

func TestValueBinaryInt(t *testing.T) {
	m := &Machine{}

	table := []struct {
		v uint
		b radix
		s string
	}{
		{1, base02, "0b00000001"},
		{5, base02, "0b00000101"},
		{255, base02, "0b11111111"},
		{256, base02, "0b100000000"},
		{4096, base02, "0b1000000000000"},
		{4097, base02, "0b1000000000001"},
		{65535, base02, "0b1111111111111111"},
		{65537, base02, "0b10000000000000001"},

		{1, base08, "001"},
		{7, base08, "007"},
		{255, base08, "0377"},
		{256, base08, "0400"},
		{4096, base08, "010000"},
		{4097, base08, "010001"},
		{65535, base08, "0177777"},
		{65537, base08, "0200001"},

		{1, base16, "0x0001"},
		{7, base16, "0x0007"},
		{255, base16, "0x00ff"},
		{256, base16, "0x0100"},
		{65535, base16, "0xffff"},
		{65537, base16, "0x10001"},
		{1193046, base16, "0x123456"},
		{4886718345, base16, "0x123456789"},
	}

	for _, tt := range table {
		m.base = tt.b

		v := Value{T: integer, V: tt.v, m: m}
		s := v.String()

		if s != tt.s {
			t.Errorf("%v: wanted %s, got %s", tt.v, tt.s, s)
		}
	}
}
