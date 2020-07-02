package stack

import "testing"

func TestResultVar(t *testing.T) {
	m := &Machine{vars: make(map[string]*Symbol)}

	m.Push(m.makeFloatVal(3))
	m.Push(m.makeFloatVal(2))

	_, err := m.Eval(1, []Expr{Add})

	if err != nil {
		t.Fatalf("add: %s", err)
	}

	// we should now have a symbol for the result

	e := GetSymbol("$1")
	v, err := m.Eval(2, []Expr{e})

	if err != nil {
		t.Fatalf("get: %s", err)
	}

	// eval results are the strings we print

	r, ok := v.(string)

	if !ok {
		t.Errorf("recall: %#v %[1]T", v)
	}

	if r != "5" {
		t.Errorf("invalid result: %#v", r)
	}
}

func TestUserVar(t *testing.T) {
	m := &Machine{vars: make(map[string]*Symbol)}

	m.Push(m.makeFloatVal(3))
	m.Push(m.makeSymbol("$a"))

	_, err := m.Eval(1, []Expr{Store})

	if err != nil {
		t.Fatalf("store: %s", err)
	}

	// we should now have a symbol for the result

	m.Push(m.makeFloatVal(1))

	e := GetUserVar("$a")
	v, err := m.Eval(2, []Expr{e, Recall, Add})

	if err != nil {
		t.Fatalf("get-add: %s", err)
	}

	// eval results are the strings we print

	r, ok := v.(string)

	if !ok {
		t.Errorf("recall: %#v %[1]T", v)
	}

	if r != "4" {
		t.Errorf("invalid result: %#v", r)
	}
}
