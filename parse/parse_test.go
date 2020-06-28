package parse

import (
	"bytes"
	"oak/scan"
	"oak/stack"
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
	p := New(m, s, 0, true)

	// we can't actually compare the parser's output directly
	// because it's a list of closures, which can't be tested
	// for equality (nor is there any way to peek inside with
	// any safety); so instead we actually run the machine and
	// look at the output line-by-line

	for i, got := 0, p.Line(); len(got) > 0 && got[0] != nil; i, got = i+1, p.Line() {
		top, err := m.Eval(0, got)

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
