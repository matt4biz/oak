package oak

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"oak/token"
)

// Config holds some configuration data for the scanner.
type ScanConfig struct {
	Base        int // TODO - remove; scanner needs to accept digits based on prefix only
	Line        int
	Interactive bool
}

// Scanner holds the state of the scanner.
type Scanner struct {
	tokens chan token.Token // channel of scanned items
	config ScanConfig
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
func NewScanner(c ScanConfig, name string, r io.ByteReader) *Scanner {
	if c.Line == 0 {
		c.Line = 1
	}

	l := Scanner{
		r:      r,
		name:   name,
		line:   c.Line,
		tokens: make(chan token.Token, 2), // We need a little room to save tokens.
		config: c,
		state:  lexAny,
	}

	return &l
}

// Next returns the next token.
func (l *Scanner) Next() token.Token {
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

	return token.Token{Type: token.EOF, Line: l.pos, Text: "EOFToken"}
}

// Line returns the current input line as a string for
// demo mode.
func (l *Scanner) Line() string {
	return l.input
}

const eof = -1

// stateFn represents the state of the scanner as a
// function that returns the next state.
type stateFn func(*Scanner) stateFn

// loadLine reads the next line of input and stores it
// in (appends it to) the input (l.input may have data
// left over when we are called).
//
// It strips carriage returns to make subsequent processing
// simpler.
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
func (l *Scanner) emit(t token.Type) {
	s := l.input[l.start:l.pos]

	if t == token.Newline || t == token.Comma || t == token.Comment {
		s = ""
	}

	// config := l.context.Config()
	//
	// if config.Debug("tokens") {
	// fmt.Fprintf(os.Stderr, "%s:%d: emit %s\n", l.name, l.line, Token{t, l.line, s})
	// }

	l.tokens <- token.Token{Type: t, Line: l.line, Text: s}
	l.start = l.pos
	l.width = 0

	if t == token.Newline || t == token.Comma || t == token.Comment {
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
	l.tokens <- token.Token{Type: token.Error, Line: l.line, Text: fmt.Sprintf(format, args...)}
	l.start = 0
	l.pos = 0
	l.input = "\n"

	return lexAny
}
