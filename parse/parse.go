package parse

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"oak/scan"
	"oak/stack"
)

type Parser struct {
	machine *stack.Machine
	scanner *scan.Scanner
	tokens  []scan.Token
	buff    [100]scan.Token
	w       io.Writer
	line    int
	base    int
	debug   bool
}

// New returns a new parser using a particular stack machine and scanner.
// In debug mode it will show tokens for each line.
func New(m *stack.Machine, s *scan.Scanner, w io.Writer, line int, debug bool) *Parser {
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

// Line returns a list of expressions for the machine to evaluate; if
// there's a comma or newline, it will emit a NOP as a placeholder, so
// we only expect an empty list when we've run out of input to parse.
// This requires the scanner to return tokens for newline or comma.
func (p *Parser) Line() ([]stack.Expr, error) {
	p.base = p.machine.Base()

	if !p.readTokensToNewline() {
		return nil, nil
	}

	if len(p.tokens) > 0 {
		if p.debug {
			fmt.Printf("%d: %s\n", p.line, p.tokens)
		}
		p.line++
	}

	return p.evaluate()
}

// readTokensToNewline clears the token buffer and fills it
// until a newline (comma) is found, or we reach EOF. For
// EOF we get true if we should keep evaluating because there
// are tokens left in the buffer, and false when not. Note
// that newline/comma tokens are explicitly added to the
// token buffer, not absorbed.
func (p *Parser) readTokensToNewline() bool {
	p.tokens = p.buff[:0]

	for {
		tok := p.scanner.Next()

		switch tok.Type {
		case scan.Error:
			p.errorf("%s", tok)

		case scan.Newline, scan.Comma:
			p.tokens = append(p.tokens, tok)
			return true

		case scan.EOF:
			return len(p.tokens) > 0
		}

		p.tokens = append(p.tokens, tok)
	}
}

// evaluate processes actual tokens to create the expression
// list that will be executed. For newline/comma, we return
// a NOP so that the list is not empty until we reach EOF.
func (p *Parser) evaluate() ([]stack.Expr, error) {
	var (
		result []stack.Expr
		e      stack.Expr
		err    error
	)

	for _, t := range p.tokens {
		switch t.Type {
		case scan.Number:
			if e, err = p.number(t.Text); err != nil {
				p.errorf("%s: %s", err, t.Text)
				return nil, err
			}

		case scan.Operator:
			if e, err = p.operator(t.Text); err != nil {
				p.errorf("bad operator: %s", t.Text)
				return nil, err
			}

		case scan.Identifier:
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

		case scan.String:
			if e, err = p.str(t.Text); err != nil {
				p.errorf("invalid string: %q", t.Text)
				return nil, err
			}

		case scan.Comma, scan.Newline, scan.Comment:
			e = stack.Nop
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

func (p *Parser) number(s string) (result stack.Expr, err error) {
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

		return stack.Integer(uint(n)), nil
	}

	if strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B") ||
		strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return nil, fmt.Errorf("binary format invalid")
	}

	var f float64

	if f, err = strconv.ParseFloat(s, 64); err != nil {
		return nil, err
	}

	return stack.Number(f), nil
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

func (p *Parser) str(s string) (stack.Expr, error) {
	if len(s) == 0 {
		return nil, fmt.Errorf("empty string")
	}

	return stack.String(s), nil
}

func (p *Parser) operator(s string) (stack.Expr, error) {
	switch s {
	case "+":
		return stack.Add, nil
	case "*":
		return stack.Multiply, nil
	case "-":
		return stack.Subtract, nil
	case "/":
		return stack.Divide, nil
	case "%":
		return stack.Modulo, nil
	case "**":
		return stack.Power, nil
	case "!":
		return stack.Store, nil
	case "@":
		return stack.Recall, nil
	case "~":
		return stack.Not, nil
	case "&":
		return stack.And, nil
	case "|":
		return stack.Or, nil
	case "^":
		return stack.Xor, nil
	case "<<":
		return stack.LeftShift, nil
	case ">>":
		return stack.RightShift, nil
	case ">>>":
		return stack.ArithShift, nil
	}

	return nil, errUnknown
}

var (
	resultVar = regexp.MustCompile(`\$[0-9]+`)
	userVar   = regexp.MustCompile(`\$[a-zA-Z][a-zA-Z_0-9]*`)
)

func (p *Parser) symbol(s string) (stack.Expr, error) {
	if resultVar.MatchString(s) {
		return stack.GetSymbol(s), nil
	}

	if userVar.MatchString(s) {
		return stack.GetUserVar(s), nil
	}

	return nil, errUnknown
}

func (p *Parser) identifier(s string) (stack.Expr, error) {
	if e := stack.Predefined(s); e != nil {
		return e, nil
	}

	if p.machine.Known(s) {
		return p.machine.Word(s)
	}

	return p.machine.Builtin(s)
}

var errUnknown = errors.New("unknown")
