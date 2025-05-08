package commands

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/leobel/pokedexcli/internal/pokeapi"
	"github.com/leobel/pokedexcli/internal/pokecache"
)

type PokemonCatcher interface {
	TryToCatch(pokemon pokeapi.Pokemon) bool
}

type PokedexPokemonCatcher struct {
	lambda float64
}

type CatcherOption interface {
	apply(cp *CommandPokedex)
}

type PokemonCatcherOption struct {
	catcher PokemonCatcher
}

func (c PokemonCatcherOption) apply(cp *CommandPokedex) {
	cp.Catcher = c.catcher
}

func WithPokemonCatcher(catcher PokemonCatcher) CatcherOption {
	return PokemonCatcherOption{catcher}
}

func (c PokedexPokemonCatcher) TryToCatch(pokemon pokeapi.Pokemon) bool {
	experience := pokemon.BaseExperience
	return rand.Float64() < math.Exp(-c.lambda*float64(experience))
}

type CommandPokedex struct {
	Pokemons map[string]pokeapi.Pokemon
	Api      pokeapi.Api
	Cache    pokecache.Cache
	Catcher  PokemonCatcher
}

func NewCommandPokedex(api pokeapi.Api, cache pokecache.Cache, catcherOpts ...CatcherOption) *CommandPokedex {
	pokedex := &CommandPokedex{
		Pokemons: map[string]pokeapi.Pokemon{},
		Api:      api,
		Cache:    cache,
		Catcher:  PokedexPokemonCatcher{lambda: 0.005},
	}
	for _, opt := range catcherOpts {
		opt.apply(pokedex)
	}

	return pokedex
}

func (c *CommandPokedex) ShowPokemons(...string) error {
	fmt.Println("Your Pokedex:")
	for _, pokemon := range c.Pokemons {
		fmt.Printf(" - %s\n", pokemon.Name)
	}
	return nil
}

func (c *CommandPokedex) CatchPokemon(params ...string) error {
	name := params[0]
	fmt.Printf("Throwing a Pokeball at %s...\n", name)
	pokemon, err := c.Api.GetPokemon(name, c.Cache)
	if err != nil {
		return err
	}
	if c.Catcher.TryToCatch(*pokemon) {
		c.Pokemons[name] = *pokemon
		fmt.Printf("%s was caught!\n", name)
		fmt.Println("You may now inspect it with the inspect command.")
	} else {
		fmt.Printf("%s escaped!\n", name)
	}
	return nil
}

func (c *CommandPokedex) InspectPokemon(params ...string) error {
	name := params[0]
	pokemon, ok := c.Pokemons[name]
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
