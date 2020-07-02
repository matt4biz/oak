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

	m1.Push(m1.makeFloatVal(3))
	m1.Push(m1.makeFloatVal(2))
	m1.Push(m1.makeFloatVal(1))

	_, err = m1.Eval(1, []Expr{Add})

	if err != nil {
		t.Fatalf("add: %s", err)
	}

	// now we'll save the machine, and read it
	// into a new, clean machine, and see if we
	// still have the values and variables

	m1.Push(m1.makeStringVal(file.Name()))

	if _, err = m1.Eval(1, []Expr{Save}); err != nil {
		t.Fatalf("save: %s", err)
	}

	m2 := New()

	m2.Push(m2.makeStringVal(file.Name()))

	if _, err = m2.Eval(1, []Expr{Load}); err != nil {
		t.Fatalf("load: %s", err)
	}

	// and see how we're doing

	e := GetSymbol("$1")
	v, err := m2.Eval(2, []Expr{e, Add})

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
