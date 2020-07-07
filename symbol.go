package oak

import "fmt"

// Symbol represents a variable
type Symbol struct {
	S        string `json:"symbol"`
	V        *Value `json:"value"`
	result   bool
	readonly bool
}

func (s Symbol) String() string {
	if s.V == nil {
		return fmt.Sprintf("{S:%s, V:<nil> rslt=%t, read=%t}", s.S, s.result, s.readonly)
	}

	return fmt.Sprintf("{S:%s, V:%#v rslt=%t, read=%t}", s.S, *s.V, s.result, s.readonly)
}

// Lookup takes a variable name and returns the symbol
// which includes its value.
func (m *Machine) Lookup(s string) *Symbol {
	if s == "$0" {
		return &Symbol{S: s, V: m.x, result: true}
	}

	return m.vars[s]
}

func (m *Machine) Builtin(s string) (Expr, error) {
	b, ok := m.builtin[s]

	if !ok {
		return nil, fmt.Errorf("%s unknown", s)
	}

	return b, nil
}

// GetSymbol returns an expression pushing the
// value of the symbol onto the stack immediately.
func GetSymbol(s string) ExprFunc {
	return func(m *Machine) error {
		v := m.Lookup(s)

		if v == nil || v.V == nil {
			return fmt.Errorf("%s undefined", s)
		}

		// we push the value which is a number
		// whose value will be used as an operand

		m.Push(*v.V)
		return nil
	}
}

// GetUserVar returns an expression pushing a user-
// defined symbol onto the stack; the symbol pointer
// resides in the value, but it may not exist yet.
func GetUserVar(s string) ExprFunc {
	return func(m *Machine) error {
		// different from above: we don't care if the
		// variable has been defined, we just need to
		// put a symbol on the stack to access it

		// we push the value which contains a symbol
		// pointer for @ or ! to reference

		v := Value{T: symbol, V: &Symbol{S: s}, m: m}

		m.Push(v)
		return nil
	}
}

// StoreVar writes (or overwrites) a given variable name
// with a new value, for use with the ! operator.
func (m *Machine) StoreVar(s *Symbol, v Value) {
	m.vars[s.S] = &Symbol{S: s.S, V: &v}
}

// RecallVar returns the value for a given symbol if it's
// been stored into the machine.
func (m *Machine) RecallVar(s *Symbol) (*Value, error) {
	if v, ok := m.vars[s.S]; ok {
		return v.V, nil
	}

	return nil, fmt.Errorf("%s undefined", s.S)
}

var (
	// Store takes a value {y} and a symbol {x} from
	// the stack to store the value into the machine
	// as a user-defined variable.
	Store ExprFunc = func(m *Machine) error {
		l := len(m.stack)

		if l < 2 {
			return errUnderflow
		}

		s := m.Pop() // we're not going to use PopX
		v := m.Pop()

		switch v.T {
		case integer, floater, stringer:
		default:
			return fmt.Errorf("store: invalid value %#v", v.V)
		}

		if s.T != symbol {
			return fmt.Errorf("store: invalid operand %#v", v.V)
		}

		u, ok := s.V.(*Symbol)

		if !ok {
			return fmt.Errorf("store: invalid symbol %#v", v.V)
		}

		// the symbol we get isn't the original, so look it up;
		// it's only read-only if it already exists and is marked

		if m.vars[u.S] != nil && m.vars[u.S].readonly {
			return fmt.Errorf("store: readonly variable")
		}

		m.StoreVar(u, *v)
		return nil
	}

	// Recall takes a user-define variable symbol from the
	// stack and replaces it with the symbol's value.
	Recall ExprFunc = func(m *Machine) error {
		l := len(m.stack)

		if l < 1 {
			return errUnderflow
		}

		s := m.Pop() // we're not going to use PopX

		if s.T != symbol {
			return fmt.Errorf("recall: invalid operand %#v", s)
		}

		u, ok := s.V.(*Symbol)

		if !ok {
			return fmt.Errorf("recall: invalid symbol %#v", u.V)
		}

		v, err := m.RecallVar(u)

		if err != nil {
			return err
		}

		m.Push(*v)
		return nil
	}
)
