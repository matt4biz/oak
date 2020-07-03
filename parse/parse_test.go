package parse

import (
	"bytes"
	"oak/scan"
	"oak/stack"
	"os"
	"testing"
)

type subTest struct {
	name  string
	input string
	want  []string // one for each line
}

func (st subTest) run(t *testing.T) {
	b := bytes.NewBufferString(st.input)
	c := scan.Config{}
	s := scan.New(c, st.name, b)

	m := stack.New()
	p := New(m, s, os.Stderr, 1, true)

	// we can't actually compare the parser's output directly
	// because it's a list of closures, which can't be tested
	// for equality (nor is there any way to peek inside with
	// any safety); so instead we actually run the machine and
	// look at the output line-by-line

	var i int

	for got, _ := p.Line(); len(got) > 0 && got[0] != nil; got, _ = p.Line() {
		top, err := m.Eval(i+1, got)

		if err != nil {
			t.Errorf("couldn't eval %v: %s", got, err)
		} else {
			t.Logf("eval[%d] = %v", i, top)
		}

		if s, ok := top.(string); ok {
			if s != st.want[i] {
				t.Errorf("line %q, wanted %v, got %v", st.input, st.want[i], top)
			}
		} else {
			t.Errorf("line %q, invalid result %#v", st.input, top)
		}

		i++
	}
}

var subTests = []subTest{
	{
		name:  "simple-add",
		input: "2 1 + `comment hex",
		want:  []string{"3"},
	},
	{
		name:  "simple-add-fixed",
		input: "2 fix 2 1 +",
		want:  []string{"3.00"},
	},
	{
		name:  "simple-add-comma",
		input: "2 1 +, 3+",
		want:  []string{"3", "6"},
	},
	{
		name:  "simple-add-2-comma",
		input: "2 1 +,,3+",
		want:  []string{"3", "3", "6"},
	},
	{
		name:  "simple-add-oct",
		input: "127 oct 234+ 007+ dec",
		want:  []string{"368"},
	},
	{
		name:  "simple-add-oct-chg",
		input: "127 oct 234+ 017+ dec 017+",
		want:  []string{"393"},
	},
	{
		name:  "simple-add-hex",
		input: "127 hex 234+ 0x07+ dec",
		want:  []string{"368"},
	},
	{
		name:  "simple-add-bin",
		input: "127 bin 234+ 0b0111 +dec",
		want:  []string{"368"},
	},
	{
		name:  "simple-mode-chg",
		input: `3 fix "rad" mode 0.5236 sin`,
		want:  []string{"0.500"},
	},
	{
		name:  "modulo",
		input: `3 4 *, 2/, 5%`,
		want:  []string{"12", "6", "1"},
	},
	{
		name:  "pow-log",
		input: `3 pow, log`,
		want:  []string{"1000", "3"},
	},
	{
		name:  "power-op-log",
		input: `10 3 **, log`,
		want:  []string{"1000", "3"},
	},
	{
		name:  "last-x",
		input: `2 1 +, 3 $0 +`,
		want:  []string{"3", "4"},
	},
	{
		name:  "result-var",
		input: `2 1 +, 3 $1 +`,
		want:  []string{"3", "6"},
	},
	{
		name:  "user-var",
		input: `2 1 +, 3 $name !, $name@+`,
		want:  []string{"3", "3", "6"},
	},
	{
		name:  "bitwise-xor",
		input: `1 bin 3 ^`,
		want:  []string{"0b00000010"},
	},
	{
		name:  "bitwise-and",
		input: `1 bin 3 &`,
		want:  []string{"0b00000001"},
	},
	{
		name:  "bitwise-or-not",
		input: `1 hex 3|,~`,
		want:  []string{"0x0003", "0xfffffffffffffffc"},
	},
	{
		name:  "shift-l",
		input: `hex 0x1 3 <<`,
		want:  []string{"0x0008"},
	},
	{
		name:  "shift-r",
		input: `hex 0x0100 3 >>`,
		want:  []string{"0x0020"},
	},
	{
		name:  "arith-shift-r",
		input: `hex 0xf000000000000100 3 >>>`,
		want:  []string{"0xfe00000000000020"},
	},
	{
		name:  "bad-parse",
		input: "x",
		want:  []string{},
	},
}

func TestParser(t *testing.T) {
	for _, st := range subTests {
		t.Run(st.name, st.run)
	}
}
