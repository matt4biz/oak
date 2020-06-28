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
		input: "2 1 +",
		want: []Token{
			{Number, 1, "2"},
			{Number, 1, "1"},
			{Operator, 1, "+"},
		},
	},
	{
		name:  "simple-add-comma",
		input: "2 1 +, 3+",
		want: []Token{
			{Number, 1, "2"},
			{Number, 1, "1"},
			{Operator, 1, "+"},
			{Newline, 2, ","},
			{Number, 2, "3"},
			{Operator, 2, "+"},
		},
	},
	{
		name: "simple-add-2-lines",
		input: `2 1 +
                3+`,
		want: []Token{
			{Number, 1, "2"},
			{Number, 1, "1"},
			{Operator, 1, "+"},
			{Newline, 2, "\n"},
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
			{Newline, 2, ","},
			{Newline, 3, ","},
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
			{Newline, 2, ","},
			{Newline, 3, ","},
			{Number, 3, "3"},
			{Operator, 3, "+"},
		},
	},
}

func TestScanner(t *testing.T) {
	for _, st := range subTests {
		t.Run(st.name, st.run)
	}
}
