package stack

import (
	"fmt"
	"math"
	"math/bits"
	"strconv"
)

type Value struct {
	T tag
	M mode
	V interface{}
	m *Machine
}

func (v Value) String() string {
	switch v.T {
	case floater:
		// floats will always print as floats, not binary
		switch v.m.disp {
		case free:
			// we don't need any special formatting
			return fmt.Sprint(v.V.(float64))

		case fixed:
			return fmt.Sprintf("%.*f", v.m.digits, v.V.(float64))

		case scientific:
			return fmt.Sprintf("%.*e", v.m.digits, v.V.(float64))

		case engineering:
			// we have to calculate an exponent that's a multiple
			// of three, and then scale the number to fit, and
			// then make our own
			f := v.V.(float64)
			n := f < 0
			d := v.m.digits
			s := '+'

			// we only use the log of a positive number
			// so if the original value is negative,
			// we'll change it here, and change back later

			if n {
				f = -f
			}

			e := int(math.Round(math.Log10(f)))

			//fmt.Printf("before: f=%v, e=%v, n=%v, d=%v, s=%q\n", f, e, n, d, s)

			// we need to find the correct multiple of 3
			// which is weird when it's a fractional number

			if e >= 0 {
				e = (e / 3) * 3
			} else {
				e = (-e + 3) / 3 * (-3)
			}

			// scale the number by the new exponent

			f *= math.Pow10(-e)

			// and now, fix the digits as needed because fix=2
			// (0.00) with 10 becomes 10.0 with two significant
			// digits after the mantissa

			if f >= 1000 {
				f /= 1000
				e += 3
			} else if f >= 100 && d > 1 {
				d -= 2
			} else if f >= 10 && d > 0 {
				d -= 1
			}

			// fix the sign of the exponent, since we're
			// making it here, not using %e, etc.

			if e < 0 {
				s = '-'
				e = -e
			}

			//fmt.Printf(" after: f=%v, e=%v, n=%v, d=%v, s=%q\n", f, e, n, d, s)

			// fiddle the negative number back now

			if n {
				f = -f
			}

			// we use .*f so we can tell the format how many
			// digits to use as the variable d

			return fmt.Sprintf("%.*fe%c%02d", d, f, s, e)
		}

	case integer:
		// we only have integers when a binary base is set (2, 8, 16)
		// and so we format the value with a prefix 0b, 0O, 0x

		// we need to find out how many bits; we will then round
		// that value based on the radix (2:8, 8:3, 16:2)
		i := v.V.(int)

		switch v.m.base {
		case base02:
			return fmt.Sprintf("%#0*b", places(i, 8, 8), i)
		case base08:
			return fmt.Sprintf("%#0*o", places(i, 3, 9), i)
		case base16:
			return fmt.Sprintf("%#0*x", places(i, 4, 16), i)
		default:
			return strconv.Itoa(v.V.(int))
		}

	case stringer:
		return v.V.(string)

	case symbol:
		return v.V.(*Symbol).V.String()

	case word:
		return fmt.Sprintf("<%s>", v.V.(*Word).N)
	}

	return "<nil>"
}

func places(i int, b int, m int) int {
	s := 0 // space for sign

	// we only calculate based on the bits
	// required for the absolute value

	if i < 0 {
		i = -i
		s = 1
	}

	n := bits.Len(uint(i))
	r := n / m

	if n%m != 0 {
		r += 1
	}

	return s + (r * b)
}
