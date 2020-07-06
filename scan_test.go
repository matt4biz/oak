package oak

import (
	"bytes"
	"reflect"
	"testing"

	"oak/token"
)

type scanTest struct {
	name  string
	input string
	want  []token.Token
}

func (st scanTest) run(t *testing.T) {
	b := bytes.NewBufferString(st.input)
	c := ScanConfig{}
	s := NewScanner(c, st.name, b)

	var got []token.Token

	for tok := s.Next(); tok.Type != token.EOF; tok = s.Next() {
		got = append(got, tok)
	}

	if !reflect.DeepEqual(st.want, got) {
		t.Errorf("line %q, wanted %v, got %v", st.input, st.want, got)
	}
}

var scanTests = []scanTest{
	{
		name:  "simple-add",
		input: "2 -1 + # comment",
		want: []token.Token{
			{Type: token.Number, Line: 1, Text: "2"},
			{Type: token.Number, Line: 1, Text: "-1"},
			{Type: token.Operator, Line: 1, Text: "+"},
			{Type: token.Comment, Line: 1},
		},
	},
	{
		name:  "simple-add-comma",
		input: "2 1 +, 3+",
		want: []token.Token{
			{Type: token.Number, Line: 1, Text: "2"},
			{Type: token.Number, Line: 1, Text: "1"},
			{Type: token.Operator, Line: 1, Text: "+"},
			{Type: token.Comma, Line: 1},
			{Type: token.Number, Line: 2, Text: "3"},
			{Type: token.Operator, Line: 2, Text: "+"},
		},
	},
	{
		name: "simple-add-2-lines",
		input: `2 1e2 +
                3+`,
		want: []token.Token{
			{Type: token.Number, Line: 1, Text: "2"},
			{Type: token.Number, Line: 1, Text: "1e2"},
			{Type: token.Operator, Line: 1, Text: "+"},
			{Type: token.Newline, Line: 1},
			{Type: token.Number, Line: 2, Text: "3"},
			{Type: token.Operator, Line: 2, Text: "+"},
		},
	},
	{
		name:  "simple-add-2-comma",
		input: "2 1 + ,, 3+",
		want: []token.Token{
			{Type: token.Number, Line: 1, Text: "2"},
			{Type: token.Number, Line: 1, Text: "1"},
			{Type: token.Operator, Line: 1, Text: "+"},
			{Type: token.Comma, Line: 1},
			{Type: token.Comma, Line: 2},
			{Type: token.Number, Line: 3, Text: "3"},
			{Type: token.Operator, Line: 3, Text: "+"},
		},
	},
	{
		name:  "simple-add-2-comma-split",
		input: "2 1 +, ,3+",
		want: []token.Token{
			{Type: token.Number, Line: 1, Text: "2"},
			{Type: token.Number, Line: 1, Text: "1"},
			{Type: token.Operator, Line: 1, Text: "+"},
			{Type: token.Comma, Line: 1},
			{Type: token.Comma, Line: 2},
			{Type: token.Number, Line: 3, Text: "3"},
			{Type: token.Operator, Line: 3, Text: "+"},
		},
	},
	{
		name:  "simple-binary",
		input: `127 oct 0x234+ 017+ dec 0b011+`,
		want: []token.Token{
			{Type: token.Number, Line: 1, Text: "127"},
			{Type: token.Identifier, Line: 1, Text: "oct"},
			{Type: token.Number, Line: 1, Text: "0x234"},
			{Type: token.Operator, Line: 1, Text: "+"},
			{Type: token.Number, Line: 1, Text: "017"},
			{Type: token.Operator, Line: 1, Text: "+"},
			{Type: token.Identifier, Line: 1, Text: "dec"},
			{Type: token.Number, Line: 1, Text: "0b011"},
			{Type: token.Operator, Line: 1, Text: "+"},
		},
	},
	{
		name:  "simple-mode-chg",
		input: `3 fix "rad" mode 0.5236 sin`,
		want: []token.Token{
			{Type: token.Number, Line: 1, Text: "3"},
			{Type: token.Identifier, Line: 1, Text: "fix"},
			{Type: token.String, Line: 1, Text: "\"rad\""},
			{Type: token.Identifier, Line: 1, Text: "mode"},
			{Type: token.Number, Line: 1, Text: "0.5236"},
			{Type: token.Identifier, Line: 1, Text: "sin"},
		},
	},
}

func TestScanner(t *testing.T) {
	for _, st := range scanTests {
		t.Run(st.name, st.run)
	}
}
