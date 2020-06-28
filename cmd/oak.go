package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"oak/parse"
	"oak/scan"
	"oak/stack"

	"github.com/chzyer/readline"
)

const pname = "oak"

var (
	machine = stack.New()
	debug   bool
)

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
		p := parse.New(machine, s, il, true, debug)
		e := p.Line()

		if i, err := machine.Eval(il, e); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%d: %v\n", il, i)
		}

		il++
	}
}

func fromInput(r io.ReadCloser) {
	defer r.Close()

	c := scan.Config{}
	s := scan.New(c, pname, bufio.NewReader(r))
	p := parse.New(machine, s, 1, true, debug)
	il := 1

	for {
		e := p.Line()
		if len(e) == 0 {
			break
		}

		if i, err := machine.Eval(il, e); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%d: %v\n", il, i)
		}

		il++
	}
}

func main() {
	var (
		fn    string
		input string
		fixed int
		scip  int
	)

	flag.StringVar(&fn, "f", "", "command file")
	flag.StringVar(&input, "e", "", "command text")
	flag.IntVar(&fixed, "fix", 0, "fixed precision")
	flag.IntVar(&scip, "sci", 0, "scientific precision")
	flag.BoolVar(&debug, "debug", false, "show parsing")
	flag.Parse()

	if fixed > 0 {
		machine.SetFixed(fixed)
	} else if scip > 0 {
		machine.SetScientific(scip)
	}

	if input != "" {
		fromInput(ioutil.NopCloser(bytes.NewBufferString(input)))
	} else if fn != "" {
		f, err := os.Open(fn)

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}

		fromInput(f)
	} else {
		fromReadline()
	}
}
