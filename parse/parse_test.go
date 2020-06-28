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
	p := New(m, s, 0, false, true)

	// we can't actually compare the parser's output directly
	// because it's a list of closures, which can't be tested
	// for equality (nor is there any way to peek inside with
	// any safety); so instead we actually run the machine and
	// look at the output line-by-line

	for i, got := 0, p.Line(); len(got) > 0; i, got = i+1, p.Line() {
		top, err := m.Eval(0, got)

		if err != nil {
			t.Errorf("couldn't eval %v: %s", got, err)
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
		input: "2 1 +",
		want:  []string{"3"},
	},
	{
		name:  "simple-add-comma",
		input: "2 1 +, 3+",
		want:  []string{"3", "6"},
	},
	{
		name:  "simple-add-fixed",
		input: "2 fix 2 1 +",
		want:  []string{"3.00"},
	},
	{
		name:  "simple-add-2-comma",
		input: "2 1 +,,3+",
		want:  []string{"3", "6"},
	},
}

func TestParser(t *testing.T) {
	for _, st := range subTests {
		t.Run(st.name, st.run)
	}
}
