package parse

import (
	"errors"
	"fmt"
	"math"
	"os"
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
	line    int
	inter   bool
	debug   bool
}

func New(m *stack.Machine, s *scan.Scanner, line int, inter, debug bool) *Parser {
	p := Parser{
		machine: m,
		scanner: s,
		line:    line,
		inter:   inter,
		debug:   debug,
	}

	return &p
}

func (p *Parser) Line() []stack.Expr {
	output := func() {
		if len(p.tokens) > 0 {
			if p.debug {
				fmt.Printf("%d: %s\n", p.line, p.tokens)
			}
			p.line++
		}
	}

	if !p.readTokensToNewline() {
		return nil
	}

	output()
	return p.evaluate()
}

func (p *Parser) readTokensToNewline() bool {
	p.tokens = p.buff[:0]

	for {
		tok := p.scanner.Next()

		switch tok.Type {
		case scan.Error:
			p.errorf("%s", tok)

		case scan.Newline:
			return true

		case scan.EOF:
			return len(p.tokens) > 0
		}

		p.tokens = append(p.tokens, tok)
	}
}

func (p *Parser) evaluate() []stack.Expr {
	var (
		result []stack.Expr
		e      stack.Expr
		err    error
	)

	for _, t := range p.tokens {
		switch t.Type {
		case scan.Number:
			if e, err = p.number(t.Text); err != nil {
				p.errorf("bad operator: %s: %w", t.Text, err)
				return nil
			}

		case scan.Operator:
			if e, err = p.operator(t.Text); err != nil {
				p.errorf("bad operator: %s", t.Text)
				return nil
			}

		case scan.Identifier:
			if strings.HasPrefix(t.Text, "$") {
				if e, err = p.symbol(t.Text); err != nil {
					p.errorf("bad symbol: %s", t.Text)
					return nil
				}
			} else {
				if e, err = p.identifier(t.Text); err != nil {
					p.errorf("unknown name: %s", t.Text)
					return nil
				}
			}

		case scan.String:
			if e, err = p.str(t.Text); err != nil {
				p.errorf("invalid string: %q", t.Text)
				return nil
			}
		}

		if e != nil {
			result = append(result, e)
		}
	}

	return result
}

func (p *Parser) errorf(format string, args ...interface{}) {
	p.tokens = p.buff[:0]

	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (p *Parser) number(s string) (stack.Expr, error) {
	f, err := strconv.ParseFloat(s, 64)

	if err != nil {
		return nil, err
	}

	return stack.Number(f), nil
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
		return stack.BinaryOp("add", func(y, x float64) float64 { return y + x }), nil
	case "*":
		return stack.BinaryOp("mul", func(y, x float64) float64 { return y * x }), nil
	case "-":
		return stack.BinaryOp("sub", func(y, x float64) float64 { return y - x }), nil
	case "/":
		return stack.BinaryOp("div", func(y, x float64) float64 { return y / x }), nil
	case "%":
		return stack.BinaryOp("div", func(y, x float64) float64 { return math.Mod(y, x) }), nil
	case "**":
		return stack.BinaryOp("sub", func(y, x float64) float64 { return math.Pow(y, x) }), nil
	}

	return nil, errUnknown
}

var dollarVar = regexp.MustCompile(`\$[0-9]+`)

func (p *Parser) symbol(s string) (stack.Expr, error) {
	if dollarVar.MatchString(s) {
		return stack.GetSymbol(s), nil
	}

	// TODO - add support for named vars + sto/rcl
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
