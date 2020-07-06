package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var input = `# commands
1 2 +
3 +
bye
`

var options = `
options:
  display_mode: fix
  digits: 3
commands:
  - status
`

type subTest struct {
	name      string
	input     string
	options   string
	wanted    []string
	file      bool
	immediate bool
}

func (s subTest) run(t *testing.T) {
	config, err := ioutil.TempFile("", "*.yml")

	if err != nil {
		t.Fatalf("tmp file: %s", err)
	}

	_, err = config.WriteString(s.options)

	if err != nil {
		t.Fatalf("tmp write: %s", err)
	}

	config.Close()
	defer os.Remove(config.Name())

	args := []string{"-c", config.Name()}

	var input io.Reader

	if s.file {
		cmds, err := ioutil.TempFile(".", "*.oak")

		if err != nil {
			t.Fatalf("tmp file: %s", err)
		}

		_, err = cmds.WriteString(s.input)

		if err != nil {
			t.Fatalf("tmp write: %s", err)
		}

		cmds.Close()
		defer os.Remove(cmds.Name())

		args = append(args, "-f", cmds.Name())
	} else if s.immediate {
		args = append(args, "-e", s.input)
	} else {
		input = bytes.NewBufferString(s.input)
	}

	buff := new(bytes.Buffer)

	if c := runApp(args, "0.666", input, buff, buff); c != 0 {
		t.Fatalf("invalid return: %d: %s", c, buff.String())
	}

	for i := 0; i < len(s.wanted); i++ {
		got, err := buff.ReadString('\n')

		if err != nil {
			t.Errorf("read string; %s", err)
			break
		}

		if strings.TrimRight(got, "\n") != s.wanted[i] {
			t.Errorf("line %d: want %q, got %q", i, s.wanted[i], got)
		}
	}

}

func TestApp(t *testing.T) {
	table := []subTest{
		{
			name:    "fromFile",
			input:   input,
			options: options,
			file:    true,
			wanted: []string{
				"base: 10 mode: deg display: fix/3",
				"1: 3.000",
				"2: 6.000",
			},
		},
		{
			name:      "fromCLI",
			input:     "1 3+, 4+",
			options:   "",
			immediate: true,
			wanted: []string{
				"1: 4",
				"2: 8",
			},
		},
		{
			name: "fromReadLine",
			input: `2 5+
1+
`,
			options: "options:\n  digits: 1\n  display_mode: fix",
			wanted: []string{
				"1: 7.0",
				"2: 8.0",
			},
		},
	}

	for _, st := range table {
		t.Run(st.name, st.run)
	}
}
