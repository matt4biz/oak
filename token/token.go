package token

import (
	"fmt"
	"strconv"
)

// Type identifies the type of lex items.
type Type int

const (
	EOF   Type = iota // zero value so closed channel delivers EOFToken
	Error             // error occurred; value is text of error
	Newline
	Comment
	// Interesting things
	Char         // printable ASCII character
	Colon        // ':'
	Comma        // ','
	Identifier   // alphanumeric identifier
	LeftBracket  // '['
	LeftBrace    // '{'
	LeftParen    // '('
	Number       // number (as float64)
	Operator     // known operator
	RightBracket // ']'
	RightBrace   // '}'
	RightParen   // ')'
	Semicolon    // ';'
	Space        // run of spaces separating
	String       // quoted string (includes quotes)
	Variable     // $+number or $+identifier
)

// Token represents a token or text string returned from the scanner.
type Token struct {
	Type Type   // The type of this item.
	Line int    // The line number on which this token appears
	Text string // The text of this item.
}

func (t Type) String() string {
	switch t {
	case EOF:
		return "EOF"
	case Error:
		return "Error"
	case Newline:
		return "Newline"
	case Comment:
		return "Comment"
	case Char:
		return "Char"
	case Colon:
		return "Colon"
	case Comma:
		return "Comma"
	case Identifier:
		return "Identifier"
	case LeftBracket:
		return "LeftBracket"
	case LeftBrace:
		return "LeftBrace"
	case LeftParen:
		return "LeftParen"
	case Number:
		return "Number"
	case Operator:
		return "Operator"
	case RightBracket:
		return "RightBracket"
	case RightBrace:
		return "RightBrace"
	case RightParen:
		return "RightParen"
	case Semicolon:
		return "Semicolon"
	case Space:
		return "Space"
	case String:
		return "String"
	}

	return "UNKNOWN[" + strconv.Itoa(int(t)) + "]"
}

func (i Token) String() string {
	switch {
	case i.Type == EOF:
		return "EOF"

	case i.Type == Error:
		return "error: " + i.Text

	case len(i.Text) > 10:
		return fmt.Sprintf("[%d] %s: %.10q...", i.Line, i.Type, i.Text)
	}

	return fmt.Sprintf("[%d] %s: %q", i.Line, i.Type, i.Text)
}
