package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/leobel/pokedexcli/internal/pokeapi"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
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

var initialUrl = "https://pokeapi.co/api/v2/location-area"

func main() {
	next := initialUrl
	config := config{
		Next: &next,
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
			description: "Display next 20 location areas of the Pokemom world",
			callback:    commandMapNext,
		},
		"mapb": {
			name:        "mapb",
			description: "Display previous 20 location areas of the Pokemom world",
			callback:    commandMapPrevious,
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
				if err := cli.callback(&config); err != nil {
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

func commandMapNext(c *config) error {
	c.Dir = Next
	if c.Next != nil {
		return commandMap(c)
	} else {
		fmt.Println("you're on the last page, consider using command: `mapb` (map back) to display previous 20 locations")
		return nil
	}
}

func commandMapPrevious(c *config) error {
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
	response, err := pokeapi.RequestApi(*url)
	if err != nil {
		return err
	}
	c.Next, c.Previous = response.Next, response.Previous
	for _, area := range response.Results {
		fmt.Println(area.Name)
	}

	return nil
}

func commandHelp(*config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for key, value := range supportedCommands {
		fmt.Printf("%s: %s\n", key, value.description)
	}
	return nil
}

func commandExit(*config) error {
	return errors.New("Closing the Pokedex... Goodbye!")
}

func cleanInput(text string) []string {
	cleanText := strings.Trim(text, " ")
	return strings.Fields(strings.ToLower(cleanText))
}
