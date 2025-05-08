package repl

import (
	"fmt"
	"os"
	"strings"

	"github.com/leobel/pokedexcli/internal/termscanner"
)

type Repl struct {
	Scanner termscanner.PokedexScanner
}

type CliCommand struct {
	Name        string
	Description string
	Callback    func(...string) error
}

func NewRepl(scanner termscanner.PokedexScanner) *Repl {
	return &Repl{scanner}
}

func (r *Repl) Init(cmds map[string]CliCommand) {
	for r.Scanner.Scan() {
		text := r.Scanner.Text()
		inputs := r.CleanInput(text)
		if len(inputs) > 0 {
			cmd := inputs[0]
			cli, ok := cmds[cmd]
			if ok {
				if err := cli.Callback(inputs[1:]...); err != nil {
					fmt.Println(err)
					os.Exit(0)
				}
			} else {
				fmt.Println("Unknown command")
			}
		}
	}
	if err := r.Scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading from input:", err)
	}
	err := cmds["exit"].Callback()
	fmt.Println(err)
	os.Exit(0)
}

func (r *Repl) CleanInput(text string) []string {
	cleanText := strings.Trim(text, " ")
	return strings.Fields(strings.ToLower(cleanText))
}
