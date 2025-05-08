package main

import (
	"strings"

	"github.com/leobel/pokedexcli/internal/termscanner"
)

type Repl struct {
	ts *termscanner.TermScanner
}

func NewRepl(ts *termscanner.TermScanner) *Repl {
	return &Repl{
		ts: ts,
	}
}

func (r *Repl) Init(cmds map[string]cliCommand) {

}

func (r *Repl) cleanInput(text string) []string {
	cleanText := strings.Trim(text, " ")
	return strings.Fields(strings.ToLower(cleanText))
}
