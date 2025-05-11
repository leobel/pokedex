package commands

import (
	"fmt"
	"maps"
	"slices"

	"github.com/leobel/pokedexcli/internal/repl"
)

type CommandHelp struct {
	commands *map[string]repl.CliCommand
}

func NewCommandHelp(commands *map[string]repl.CliCommand) *CommandHelp {
	return &CommandHelp{commands}
}

func (c *CommandHelp) Help(...string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("")
	cmds := *c.commands
	keys := slices.Collect(maps.Keys(cmds))
	slices.Sort(keys)
	for _, key := range keys {
		fmt.Printf("%s: %s\n", key, cmds[key].Description)
	}
	fmt.Println("Up/Down keys: Use it to navigate between previous and next commands")
	return nil
}
