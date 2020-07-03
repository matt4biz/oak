package scan

import (
	"unicode"
	"unicode/utf8"
)

// state functions

// lexComment scans a comment. The comment marker has been consumed.
func lexComment(l *Scanner) stateFn {
	for {
		r := l.next()

		if r == eof || r == '\n' {
			if l.config.Interactive {
				l.emit(EOF)
			}
			break
		}
	}

	if len(l.input) > 0 {
		l.pos = len(l.input)
		l.start = l.pos - 1

		// Emitting comment also advances l.line.
		l.emit(Comment) // but the text is lost
	}

	return lexSpace
}

// lexAny scans non-space items.
func lexAny(l *Scanner) stateFn {
	r := l.next()
	// fmt.Fprintf(os.Stderr, "state: get %q\n", r)

	switch {
	case r == eof:
		if l.config.Interactive {
			l.emit(EOF)
		}
		return nil

	case r == '\n': // TODO: \r
		l.emit(Newline)
		return lexAny

	case r == ',':
		l.emit(Comma)
		return lexAny

	case r == ':':
		l.emit(Colon)
		return lexAny

	case r == ';':
		l.emit(Semicolon)
		return lexAny

	case r == '#':
		return lexComment

	case isSpace(r):
		return lexSpace

	case r == '\'':
		return lexChar

	case r == '"':
		return lexQuote

	case r == '-':
		// It's an operator if it's preceded immediately (no spaces) by an operand, which is
		// an identifier, an indexed expression, or a parenthesized expression.
		// Otherwise it could be a signed number.
		if l.start > 0 {
			rr, _ := utf8.DecodeLastRuneInString(l.input[:l.start])
			if isAlphaNumeric(rr) || rr == ')' || rr == ']' {
				l.emit(Operator)
				return lexAny
			}
		}
		fallthrough

	case r == '.' || '0' <= r && r <= '9':
		l.backup()
		return lexNumber

	case r == '=':
		l.next()
		fallthrough // for ==

	case l.isOperator(r):
		// Must be after after = so == is an operator,
		// and after numbers, so '-' can be a sign.
		return lexOperator

	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier

	case r == '[':
		l.emit(LeftBracket)
		return lexAny

	case r == ']':
		l.emit(RightBracket)
		return lexAny

	case r == '{':
		l.emit(LeftBrace)
		return lexAny

	case r == '}':
		l.emit(RightBrace)
		return lexAny

	case r == '(':
		l.emit(LeftParen)
		return lexAny

	case r == ')':
		l.emit(RightParen)
		return lexAny

	case r <= unicode.MaxASCII && unicode.IsPrint(r):
		l.emit(Char)
		return lexAny

	default:
		return l.errorf("unrecognized character: %#U", r)
	}
}

// lexSpace scans a run of space characters,
// assuming one space has already been seen.
func lexSpace(l *Scanner) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}

	l.ignore()
	return lexAny
}

// lexIdentifier scans an alphanumeric.
// If the input base is greater than 10, some identifiers
// are actually numbers. We handle this here.
func lexIdentifier(l *Scanner) stateFn {
loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb

		default:
			l.backup()

			if !l.atTerminator() {
				return l.errorf("bad character %#U", r)
			}

			word := l.input[l.start:l.pos]

			switch {
			case isAllDigits(word, l.config.Base):
				l.emit(Number)
			default:
				l.emit(Identifier)
			}

			break loop
		}
	}

	return lexAny
}

// lexOperator completes scanning an operator. We have already accepted the + or
// whatever; there may be a reduction or inner or outer product.
func lexOperator(l *Scanner) stateFn {
	if isIdentifier(l.input[l.start:l.pos]) {
		l.emit(Identifier)
	} else {
		l.emit(Operator)
	}

	return lexSpace
}

// atTerminator reports whether the input is at valid
// termination character to appear after an identifier.
func (l *Scanner) atTerminator() bool {
	r := l.peek()

	if r == eof || isSpace(r) || isEndOfLine(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
		return true
	}

	return false
}

// lexChar scans a character constant. The initial quote is
// already scanned. Syntax checking is done by the parser.
func lexChar(l *Scanner) stateFn {
loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}

			fallthrough

		case eof, '\n':
			return l.errorf("unterminated character constant")

		case '\'':
			break loop
		}
	}

	l.emit(Char)
	return lexAny
}

// lexNumber scans a number: decimal, octal, hex, float, or imaginary. This
// isn't a perfect number scanner - for instance it accepts "." and "0x0.2"
// and "089" - but when it's wrong the input is invalid and the parser (via
// strconv) will notice.
func lexNumber(l *Scanner) stateFn {
	// Optional leading sign.
	if l.accept("-") {
		// Might not be a number.
		r := l.peek()

		if r != '.' && !l.isNumeral(r) {
			l.emit(Operator)
			return lexAny
		}
	}

	if !l.scanNumber() {
		return l.errorf("bad number syntax: %s", l.input[l.start:l.pos])
	}

	l.emit(Number)
	return lexAny
}

func (l *Scanner) scanNumber() bool {
	base := l.config.Base
	digits := digitsForBase(base)

	// If base 0, acccept binary for 0b / 0B or hex for 0x / 0X.
	// fmt.Fprintf(os.Stderr, "base=%d\n", base)

	if base == 0 {
		if l.accept("0") {
			if l.accept("xX") {
				digits = digitsForBase(16)
			} else if l.accept("bB") {
				digits = "01"
			}
		}
		// Otherwise leave it decimal (0); strconv.ParseInt will take care of it.
		// We can't set it to 8 in case it's a leading-0 float like 0.69 or 09e4.
	}
	// fmt.Fprintf(os.Stderr, "digits=%s\n", digits)

	l.acceptRun(digits)

	if l.accept(".") {
		l.acceptRun(digits)
	}

	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}

	r := l.peek()

	// Next thing mustn't be alphanumeric except possibly an o for outer product (3o.+2).
	if isAlphaNumeric(r) {
		l.next()
		return false
	}

	if r == '.' || !l.atTerminator() {
		l.next()
		return false
	}

	return true
}

var digits [16 + 1]string

const (
	decimal = "0123456789"
	lower   = "abcdef"
	upper   = "ABCDEF"
)

// digitsForBase returns the digit set for numbers in the specified base.
func digitsForBase(base int) string {
	if base == 0 {
		base = 10
	}

	d := digits[base]

	if d == "" {
		if base <= 10 {
			// Always accept a maximal string of numerals.
			// Whatever the input base, if it's <= 10 let the parser
			// decide if it's valid. This also helps us get the always-
			// base-10 numbers for )specials.
			d = decimal[:10]
		} else {
			d = decimal + lower[:base-10] + upper[:base-10]
		}

		digits[base] = d
	}

	return d
}

// lexQuote scans a quoted string.
func lexQuote(l *Scanner) stateFn {
loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough

		case eof, '\n':
			return l.errorf("unterminated quoted string")

		case '"':
			break loop
		}
	}

	l.emit(String)
	return lexAny
}
