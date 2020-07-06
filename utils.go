package oak

import "unicode"

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line (really end-of-statement) character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isIdentifier reports whether the string is a valid identifier.
func isIdentifier(s string) bool {
	if s == "_" {
		return false // Special symbol; can't redefine.
	}

	first := true

	for _, r := range s {
		if unicode.IsDigit(r) {
			if first {
				return false
			}
		} else if r == '$' {
			if !first {
				return false
			}
		} else if r != '_' && !unicode.IsLetter(r) {
			return false
		}

		first = false
	}

	return true
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore
// or variable marker $ (must be leading ... see above for the check).
func isAlphaNumeric(r rune) bool {
	return r == '_' || r == '$' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isNumeral reports whether r is a numeral in the input base.
// A decimal digit is always taken as a numeral, because otherwise parsing
// would be muddled. (In base 8, 039 shouldn't be scanned as two numbers.)
// The parser will check that the scanned number is legal.
func (l *Scanner) isNumeral(r rune) bool {
	if '0' <= r && r <= '9' {
		return true
	}

	base := l.config.Base

	if base < 10 {
		return false
	}

	top := rune(base - 10)

	if 'a' <= r && r <= 'a'+top {
		return true
	}

	if 'A' <= r && r <= 'A'+top {
		return true
	}

	return false
}

// isAllDigits reports whether s consists of digits in the specified base.
func isAllDigits(s string, base int) bool {
	top := 'a' + rune(base-10) - 1
	TOP := 'A' + rune(base-10) - 1

	for _, c := range s {
		if '0' <= c && c <= '9' {
			continue
		}

		if 'a' <= c && c <= top {
			continue
		}

		if 'A' <= c && c <= TOP {
			continue
		}

		return false
	}

	return true
}

// isOperator reports whether r is an operator. It may advance the lexer one character
// if it is a two-character operator.
func (l *Scanner) isOperator(r rune) bool {
	switch r {
	case '~', '@', '#', '%', '?', '+', '-', '/', '|', '^':
		// No follow-on possible.

	case '!':
		if l.peek() == '=' {
			l.next()
		}

	case '&':
		if l.peek() == '^' {
			l.next()
		}

	case '>':
		switch l.peek() {
		case '>', '=':
			l.next()
		}

	case '<':
		switch l.peek() {
		case '<', '=':
			l.next()
		}

	case '*':
		switch l.peek() {
		case '*':
			l.next()
		}

	case '=':
		if l.peek() != '=' {
			return false
		}
		l.next()

	case '∑':
		switch l.peek() {
		case '+', '-':
			l.next()
		}

	case '\\':
		switch l.peek() {
		case '+', '*':
			l.next()
		}

	default:
		return false
	}

	return true
}
