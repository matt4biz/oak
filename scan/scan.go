package scan

import (
	"fmt"
	"io"
	"strconv"

	// "os"
	"strings"
	"unicode/utf8"
)

// Type identifies the type of lex items.
type Type int

const (
	EOF   Type = iota // zero value so closed channel delivers EOF
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

// Config holds some configuration data for the scanner.
type Config struct {
	Base        int // TODO - remove; scanner needs to accept digits based on prefix only
	Line        int
	Interactive bool
}

// Scanner holds the state of the scanner.
type Scanner struct {
	tokens chan Token // channel of scanned items
	config Config
	r      io.ByteReader
	done   bool
	name   string // the name of the input; used only for error reports
	buf    []byte
	input  string  // the line of text being scanned.
	state  stateFn // the next lexing function to enter
	line   int     // line number in input
	pos    int     // current position in the input
	start  int     // start position of this item
	width  int     // width of last rune read from input
}

// New creates a new scanner for the input string.
func New(c Config, name string, r io.ByteReader) *Scanner {
	if c.Line == 0 {
		c.Line = 1
	}

	l := Scanner{
		r:      r,
		name:   name,
		line:   c.Line,
		tokens: make(chan Token, 2), // We need a little room to save tokens.
		config: c,
		state:  lexAny,
	}

	return &l
}

// Next returns the next token.
func (l *Scanner) Next() Token {
	// The lexer is concurrent but we don't want it to run in parallel
	// with the rest of the interpreter, so we only run the state machine
	// when we need a token.

	for l.state != nil {
		select {
		case tok := <-l.tokens:
			return tok

		default:
			// Run the machine
			l.state = l.state(l)
		}
	}

	if l.tokens != nil {
		close(l.tokens)
		l.tokens = nil
	}

	return Token{EOF, l.pos, "EOF"}
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

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*Scanner) stateFn

// loadLine reads the next line of input and stores it in (appends it to) the input.
// (l.input may have data left over when we are called.)
// It strips carriage returns to make subsequent processing simpler.
//
//
// NOTE: we'll need to tie this into the readline stuff ... or maybe we just
// make a scanner for each line by itself?
func (l *Scanner) loadLine() {
	l.buf = l.buf[:0]

	for {
		c, err := l.r.ReadByte()

		if err != nil {
			l.done = true
			break
		}

		if c != '\r' {
			l.buf = append(l.buf, c)
		}

		if c == '\n' {
			break
		}
	}

	l.input = l.input[l.start:l.pos] + string(l.buf)
	l.pos -= l.start
	l.start = 0
}

// next returns the next rune in the input.
func (l *Scanner) next() rune {
	if !l.done && l.pos == len(l.input) {
		l.loadLine()
	}

	if len(l.input) == l.pos {
		l.width = 0
		return eof
	}

	r, w := utf8.DecodeRuneInString(l.input[l.pos:])

	l.width = w
	l.pos += l.width

	return r
}

// peek returns but does not consume the next rune in the input.
func (l *Scanner) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune; it may
// only be called once per call of next.
func (l *Scanner) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *Scanner) emit(t Type) {
	s := l.input[l.start:l.pos]

	if t == Newline || t == Comma || t == Comment {
		s = ""
	}

	// config := l.context.Config()
	//
	// if config.Debug("tokens") {
	// fmt.Fprintf(os.Stderr, "%s:%d: emit %s\n", l.name, l.line, Token{t, l.line, s})
	// }

	l.tokens <- Token{t, l.line, s}
	l.start = l.pos
	l.width = 0

	if t == Newline || t == Comma || t == Comment {
		l.line++
	}
}

// ignore skips over the pending input before this point.
func (l *Scanner) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *Scanner) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}

	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *Scanner) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}

	l.backup()
}

// errorf returns an error token, replaces the input line with a
// newline (so the next token will be a newline, skipping the
// rest of the current line), and continues to scan.
func (l *Scanner) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- Token{Error, l.line, fmt.Sprintf(format, args...)}
	l.start = 0
	l.pos = 0
	l.input = "\n"

	return lexAny
}
