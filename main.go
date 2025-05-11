package main

import (
	"fmt"
	"os"
	"time"

	"github.com/leobel/pokedexcli/internal/commands"
	"github.com/leobel/pokedexcli/internal/pokeapi"
	"github.com/leobel/pokedexcli/internal/pokecache"
	"github.com/leobel/pokedexcli/internal/repl"
	"github.com/leobel/pokedexcli/internal/termscanner"
)

var supportedCommands map[string]repl.CliCommand

func main() {
	cache := pokecache.NewPokeCache(10 * time.Second)
	api := pokeapi.NewPokeApi("https://pokeapi.co/api/v2", cache)

	helpCmd := commands.NewCommandHelp(&supportedCommands)
	exitCmd := commands.NewCommandExit(api.Cache)
	mapCmd := commands.NewCommandMap[pokecache.Cache](api)
	pokedexCmd := commands.NewCommandPokedex[pokecache.Cache](api)

	supportedCommands = map[string]repl.CliCommand{
		"exit": {
			Name:        "exit",
			Description: "Exit the Pokedex",
			Callback:    exitCmd.Exit,
		},
		"help": {
			Name:        "help",
			Description: "Displays this help message",
			Callback:    helpCmd.Help,
		},
		"map": {
			Name:        "map",
			Description: fmt.Sprintf("Display next %d location areas of the Pokemon world", api.Config.Limit),
			Callback:    mapCmd.NextArea(),
		},
		"mapb": {
			Name:        "mapb",
			Description: fmt.Sprintf("Display previous %d location areas of the Pokemon world", api.Config.Limit),
			Callback:    mapCmd.PreviousArea(),
		},
		"explore": {
			Name:        "explore",
			Description: "List of all the Pokemons located in a specific area",
			Callback:    mapCmd.ExploreArea,
		},
		"catch": {
			Name:        "catch",
			Description: "Trying to catch a Pokemon by name",
			Callback:    pokedexCmd.CatchPokemon,
		},
		"inspect": {
			Name:        "inspect",
			Description: "Show name, height, weight, stats and type(s) of Pokemon",
			Callback:    pokedexCmd.InspectPokemon,
		},
		"pokedex": {
			Name:        "pokedex",
			Description: "Show all Pokemon you've caught so far",
			Callback:    pokedexCmd.ShowPokemons,
		},
	}

	scanner := termscanner.New("Pokedex > ", os.Stdin, termscanner.RealTerm{})

	cliRepl := repl.NewRepl(scanner)

	// init REPL cli
	cliRepl.Init(supportedCommands)
}
