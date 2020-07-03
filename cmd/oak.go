package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"oak/stack"
)

const pname = "oak"

var (
	machine = stack.New()
	debug   bool
	version string // do not modify or remove
)

func main() {
	var (
		fn      string
		input   string
		image   string
		fixed   uint
		scip    uint
		engr    uint
		radians bool
		show    bool
	)

	flag.StringVar(&fn, "f", "", "command file")
	flag.StringVar(&input, "e", "", "command text")
	flag.StringVar(&image, "i", "", "saved image")
	flag.UintVar(&fixed, "fix", 0, "fixed precision")
	flag.UintVar(&scip, "sci", 0, "scientific precision")
	flag.UintVar(&engr, "eng", 0, "engineering mode")
	flag.BoolVar(&radians, "rad", false, "use radians mode")
	flag.BoolVar(&debug, "debug", false, "show parsing")
	flag.BoolVar(&show, "version", false, "show version")
	flag.Parse()

	if show {
		fmt.Fprintln(os.Stderr, "version", version)
		return
	}

	if fixed > 0 {
		machine.SetFixed(fixed)
	} else if scip > 0 {
		machine.SetScientific(scip)
	} else if engr > 0 {
		machine.SetEngineering(engr)
	}

	if radians {
		machine.SetRadians()
	}

	home, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintln(os.Stderr, "home:", err)
		os.Exit(-2)
	}

	if err := readConfig(home); err != nil {
		fmt.Fprintf(os.Stderr, "config: %s\n", err)
		os.Exit(-1)
	}

	if image != "" {
		if err := machine.LoadFromFile(image); err != nil {
			fmt.Fprintf(os.Stderr, "image: %s\n", err)
			os.Exit(-1)
		}
	}

	if input != "" {
		fromInput(os.Stdout, ioutil.NopCloser(bytes.NewBufferString(input)))
	} else if fn != "" {
		f, err := os.Open(fn)

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}

		fromInput(os.Stdout, f)
	} else {
		fromReadline(home)
	}
}
