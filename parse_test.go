package oak

import (
	"bytes"
	"os"
	"testing"
)

type parseTest struct {
	name  string
	input string
	want  []string // one for each line
	err   string
}

func (st parseTest) run(t *testing.T) {
	b := bytes.NewBufferString(st.input)
	c := ScanConfig{}
	s := NewScanner(c, st.name, b)

	m := New(os.Stdout)
	p := NewParser(m, s, os.Stderr, 1, true)

	// we can't actually compare the parser's output directly
	// because it's a list of closures, which can't be tested
	// for equality (nor is there any way to peek inside with
	// any safety); so instead we actually run the machine and
	// look at the output line-by-line

	var i int

	for {
		got, _, err := p.Line()

		if err != nil {
			if st.err != err.Error() {
				t.Fatalf("couldn't parse: %s", err)
			}

			// we got our error, so we're done
			return
		} else if st.err != "" {
			t.Fatalf("expected err=%s, didn't get it", st.err)
		}

		if len(got) == 0 || got[0] == nil {
			break
		}

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

var parseTests = []parseTest{
	{
		name:  "simple-add",
		input: "2 1 + # comment hex",
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
		name:  "simple-macro",
		input: `3 fix :dB log 10*; 4 dB`,
		want:  []string{"6.021"},
	},
	{
		name:  "failed-macro",
		input: `:1;`,
		err:   "invalid definition",
	},
	{
		name:  "failed-macro-var",
		input: `:db $1 10*;`,
		err:   "invalid result var $1",
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
		name:  "trig",
		input: `3 fix 30 sin, 2 * atan, 15+ cos`,
		want:  []string{"0.500", "45.000", "0.500"},
	},
	{
		name:  "log",
		input: `3 fix 100 log, exp`,
		want:  []string{"2.000", "7.389"},
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
		name:  "percent",
		input: `40 5 perc, +`,
		want:  []string{"2", "42"},
	},
	{
		name:  "stats-mean",
		input: `2 fix 4.63 0 sum, 5.78 20 sum, 6.61 40 sum, 7.21 60 sum, 7.78 80 sum, mean, swap`,
		want:  []string{"1.00", "2.00", "3.00", "4.00", "5.00", "40.00", "6.40"},
	},
	{
		name:  "stats-one-var",
		input: `2 fix 4.63 sum, 5.78 sum, 6.61 sum, 7.21 sum, 7.78 sum, mean`,
		want:  []string{"1.00", "2.00", "3.00", "4.00", "5.00", "6.40"},
	},
	{
		name:  "stats-sdev",
		input: `2 fix 4.63 0 ∑+, 5.78 20 ∑+, 6.61 40 ∑+, 7.21 60 ∑+, 7.78 80 ∑+, sdev, swap`,
		want:  []string{"1.00", "2.00", "3.00", "4.00", "5.00", "31.62", "1.24"},
	},
	{
		name:  "stats-line",
		input: `2 fix 4.63 0 sum, 5.78 20 sum, 6.61 40 sum, 7.21 60 sum, 7.78 80 sum, line, swap`,
		want:  []string{"1.00", "2.00", "3.00", "4.00", "5.00", "4.86", "0.04"},
	},
	{
		name:  "stats-estm",
		input: `2 fix 4.63 0 sum, 5.78 20 sum, 6.61 40 sum, 7.21 60 sum, 7.78 80 sum, 70 estm, swap`,
		want:  []string{"1.00", "2.00", "3.00", "4.00", "5.00", "7.56", "0.99"},
	},
	{
		name:  "stats-correction",
		input: `2 fix 4.63 0 ∑+, 4.78 20 ∑+, 6.61 40 ∑+, 7.21 60 ∑+, 7.78 80 ∑+, 4.78 20 ∑-, 5.78 20 ∑+, mean, swap`,
		want:  []string{"1.00", "2.00", "3.00", "4.00", "5.00", "4.00", "5.00", "40.00", "6.40"},
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
		err:   "x unknown",
	},
}

func TestParser(t *testing.T) {
	for _, st := range parseTests {
		t.Run(st.name, st.run)
	}
}
