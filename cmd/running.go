package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"oak/parse"
	"oak/scan"

	"github.com/chzyer/readline"
)

// from Readline runs the REPL and
// parses one line at a time
func fromReadline() {
	il := 1
	rl, err := readline.New("> ")

	if err != nil {
		panic(err)
	}

	defer rl.Close()

	for {
		line, err := rl.Readline()

		if err != nil { // io.EOF
			break
		}

		c := scan.Config{Base: 0, Interactive: true}
		b := bytes.NewBufferString(line)
		s := scan.New(c, pname, b)
		p := parse.New(machine, s, os.Stdout, il, debug)
		e, _ := p.Line()

		if i, err := machine.Eval(il, e); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%d: %v\n", il, i)
		}

		il++
	}
}

// from Input processes a file or string from
// the command line, parsing the whole thing
func fromInput(w io.Writer, r io.ReadCloser) (erred bool) {
	defer r.Close()

	c := scan.Config{}
	s := scan.New(c, pname, bufio.NewReader(r))
	p := parse.New(machine, s, w, 1, debug)
	il := 1

	for {
		e, err := p.Line()

		if err != nil {
			erred = true
		}

		if len(e) == 0 {
			break
		}

		if i, err := machine.Eval(il, e); err != nil {
			fmt.Fprintln(w, err)
			erred = true
		} else {
			fmt.Fprintf(w, "%d: %v\n", il, i)
		}

		il++
	}

	return
}

// readRuncom takes in the .oakrc file and
// processes it silently (unless there's an
// error, in which case it prints to stderr)
func readRuncom() error {
	var b bytes.Buffer

	home, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	file, err := os.Open(path.Join(home, ".oakrc"))

	// it's OK if the file isn't even there;
	// we just won't process it

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	defer file.Close()

	if erred := fromInput(&b, file); erred {
		fmt.Fprint(os.Stderr, "parsing .oakrc: ", b.String())
		return fmt.Errorf("parse error")
	}

	return nil
}
