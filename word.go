package oak

import (
	"fmt"
	"strings"

	"oak/token"
)

// Word represents a stack-based function (macro).
//
// I think we're going to run the parser to collect the
// tokens that make up the word, then ask it to compile
// itself, at which point it becomes an expression that
// can be called.
//
// so ":name op op... ;" turns into a list of tokens
// starting with the colon and ending with the semicolon.
//
// the first will become the word's name, and all the
// other tokens up to ";" will make the expressions in
// the word.
//
// the compile function needs the machine to find symbols
// (and identify result vars, which are not allowed).
type Word struct {
	N string
	E []Expr
	T []token.Token
}

func (w *Word) Eval(m *Machine) error {
	if len(w.E) == 0 {
		return fmt.Errorf("%s: not compiled", w.N)
	}

	for _, e := range w.E {
		if e == nil {
			return fmt.Errorf("found nil expression")
		}

		if err := e.Eval(m); err != nil {
			return err
		}
	}

	return nil
}

// Compile in its initial version will not support logic
// or iteration, just calling a sequence of existing words.
func (w *Word) Compile(m *Machine) error {
	var err error

	l := len(w.T)

	if l < 4 || w.T[0].Type != token.Colon || w.T[1].Type != token.Identifier || w.T[l-1].Type != token.Semicolon {
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

func (w *Word) Definition() string {
	var s []string
	for _, t := range w.T[2 : len(w.T)-1] {
		s = append(s, t.Text)
	}
	return strings.Join(s, " ")
}

func (m *Machine) Install(w *Word) {
	m.words[w.N] = w
}

func (m *Machine) Word(s string) Expr {
	w, ok := m.words[s]

	if !ok {
		return nil
	}

	return w
}
