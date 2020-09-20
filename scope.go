package oak

import (
	"fmt"
)

// Scope holds the local variables of a word; note
// that for now (at least), scopes don't nest
type Scope struct {
	vars map[string]*Symbol
}

// Add puts the local variable name in scope and adds
// to the word's expression list a function to pop a
// stack value into the variable
func (sc *Scope) Add(s string) (ExprFunc, error) {
	if sc.vars == nil {
		sc.vars = make(map[string]*Symbol)
	} else if _, ok := sc.vars[s]; ok {
		return nil, fmt.Errorf("duplicate local %s", s)
	}

	sc.vars[s] = &Symbol{S: s, result: true}

	// this is a expression to pop a value from
	// the stack in order to set up a local var

	f := func(m *Machine) error {
		// we don't need to check, we know it's there
		// and we may be resetting its value in a loop

		v := sc.vars[s]

		v.V = m.PopX()
		return nil
	}

	return f, nil
}

// Has returns true if there's a valid scope with
// this variable in it (not including leading '$')
func (sc *Scope) Has(s string) bool {
	if sc == nil {
		return false
	}

	_, ok := sc.vars[s[1:]]

	return ok
}

// GetSymbol returns an expression pushing the
// value of the symbol onto the stack immediately
// from the local scope (ignoring the leading '$')
func (sc *Scope) GetSymbol(s string) ExprFunc {
	s = s[1:] // ignore the leading $

	return func(m *Machine) error {
		v, ok := sc.vars[s]

		if !ok || v.V == nil {
			return fmt.Errorf("%s undefined", s)
		}

		// we push the value which is a number
		// whose value will be used as an operand

		m.Push(*v.V)
		return nil
	}
}
