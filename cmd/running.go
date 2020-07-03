package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"oak/parse"
	"oak/scan"

	"github.com/chzyer/readline"
	"gopkg.in/yaml.v3"
)

// from Readline runs the REPL and
// parses one line at a time
func fromReadline(home string) {
	machine.SetInteractive()

	config := readline.Config{
		Prompt:                 "> ",
		HistoryFile:            path.Join(home, ".oakhist"),
		HistoryLimit:           50,
		DisableAutoSaveHistory: false,
		HistorySearchFold:      false,
	}

	il := 1
	rl, err := readline.NewEx(&config)

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

type config struct {
	Options  map[string]string
	Commands []string
}

// readConfig takes in the .oak.yml file and
// processes it silently (unless there's an
// error, in which case it prints to stderr)
func readConfig(home string) error {
	file, err := os.Open(path.Join(home, ".oak.yml"))

	// it's OK if the file isn't even there;
	// we just won't process it

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	defer file.Close()

	var b bytes.Buffer
	var c config

	err = yaml.NewDecoder(file).Decode(&c)

	if err != nil {
		return err
	}

	s := strings.Join(c.Commands, ", ")
	r := ioutil.NopCloser(bytes.NewBufferString(s))

	machine.SetOptions(c.Options)

	if erred := fromInput(&b, r); erred {
		fmt.Fprintln(os.Stderr, b.String())
		return fmt.Errorf("error parsing .oak.yml")
	}

	return nil
}
