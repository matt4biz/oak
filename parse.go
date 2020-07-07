package oak

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"oak/token"
)

type Parser struct {
	machine *Machine
	scanner *Scanner
	tokens  []token.Token
	buff    [100]token.Token
	w       io.Writer
	word    *Word
	line    int
	base    int
	debug   bool
	compile bool
}

// New returns a new parser using a particular stack machine and scanner.
// In debug mode it will show tokens for each line.
func NewParser(m *Machine, s *Scanner, w io.Writer, line int, debug bool) *Parser {
	p := Parser{
		machine: m,
		scanner: s,
		w:       w,
		line:    line,
		base:    m.Base(),
		debug:   debug,
	}

	return &p
}

// WordParser returns a new parser with tokens from a word definition;
// it cannot scan for more tokens.
func WordParser(m *Machine, t []token.Token) *Parser {
	p := Parser{
		machine: m,
		tokens:  t,
		w:       m.output,
		line:    t[0].Line,
		base:    m.Base(),
		debug:   m.debug,
		compile: true,
	}

	return &p
}

// Line returns a list of expressions for the machine to evaluate; if
// there's a comma or newline, it will emit a NOP as a placeholder, so
// we only expect an empty list when we've run out of input to parse.
// This requires the scanner to return tokens for newline or comma.
func (p *Parser) Line() ([]Expr, string, error) {
	p.base = p.machine.Base()

	s, ok := p.readTokensToNewline()

	if !ok {
		return nil, "", nil
	}

	if len(p.tokens) > 0 {
		if p.debug {
			fmt.Printf("%d: %s\n", p.line, p.tokens)
		}
		p.line++
	}

	e, err := p.evaluate()

	return e, s, err
}

// Compile is only used for a WordParser, to compile
// the word's tokens once they've all been picked up.
func (p *Parser) Compile() ([]Expr, error) {
	if len(p.tokens) > 0 && p.debug {
		fmt.Printf("%d: %s\n", p.line, p.tokens)
	}

	return p.evaluate()
}

// readTokensToNewline clears the token buffer and fills it
// until a newline (comma) is found, or we reach EOF. For
// EOF we get true if we should keep evaluating because there
// are tokens left in the buffer, and false when not. Note
// that newline/comma tokens are explicitly added to the
// token buffer, not absorbed.
func (p *Parser) readTokensToNewline() (string, bool) {
	p.tokens = p.buff[:0]

	for {
		t := p.scanner.Next()

		switch t.Type {
		case token.Error:
			p.errorf("%s", t)

		case token.Newline, token.Comma:
			p.tokens = append(p.tokens, t)
			return p.scanner.Line(), true

		case token.EOF:
			return p.scanner.Line(), len(p.tokens) > 0
		}

		p.tokens = append(p.tokens, t)
	}
}

// evaluate processes actual tokens to create the expression
// list that will be executed. For newline/comma, we return
// a NOP so that the list is not empty until we reach EOF.
func (p *Parser) evaluate() ([]Expr, error) {
	var (
		result []Expr
		e      Expr
		err    error
	)

	for _, t := range p.tokens {
		if p.word != nil {
			p.word.T = append(p.word.T, t)

			if t.Type == token.Semicolon {
				if err := p.word.Compile(p.machine); err != nil {
					p.errorf("invalid word: %s", err)
					return nil, err
				}

				p.machine.Install(p.word)
				p.word = nil
			}

			continue
		}

		switch t.Type {
		case token.Number:
			if e, err = p.number(t.Text); err != nil {
				p.errorf("%s: %s", err, t.Text)
				return nil, err
			}

		case token.Operator:
			if e, err = p.operator(t.Text); err != nil {
				p.errorf("bad operator: %s", t.Text)
				return nil, err
			}

		case token.Identifier:
			if strings.HasPrefix(t.Text, "$") {
				if e, err = p.symbol(t.Text); err != nil {
					p.errorf("bad symbol: %s", t.Text)
					return nil, err
				}
			} else {
				if e, err = p.identifier(t.Text); err != nil {
					p.errorf("unknown name: %s", t.Text)
					return nil, err
				}

				// we need to allow binary input in the middle
				// of an input line when there's a mode change
				p.checkForBaseChange(t.Text)
			}

		case token.String:
			if e, err = p.str(t.Text); err != nil {
				p.errorf("invalid string: %q", t.Text)
				return nil, err
			}

		case token.Colon:
			p.word = new(Word)
			p.word.T = append(p.word.T, t)
			e = nil

		case token.Comma, token.Newline, token.Comment:
			e = Nop
		}

		if e != nil {
			result = append(result, e)
		}
	}

	return result, nil
}

func (p *Parser) errorf(format string, args ...interface{}) {
	p.tokens = p.buff[:0]

	fmt.Fprintf(p.w, format+"\n", args...)
}

func (p *Parser) number(s string) (result Expr, err error) {
	if p.base != 10 {
		// if we're in integer mode, we want to parse integers, possibly
		// with a leading 0 or 0x/0b prefix; floats will not work here

		var n uint64

		if len(s) > 0 && s[0] == '0' {
			if len(s) > 2 {
				// if we don't have a prefix, we have a leading 0 for octal;
				// however, we aren't going to accept (e.g.) "0x" by itself

				if s[1] == 'x' || s[1] == 'X' {
					n, err = strconv.ParseUint(s[2:], 16, 64)
				} else if s[1] == 'b' || s[1] == 'B' {
					n, err = strconv.ParseUint(s[2:], 2, 64)
				} else {
					n, err = strconv.ParseUint(s, 8, 64)
				}
			} else {
				n, err = strconv.ParseUint(s, 10, 64)
			}
		} else {
			n, err = strconv.ParseUint(s, 10, 64)
		}

		if err != nil {
			return nil, err
		}

		return Integer(uint(n)), nil
	}

	if strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B") ||
		strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return nil, fmt.Errorf("binary format invalid")
	}

	var f float64

	if f, err = strconv.ParseFloat(s, 64); err != nil {
		return nil, err
	}

	return Number(f), nil
}

func (p *Parser) checkForBaseChange(id string) {
	switch id {
	case "bin":
		p.base = 2
	case "dec":
		p.base = 10
	case "oct":
		p.base = 8
	case "hex":
		p.base = 16
	}
}

func (p *Parser) str(s string) (Expr, error) {
	if len(s) == 0 {
		return nil, fmt.Errorf("empty string")
	}

	return String(s), nil
}

func (p *Parser) operator(s string) (Expr, error) {
	switch s {
	case "+":
		return Add, nil
	case "*":
		return Multiply, nil
	case "-":
		return Subtract, nil
	case "/":
		return Divide, nil
	case "%":
		return Modulo, nil
	case "**":
		return Power, nil
	case "!":
		return Store, nil
	case "@":
		return Recall, nil
	case "~":
		return Not, nil
	case "&":
		return And, nil
	case "|":
		return Or, nil
	case "^":
		return Xor, nil
	case "<<":
		return LeftShift, nil
	case ">>":
		return RightShift, nil
	case ">>>":
		return ArithShift, nil
	case "∑+":
		return StatsOpAdd, nil
	case "∑-":
		return StatsOpRm, nil
	}

	return nil, errUnknown
}

var (
	resultVar = regexp.MustCompile(`\$[0-9]+`)
	userVar   = regexp.MustCompile(`\$[a-zA-Z][a-zA-Z_0-9]*`)
)

func (p *Parser) symbol(s string) (Expr, error) {
	if resultVar.MatchString(s) {
		if p.compile {
			return nil, fmt.Errorf("invalid result var %s", s)
		}

		return GetSymbol(s), nil
	}

	if userVar.MatchString(s) {
		return GetUserVar(s), nil
	}

	return nil, errUnknown
}

func (p *Parser) identifier(s string) (Expr, error) {
	if e := Predefined(s); e != nil {
		return e, nil
	}

	if w := p.machine.Word(s); w != nil {
		return w, nil
	}

	return p.machine.Builtin(s)
}

var errUnknown = errors.New("unknown")
