package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/leobel/pokedexcli/internal/pokeapi"
	"github.com/leobel/pokedexcli/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config, ...string) error
}

type LocationAreaDirection int

const (
	Previous LocationAreaDirection = iota
	Next
)

type config struct {
	Dir      LocationAreaDirection
	Previous *string
	Next     *string
}

var supportedCommands map[string]cliCommand

var initialUrl = "https://pokeapi.co/api/v2/location-area?offset=0limit=20"

var cache = pokecache.NewCache(10 * time.Second)

var capturedPokemons = map[string]pokeapi.Pokemon{}

func main() {
	config := config{
		Next: &initialUrl,
	}
	supportedCommands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Display next 20 location areas of the Pokemon world",
			callback:    commandMapNext,
		},
		"mapb": {
			name:        "mapb",
			description: "Display previous 20 location areas of the Pokemon world",
			callback:    commandMapPrevious,
		},
		"explore": {
			name:        "explore",
			description: "List of all the Pokémon located in a specific area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Trying to catch a Pokemon by name",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Show name, height, weight, stats and type(s) of Pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Show all Pokemon you've caught so far",
			callback:    commandPokedex,
		},
	}
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Pokedex > ")
	for scanner.Scan() {
		text := scanner.Text()
		inputs := cleanInput(text)
		if len(inputs) > 0 {
			cmd := inputs[0]
			cli, ok := supportedCommands[cmd]
			if ok {
				if err := cli.callback(&config, inputs[1:]...); err != nil {
					fmt.Println(err)
					os.Exit(0)
				}
			} else {
				fmt.Println("Unknown command")
			}
		}
		fmt.Print("Pokedex > ")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading from input:", err)
	}
}

func commandPokedex(*config, ...string) error {
	fmt.Println("Your Pokedex:")
	for _, pokemon := range capturedPokemons {
		fmt.Printf(" - %s\n", pokemon.Name)
	}
	return nil
}

func commandInspect(_ *config, params ...string) error {
	name := params[0]
	pokemon, ok := capturedPokemons[name]
	if !ok {
		fmt.Println("you have not caught that pokemon")
	} else {
		fmt.Printf("Name: %s\n", pokemon.Name)
		fmt.Printf("Height: %d\n", pokemon.Height)
		fmt.Printf("Weight: %d\n", pokemon.Weight)
		fmt.Println("Stats:")
		for _, stat := range pokemon.Stats {
			fmt.Printf(" -%s: %d\n", stat.Stat.Name, stat.BaseStat)
		}
		fmt.Println("Types:")
		for _, t := range pokemon.Types {
			fmt.Printf(" - %s\n", t.Type.Name)
		}
	}
	return nil
}

func commandCatch(_ *config, params ...string) error {
	name := params[0]
	fmt.Printf("Throwing a Pokeball at %s...\n", name)
	pokemon, err := pokeapi.GetPokemon(name, cache)
	if err != nil {
		return err
	}
	if tryToCatchPokemon(pokemon.BaseExperience, 0.005) {
		capturedPokemons[name] = *pokemon
		fmt.Printf("%s was caught!\n", name)
		fmt.Println("You may now inspect it with the inspect command.")
	} else {
		fmt.Printf("%s escaped!\n", name)
	}
	return nil
}

func tryToCatchPokemon(experience int, lambda float64) bool {
	return rand.Float64() < math.Exp(-lambda*float64(experience))
}

func commandExplore(_ *config, params ...string) error {
	if len(params) == 0 {
		return errors.New("invalid: no area to explore")
	}
	area := params[0]
	fmt.Println("Exploring pastoria-city-area...")
	response, err := pokeapi.GetLocationAreaDetails(area, cache)
	if err != nil {
		return err
	}
	fmt.Println("Found Pokemon:")
	for _, encounter := range response.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}

func commandMapNext(c *config, _ ...string) error {
	c.Dir = Next
	if c.Next != nil {
		return commandMap(c)
	} else {
		fmt.Println("you're on the last page, consider using command: `mapb` (map back) to display previous 20 locations")
		return nil
	}
}

func commandMapPrevious(c *config, _ ...string) error {
	c.Dir = Previous
	if c.Previous != nil {
		return commandMap(c)
	} else {
		fmt.Println("you're on the first page, consider using command: `map` (map forward) to display next 20 locations")
		return nil
	}
}

func commandMap(c *config) error {
	var url *string
	if c.Dir == Previous {
		url = c.Previous
	} else {
		url = c.Next
	}
	response, err := pokeapi.GetLocationArea(*url, cache)
	if err != nil {
		return err
	}
	c.Next, c.Previous = response.Next, response.Previous
	for _, area := range response.Results {
		fmt.Println(area.Name)
	}

	return nil
}

func commandHelp(*config, ...string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for key, value := range supportedCommands {
		fmt.Printf("%s: %s\n", key, value.description)
	}
	return nil
}

func commandExit(*config, ...string) error {
	cache.Stop()
	return errors.New("Closing the Pokedex... Goodbye!")
}

func cleanInput(text string) []string {
	cleanText := strings.Trim(text, " ")
	return strings.Fields(strings.ToLower(cleanText))
}
