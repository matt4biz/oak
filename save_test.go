package oak

import (
	"io/ioutil"
	"os"
	"testing"

	"oak/token"
)

func TestSaveLoad(t *testing.T) {
	file, err := ioutil.TempFile(".", "*.img")

	if err != nil {
		t.Fatalf("tmp file: %s", err)
	}

	defer os.Remove(file.Name())

	m1 := New(os.Stdout)
	w := Word{
		N: "dB",
		T: []token.Token{
			{Text: ":", Type: token.Colon},
			{Text: "dB", Type: token.Identifier},
			{Text: "log", Type: token.Identifier},
			{Text: "10", Type: token.Number},
			{Text: "*", Type: token.Operator},
			{Text: ";", Type: token.Semicolon},
		},
	}

	err = w.Compile(m1)

	if err != nil {
		t.Fatalf("compile: %s", err)
	} else if l := len(w.E); l != 3 {
		t.Fatalf("dB: invalid length %d", l)
	}

	m1.Install(&w)
	m1.SetFixed(3)

	_, err = m1.Eval(1, []Expr{Number(3), Number(2), Number(1), Add})

	if err != nil {
		t.Fatalf("add: %s", err)
	}

	_, err = m1.Eval(1, []Expr{Number(4), m1.Word("dB"), Swap})

	if err != nil {
		t.Fatalf("dB: %s", err)
	}

	// now we'll save the machine, and read it into a new, clean
	// machine, and see if we still have the values and variables

	if _, err = m1.Eval(1, []Expr{String(file.Name()), Save}); err != nil {
		t.Fatalf("save: %s", err)
	}

	m2 := New(os.Stdout)

	if _, err = m2.Eval(1, []Expr{String(file.Name()), Load}); err != nil {
		t.Fatalf("load: %s", err)
	}

	// and see how we're doing

	v, err := m2.Eval(2, []Expr{Number(2), GetSymbol("$1"), Add, Show})

	if err != nil {
		t.Fatalf("get: %s", err)
	}

	// eval results are the strings we print

	r, ok := v.(string)

	if !ok {
		t.Errorf("recall: %#v %[1]T", v)
	}

	if r != "5.000" {
		t.Errorf("invalid result: %#v", r)
	}

	v, err = m2.Eval(2, []Expr{Number(4), m2.Word("dB"), Show})

	if err != nil {
		t.Fatalf("dB: %s", err)
	}

	r, ok = v.(string)

	if !ok {
		t.Errorf("recall: %#v %[1]T", v)
	}

	if r != "6.021" {
		t.Errorf("invalid result: %#v", r)
	}
}
