package stack

import (
	"fmt"
	"math"
	"math/bits"
	"strconv"
)

type tag int

const (
	floater tag = iota
	integer
	stringer
	symbol
	word
)

type (
	mode    int
	radix   int
	display int
)

const (
	degrees mode = iota
	radians
)

const (
	base10 radix = iota
	base02
	base08
	base16
)

const (
	free display = iota
	fixed
	scientific
	engineering
)

// Machine represents a stack-based computational engine
// where all operations take items from the stack and/or
// push items onto the stack. It has a "last x" side
// register as well as a map of variables that aren't
// on the stack.
type Machine struct {
	stack  []*Value
	x      *Value
	vars   map[string]*Symbol
	words  map[string]*Word
	built  map[string]Expr
	digits int
	disp   display
	base   radix
	mode   mode
}

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

			e := int(math.Log10(f))

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
			} else if f >= 100 {
				d -= 2
			} else if f >= 10 {
				d -= 1
			}

			// fix the sign of the exponent, since we're
			// making it here, not using %e, etc.

			if e < 0 {
				s = '-'
			}

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

// Symbol represents a variable
type Symbol struct {
	S string
	V *Value
}

// Word represents a stack-based function (macro).
type Word struct {
	N string
	P []Expr
}

// Expr represents an expression (operation) that runs
// against the stack.
type Expr func(m *Machine) error

// New returns a new stack-based machine with an empty
// stack.
func New() *Machine {
	m := Machine{
		vars:  make(map[string]*Symbol, 1024),
		words: make(map[string]*Word),
	}

	m.Load()
	return &m
}

// Eval takes a list of expressions (from the parser)
// and applies them; it stops and discards the list if
// any expression results in a failure. Note that the
// operations the stack exports on itself don't return
// errors, only values (possibly nil, stack unchanged)
func (m *Machine) Eval(line int, exprs []Expr) (interface{}, error) {
	for _, e := range exprs {
		if err := e(m); err != nil {
			return nil, err
		}
	}

	s := fmt.Sprintf("$%d", line)
	t := m.Top()

	if t == nil {
		return nil, nil
	}

	m.vars[s] = &Symbol{s, t}

	return t.String(), nil
}

// Last returns the last top-of-stack value that
// was popped with PopX for use in a calculation.
func (m *Machine) Last() *Value {
	return m.x
}

func (m *Machine) Top() *Value {
	if l := len(m.stack); l > 0 {
		return m.stack[l-1]
	}

	return nil
}

// Pop just removes and returns the top of stack.
func (m *Machine) Pop() *Value {
	switch l := len(m.stack); l {
	case 0:
		return nil
	case 1:
		i := m.stack[0]
		m.stack = nil
		return i
	default:
		i := m.stack[l-1]
		m.stack = m.stack[:l-1]
		return i
	}
}

// PopX is used when we're going to push back some
// result involving the TOS, so Last() can return it.
func (m *Machine) PopX() *Value {
	switch l := len(m.stack); l {
	case 0:
		return nil
	case 1:
		i := m.stack[0]
		m.stack = nil
		m.x = i
		return i
	default:
		i := m.stack[l-1]
		m.stack = m.stack[:l-1]
		m.x = i
		return i
	}
}

func (m *Machine) Push(v Value) {
	m.stack = append(m.stack, &v)
}

func (m *Machine) Swap() {
	if l := len(m.stack); l > 1 {
		m.stack = append(m.stack[0:l-2], m.stack[l-1], m.stack[l-2])
	}
}

// Dup takes the top-of-stack (x) and duplicates
// it on top of the stack.
func (m *Machine) Dup() {
	if l := len(m.stack); l > 0 {
		m.stack = append(m.stack, m.stack[l-1])
	}
}

// Dup2 takes {y,x} and duplicates those two values
// in order on top of the stack.
func (m *Machine) Dup2() {
	if l := len(m.stack); l > 1 {
		m.stack = append(m.stack, m.stack[l-2], m.stack[l-1])
	}
}

// Roll moves the top-of-stack item onto the bottom,
// exposing the next item (y) as the top.
func (m *Machine) Roll() {
	switch l := len(m.stack); l {
	case 0, 1:
		// nothing to do

	case 2:
		m.Swap()

	default:
		tmp := []*Value{m.stack[l-1]}

		m.stack = append(tmp, m.stack[0:l-1]...)
	}
}

func (m *Machine) SetFixed(d int) {
	m.disp = fixed
	m.digits = d
}

func (m *Machine) SetScientific(d int) {
	m.disp = scientific
	m.digits = d
}

func (m *Machine) SetEngineering(d int) {
	m.disp = engineering
	m.digits = d
}

// Lookup takes a variable name and returns the symbol
// which includes its value.
func (m *Machine) Lookup(s string) *Symbol {
	if s == "$0" {
		return &Symbol{s, m.x}
	}

	return m.vars[s]
}

// AddSymbol adds an empty variable to the machine for parsing.
func (m *Machine) AddSymbol(s string) {
	m.vars[s] = &Symbol{s, nil}
}

// Store writes (or overwrites) a given variable name with
// a new value.
func (m *Machine) Store(s *Symbol, v Value) {
	m.vars[s.S] = &Symbol{s.S, &v}
}

func (m *Machine) Known(s string) bool {
	_, ok := m.words[s]
	return ok
}

func (m *Machine) Word(s string) (Expr, error) {
	return nil, nil
}

func (m *Machine) Builtin(s string) (Expr, error) {
	b, ok := m.built[s]

	if !ok {
		return nil, fmt.Errorf("%s unkown", s)
	}

	return b, nil
}

func (m *Machine) Load() {
	m.built = make(map[string]Expr, 1024)

	// CONSTANTS

	m.built["e"] = func(m *Machine) error {
		m.Push(Value{floater, m.mode, math.E, m})
		return nil
	}

	m.built["pi"] = func(m *Machine) error {
		m.Push(Value{floater, m.mode, math.Pi, m})
		return nil
	}

	m.built["phi"] = func(m *Machine) error {
		m.Push(Value{floater, m.mode, math.Phi, m})
		return nil
	}

	// MISCELLANY

	m.built["chs"] = func(m *Machine) error {
		if t := m.Top(); t != nil {
			switch t.T {
			case floater:
				t.V = -t.V.(float64)
				return nil

			case integer:
				t.V = -t.V.(int)
				return nil
			}

			return fmt.Errorf("chs: invalid operand x=%#v", *t)
		}

		return nil
	}

	m.built["clr"] = func(m *Machine) error {
		if t := m.Top(); t != nil {
			switch t.T {
			case floater:
				t.V = 0.0
				return nil

			case integer:
				t.V = 0
				return nil

			case stringer:
				t.V = ""
				return nil
			}

			return fmt.Errorf("chs: invalid operand x=%#v", *t)
		}

		return nil
	}

	// STACK OPERATIONS

	m.built["drop"] = func(m *Machine) error {
		m.Pop()
		return nil
	}

	m.built["dup"] = func(m *Machine) error {
		m.Dup()
		return nil
	}

	m.built["dup2"] = func(m *Machine) error {
		m.Dup2()
		return nil
	}

	m.built["roll"] = func(m *Machine) error {
		m.Roll()
		return nil
	}

	m.built["show"] = func(m *Machine) error {
		m.Dup()
		m.Pop()
		return nil
	}

	m.built["swap"] = func(m *Machine) error {
		m.Swap()
		return nil
	}

	m.built["depth"] = func(m *Machine) error {
		var v interface{}
		var t tag

		s := float64(len(m.stack))

		if m.base == base10 {
			v = s
			t = floater
		} else {
			v = int(s)
			t = integer
		}

		m.Push(Value{t, m.mode, v, m})
		return nil
	}

	// BASE CONVERSION

	m.built["base"] = func(m *Machine) error {
		var b int

		x := m.Pop()

		if x == nil {
			return fmt.Errorf("mode: empty stack")
		}

		switch x.T {
		case floater:
			b = int(x.V.(float64))
		case integer:
			b = x.V.(int)
		case stringer:
			b, _ = strconv.Atoi(x.V.(string))
		}

		switch b {
		case 2:
			m.base = base02
			return nil
		case 8:
			m.base = base08
			return nil
		case 10:
			m.base = base10
			return nil
		case 16:
			m.base = base16
			return nil
		}

		return fmt.Errorf("mode: invalid operand %#v", x.V)
	}

	m.built["hex"] = func(m *Machine) error {
		m.base = base16

		if t := m.Top(); t != nil {
			switch t.T {
			case floater:
				t.V = int(t.V.(float64))
				t.T = integer
				return nil

			case integer:
				return nil
			}

			return fmt.Errorf("hex: invalid operand x=%#v", t.V)
		}

		return nil
	}

	m.built["bin"] = func(m *Machine) error {
		m.base = base02

		if t := m.Top(); t != nil {
			switch t.T {
			case floater:
				t.V = int(t.V.(float64))
				t.T = integer
				return nil

			case integer:
				return nil
			}

			return fmt.Errorf("bin: invalid operand x=%#v", t.V)
		}

		return nil
	}

	m.built["oct"] = func(m *Machine) error {
		m.base = base08

		if t := m.Top(); t != nil {
			switch t.T {
			case floater:
				t.V = int(t.V.(float64))
				t.T = integer
				return nil

			case integer:
				return nil
			}

			return fmt.Errorf("oct: invalid operand x=%#v", t.V)
		}

		return nil
	}

	m.built["dec"] = func(m *Machine) error {
		m.base = base10

		if t := m.Top(); t != nil {
			switch t.T {
			case floater:
				return nil

			case integer:
				t.V = float64(t.V.(int))
				t.T = floater
				return nil
			}

			return fmt.Errorf("dec: invalid operand x=%#v", t.V)
		}

		return nil
	}

	// ANGULAR MODE

	m.built["mode"] = func(m *Machine) error {
		x := m.Pop()

		if x == nil {
			return fmt.Errorf("mode: empty stack")
		}

		if x.T == stringer {
			switch x.V.(string) {
			case "deg":
				m.mode = degrees
				return nil
			case "rad":
				m.mode = radians
				return nil
			}
		}

		return fmt.Errorf("mode: invalid operand %#v", x.V)
	}

	// FORMATTING

	m.built["fix"] = func(m *Machine) error {
		x := m.Pop()

		if x == nil {
			return fmt.Errorf("fix: empty stack")
		}

		switch x.T {
		case floater:
			m.SetFixed(int(x.V.(float64)))
		case integer:
			m.SetFixed(x.V.(int))
		}

		return nil
	}

	m.built["sci"] = func(m *Machine) error {
		x := m.Pop()

		if x == nil {
			return fmt.Errorf("sci: empty stack")
		}

		switch x.T {
		case floater:
			m.SetScientific(int(x.V.(float64)))
		case integer:
			m.SetScientific(x.V.(int))
		}

		return nil
	}

	m.built["eng"] = func(m *Machine) error {
		x := m.Pop()

		if x == nil {
			return fmt.Errorf("eng: empty stack")
		}

		switch x.T {
		case floater:
			m.SetEngineering(int(x.V.(float64)))
		case integer:
			m.SetEngineering(x.V.(int))
		}

		return nil
	}
}

func BinaryOp(op string, f func(float64, float64) float64) Expr {
	return func(m *Machine) error {
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
				m.Push(m.makeVal(s))
				return nil

			case integer:
				s := f(float64(y.V.(int)), x.V.(float64))
				m.Push(m.makeVal(s))
				return nil
			}

		case integer:
			switch y.T {
			case floater:
				s := f(y.V.(float64), float64(x.V.(int)))
				m.Push(m.makeVal(s))
				return nil

			case integer:
				s := f(float64(y.V.(int)), float64(x.V.(int)))
				m.Push(m.makeVal(s))
				return nil
			}
		}

		return fmt.Errorf("%s: mismatched operands y=%#v, x=%#v", op, y.V, x.V)
	}
}

func UnaryOp(op string, f func(float64) float64) Expr {
	return func(m *Machine) error {
		x := m.PopX()

		if x == nil {
			return fmt.Errorf("%s: empty stack", op)
		}

		switch x.T {
		case floater:
			s := f(x.V.(float64))
			m.Push(m.makeVal(s))
			return nil

		case integer:
			s := f(float64(x.V.(int)))
			m.Push(m.makeVal(s))
			return nil
		}

		return fmt.Errorf("%s: invalid operand x=%#v", op, x.V)
	}
}

func TrigonometryOp(op string, f func(float64) float64) Expr {
	return func(m *Machine) error {
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
			m.Push(m.makeVal(s))
			return nil

		case integer:
			var s = float64(x.V.(int))

			if x.M == degrees {
				s *= math.Pi / 180
			}

			s = f(s)
			m.Push(m.makeVal(s))
			return nil
		}

		return fmt.Errorf("%s: invalid operand x=%#v", op, x.V)
	}
}

var Degrees = func(m *Machine) error {
	m.mode = degrees

	if t := m.Top(); t != nil {
		switch t.T {
		case floater:
			t.V = t.V.(float64) * 180 / math.Pi
			t.M = degrees
			return nil

		case integer:
			t.V = int(float64(t.V.(int)) * 180 / math.Pi)
			t.M = degrees
			return nil
		}

		return fmt.Errorf("hex: invalid operand x=%#v", t.V)
	}

	return nil
}

var Radians = func(m *Machine) error {
	m.mode = radians

	if t := m.Top(); t != nil {
		switch t.T {
		case floater:
			t.V = t.V.(float64) * math.Pi / 180
			t.M = radians
			return nil

		case integer:
			t.V = int(float64(t.V.(int)) * math.Pi / 180)
			t.M = radians
			return nil
		}

		return fmt.Errorf("hex: invalid operand x=%#v", t.V)
	}

	return nil
}

func Predefined(s string) Expr {
	switch s {
	case "abs":
		return UnaryOp(s, math.Abs)
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
		return BinaryOp(s, func(y, x float64) float64 { return (x - y) / y * 100 })
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
	case "max":
		return BinaryOp(s, math.Max)
	case "min":
		return BinaryOp(s, math.Min)
	case "perc":
		return BinaryOp(s, func(y, x float64) float64 { return y * x / 100 })
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

// Put a numerical value onto the stack.
func Number(f float64) Expr {
	return func(m *Machine) error {
		m.Push(m.makeVal(f))
		return nil
	}
}

// Put a string value onto the stack.
func String(s string) Expr {
	return func(m *Machine) error {
		m.Push(Value{stringer, m.mode, trimQuotes(s), m})
		return nil
	}
}

// GetSymbol returns the value of the symbol (that may not
// always be what we want, when it's time to load/store)
func GetSymbol(s string) Expr {
	return func(m *Machine) error {
		v := m.Lookup(s)

		if v == nil {
			return fmt.Errorf("can't find %s", s)
		}

		// TODO - if it's a dollarVar push the value, but if
		//   it's any other symbol, push the symbol instead

		m.Push(*v.V)
		return nil
	}
}

func (m *Machine) makeVal(s float64) Value {
	var v interface{}
	var t tag

	if m.base == base10 {
		v = s
		t = floater
	} else {
		v = int(s)
		t = integer
	}

	return Value{t, m.mode, v, m}
}

func places(i int, b int, m int) int {
	n := bits.Len(uint(i))
	r := n / m

	if n%m != 0 {
		r += 1
	}

	return r * b
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}
