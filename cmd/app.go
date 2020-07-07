package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"oak"

	"github.com/chzyer/readline"
	"gopkg.in/yaml.v3"
)

const pname = "oak"

// runApp creates and runs the app given the arguments.
func runApp(args []string, version string, input io.Reader, output, errout io.Writer) int {
	a := app{
		stdIn:   input,
		stdOut:  output,
		errOut:  errout,
		version: version,
	}

	if err := a.fromArgs(args); err != nil {
		fmt.Fprintln(errout, err)
		return -2
	}

	if err := a.run(); err != nil {
		fmt.Fprintln(errout, err)
		return -1
	}

	return 0
}

type app struct {
	machine *oak.Machine
	stdIn   io.Reader
	stdOut  io.Writer
	errOut  io.Writer
	version string
	debug   bool
	fn      string
	config  string
	input   string
	image   string
	fixed   uint
	scip    uint
	engr    uint
	demo    bool
	radians bool
	show    bool
}

// fromArgs reads the flags and updates the app accordingly.
func (a *app) fromArgs(args []string) error {
	fl := flag.NewFlagSet(pname, flag.ContinueOnError)

	fl.StringVar(&a.fn, "f", "", "command file")
	fl.StringVar(&a.input, "e", "", "immediate commands")
	fl.StringVar(&a.image, "i", "", "saved image")
	fl.StringVar(&a.config, "c", ".oak.yml", "config file")
	fl.UintVar(&a.fixed, "fix", 0, "fixed precision")
	fl.UintVar(&a.scip, "sci", 0, "scientific precision")
	fl.UintVar(&a.engr, "eng", 0, "engineering mode")
	fl.BoolVar(&a.radians, "rad", false, "use radians mode")
	fl.BoolVar(&a.debug, "debug", false, "show parsing")
	fl.BoolVar(&a.demo, "demo", false, "run in demo mode")
	fl.BoolVar(&a.show, "version", false, "show version")

	if err := fl.Parse(args); err != nil {
		return err
	}

	if a.show {
		return fmt.Errorf("version %s", a.version)
	}

	a.machine = oak.New(a.stdOut)
	return nil
}

// run uses the flag settings to execute from readline or from
// a file (-f) or immediate expression list (-e) once the config
// has been read. If demo mode is selected, other flag or config
// settings are ignored.
func (a *app) run() error {
	if a.demo {
		if a.fn == "" {
			return fmt.Errorf("no demo file")
		}

		return a.fromFile(a.fn)
	}

	home, err := os.UserHomeDir()

	if err != nil {
		return fmt.Errorf("home: %s", err)
	}

	if err := a.readConfig(home); err != nil && err != io.EOF {
		return fmt.Errorf("config: %s", err)
	}

	if a.image != "" {
		if err := a.machine.LoadFromFile(a.image); err != nil {
			return fmt.Errorf("image: %s\n", err)
		}
	}

	if a.fixed > 0 {
		a.machine.SetFixed(a.fixed)
	} else if a.scip > 0 {
		a.machine.SetScientific(a.scip)
	} else if a.engr > 0 {
		a.machine.SetEngineering(a.engr)
	}

	if a.radians {
		a.machine.SetRadians()
	}

	if a.input != "" {
		return a.fromInput(a.stdOut, ioutil.NopCloser(bytes.NewBufferString(a.input)))
	}

	if a.fn != "" {
		return a.fromFile(a.fn)
	}

	return a.fromReadline(home)
}

// fromReadline runs the REPL and parses one line at a time.
func (a *app) fromReadline(home string) error {
	a.machine.SetInteractive()

	config := readline.Config{
		Stdin:                  readline.NewCancelableStdin(a.stdIn),
		Stdout:                 a.stdOut,
		Stderr:                 a.errOut,
		Prompt:                 "> ",
		HistoryFile:            path.Join(home, ".oakhist"),
		HistoryLimit:           50,
		DisableAutoSaveHistory: false,
		HistorySearchFold:      false,
	}

	il := 1
	rl, err := readline.NewEx(&config)

	if err != nil {
		return err
	}

	defer rl.Close()

	for {
		line, err := rl.Readline()

		if err != nil { // io.EOF
			break
		}

		c := oak.ScanConfig{Base: 0, Interactive: true}
		b := bytes.NewBufferString(line)
		s := oak.NewScanner(c, pname, b)
		p := oak.NewParser(a.machine, s, a.stdOut, il, a.debug)

		e, _, _ := p.Line()
		i, err := a.machine.Eval(il, e)

		if err == io.EOF { // bye
			return nil
		} else if err != nil {
			fmt.Fprintln(a.stdOut, err)
		} else {
			fmt.Fprintf(a.stdOut, "%d: %v\n", il, i)
		}

		il++
	}

	return nil
}

// fromFile collects the file and turns it into an immediate
// expression (so we can use that common code for all non-
// interactive runs).
func (a *app) fromFile(fn string) error {
	f, err := os.Open(fn)

	if err != nil {
		return err
	}

	return a.fromInput(a.stdOut, f)
}

// from Input processes a file or string from the command line
// (as well as config file commands), parsing the whole thing.
func (a *app) fromInput(w io.Writer, r io.ReadCloser) error {
	defer r.Close()

	c := oak.ScanConfig{}
	s := oak.NewScanner(c, pname, bufio.NewReader(r))
	p := oak.NewParser(a.machine, s, w, 1, a.debug)
	il := 1

	for {
		e, s, err := p.Line()

		if err != nil {
			return err
		}

		if len(e) == 0 {
			break
		}

		if a.demo {
			fmt.Fprintln(a.stdOut, ">", strings.TrimRight(s, "\n"))
		}

		if i, err := a.machine.Eval(il, e); err != nil {
			if err == io.EOF {
				return nil
			}

			fmt.Fprintln(w, err)
			return err
		} else {
			fmt.Fprintf(w, "%d: %v\n", il, i)
		}

		il++
	}

	return nil
}

// config represents what can be put into the YAML config file.
type config struct {
	Options  map[string]string
	Commands []string
}

// readConfig takes in .oak.yml and processes it silently
// (unless there's an error, which prints to stderr).
func (a *app) readConfig(home string) error {
	if path.Base(a.config) == a.config {
		a.config = path.Join(home, a.config)
	}

	file, err := os.Open(a.config)

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

	// all commands get joined into one expression list

	s := strings.Join(c.Commands, ", ")
	r := ioutil.NopCloser(bytes.NewBufferString(s))

	a.machine.SetOptions(c.Options)

	if err := a.fromInput(&b, r); err != nil {
		fmt.Fprintln(a.errOut, b.String())
		return fmt.Errorf("parsing %s: %s", a.config, err)
	}

	return nil
}
