package scan

import (
	"bytes"
	"reflect"
	"testing"
)

type subTest struct {
	name  string
	input string
	want  []Token
}

func (st subTest) run(t *testing.T) {
	b := bytes.NewBufferString(st.input)
	c := Config{}
	s := New(c, st.name, b)

	var got []Token

	for tok := s.Next(); tok.Type != EOF; tok = s.Next() {
		got = append(got, tok)
	}

	if !reflect.DeepEqual(st.want, got) {
		t.Errorf("line %q, wanted %v, got %v", st.input, st.want, got)
	}
}

var subTests = []subTest{
	{
		name:  "simple-add",
		input: "2 -1 + # comment",
		want: []Token{
			{Number, 1, "2"},
			{Number, 1, "-1"},
			{Operator, 1, "+"},
			{Comment, 1, ""},
		},
	},
	{
		name:  "simple-add-comma",
		input: "2 1 +, 3+",
		want: []Token{
			{Number, 1, "2"},
			{Number, 1, "1"},
			{Operator, 1, "+"},
			{Comma, 1, ""},
			{Number, 2, "3"},
			{Operator, 2, "+"},
		},
	},
	{
		name: "simple-add-2-lines",
		input: `2 1e2 +
                3+`,
		want: []Token{
			{Number, 1, "2"},
			{Number, 1, "1e2"},
			{Operator, 1, "+"},
			{Newline, 1, ""},
			{Number, 2, "3"},
			{Operator, 2, "+"},
		},
	},
	{
		name:  "simple-add-2-comma",
		input: "2 1 + ,, 3+",
		want: []Token{
			{Number, 1, "2"},
			{Number, 1, "1"},
			{Operator, 1, "+"},
			{Comma, 1, ""},
			{Comma, 2, ""},
			{Number, 3, "3"},
			{Operator, 3, "+"},
		},
	},
	{
		name:  "simple-add-2-comma-split",
		input: "2 1 +, ,3+",
		want: []Token{
			{Number, 1, "2"},
			{Number, 1, "1"},
			{Operator, 1, "+"},
			{Comma, 1, ""},
			{Comma, 2, ""},
			{Number, 3, "3"},
			{Operator, 3, "+"},
		},
	},
	{
		name:  "simple-binary",
		input: `127 oct 0x234+ 017+ dec 0b011+`,
		want: []Token{
			{Number, 1, "127"},
			{Identifier, 1, "oct"},
			{Number, 1, "0x234"},
			{Operator, 1, "+"},
			{Number, 1, "017"},
			{Operator, 1, "+"},
			{Identifier, 1, "dec"},
			{Number, 1, "0b011"},
			{Operator, 1, "+"},
		},
	},
	{
		name:  "simple-mode-chg",
		input: `3 fix "rad" mode 0.5236 sin`,
		want: []Token{
			{Number, 1, "3"},
			{Identifier, 1, "fix"},
			{String, 1, "\"rad\""},
			{Identifier, 1, "mode"},
			{Number, 1, "0.5236"},
			{Identifier, 1, "sin"},
		},
	},
}

func TestScanner(t *testing.T) {
	for _, st := range subTests {
		t.Run(st.name, st.run)
	}
}
