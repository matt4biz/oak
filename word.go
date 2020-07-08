package oak

import (
	"fmt"
	"strings"

	"oak/token"
)

// Word represents a stack-based function (macro).
//
// The word contains its tokens and runs a separate
// parser to compile its logic to a list of expressions.
// The token list includes the colon, name and semicolon.
//
// the first will become the word's name, and all the
// other tokens up to ";" will make the expressions in
// the word. These expressions will evaluate in turn
// when Eval() is called.
//
// the compile function needs the machine to find symbols
// (and identify result vars, which are not allowed).
type Word struct {
	N string        `json:"name"`
	T []token.Token `json:"tokens"`
	E []Expr        `json:"-"`
}

// Eval runs all the expressions in the word's definition
// against the machine, but preserves the old top-of-stack
// so that its X input is captured as the "last X" value
// (and not something used internally in the word).
func (w *Word) Eval(m *Machine) error {
	if len(w.E) == 0 {
		return fmt.Errorf("%s: not compiled", w.N)
	}

	// we need to save the top of stack and
	// then reset last x to that value below
	// so that the word appears to be one op

	x := m.Top()

	for _, e := range w.E {
		if e == nil {
			return fmt.Errorf("found nil expression")
		}

		if err := e.Eval(m); err != nil {
			return err
		}
	}

	m.x = x

	return nil
}

// Compile in its initial version will not support logic
// or iteration, just calling a sequence of existing words.
func (w *Word) Compile(m *Machine) error {
	var err error

	l := len(w.T)

	if l < 4 ||
		w.T[0].Type != token.Colon ||
		w.T[1].Type != token.Identifier ||
		w.T[l-1].Type != token.Semicolon {
		return fmt.Errorf("invalid definition")
	}

	p := WordParser(m, w.T[2:l-1])

	w.N = w.T[1].Text
	w.E, err = p.Compile()

	if err != nil {
		return fmt.Errorf("%s", err)
	}

	return nil
}

// Definition returns a string representation of the
// words definition (tokens not including : name ;)
func (w *Word) Definition() string {
	var s []string
	for _, t := range w.T[2 : len(w.T)-1] {
		s = append(s, t.Text)
	}
	return strings.Join(s, " ")
}

// Install installs a word into the machine
// (assuming it compiled successfully).
func (m *Machine) Install(w *Word) {
	m.words[w.N] = w
}

// Word returns the word for the given name if
// known, and an error otherwise.
func (m *Machine) Word(s string) Expr {
	w, ok := m.words[s]

	if !ok {
		return nil
	}

	return w
}
