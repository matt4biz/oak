package oak

import "testing"

func TestResultVar(t *testing.T) {
	m := &Machine{vars: make(map[string]*Symbol)}

	_, err := m.Eval(1, []Expr{Integer(3), Integer(2), Add})

	if err != nil {
		t.Fatalf("add: %s", err)
	}

	// we should now have a symbol for the result

	v, err := m.Eval(2, []Expr{GetSymbol("$1")})

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

	_, err := m.Eval(1, []Expr{Integer(3), GetUserVar("$a"), Store})

	if err != nil {
		t.Fatalf("store: %s", err)
	}

	// we should now have a symbol for the result

	v, err := m.Eval(2, []Expr{Integer(1), GetUserVar("$a"), Recall, Add})

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
