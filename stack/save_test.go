package stack

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestSaveLoad(t *testing.T) {
	file, err := ioutil.TempFile(".", "*.img")

	if err != nil {
		t.Fatalf("tmp file: %s", err)
	}

	defer os.Remove(file.Name())

	m1 := New()

	_, err = m1.Eval(1, []Expr{Number(3), Number(2), Number(1), Add})

	if err != nil {
		t.Fatalf("add: %s", err)
	}

	// now we'll save the machine, and read it
	// into a new, clean machine, and see if we
	// still have the values and variables

	if _, err = m1.Eval(1, []Expr{String(file.Name()), Save}); err != nil {
		t.Fatalf("save: %s", err)
	}

	m2 := New()

	if _, err = m2.Eval(1, []Expr{String(file.Name()), Load}); err != nil {
		t.Fatalf("load: %s", err)
	}

	// and see how we're doing

	v, err := m2.Eval(2, []Expr{GetSymbol("$1"), Add, Show})

	if err != nil {
		t.Fatalf("get: %s", err)
	}

	// eval results are the strings we print

	r, ok := v.(string)

	if !ok {
		t.Errorf("recall: %#v %[1]T", v)
	}

	if r != "6" {
		t.Errorf("invalid result: %#v", r)
	}
}
