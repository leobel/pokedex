package commands

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/leobel/pokedexcli/internal/pokeapi"
	"github.com/leobel/pokedexcli/internal/pokecache"
)

type CommandMap[T pokecache.Cache] struct {
	Previous *string
	Next     *string
	Api      pokeapi.Api[T]
}

func NewCommandMap[T pokecache.Cache](api pokeapi.Api[T]) *CommandMap[T] {
	next := fmt.Sprintf("%s/location-area?offset=%d&limit=%d", api.GetBaseUrl(), 0, api.GetConfig().Limit)
	return &CommandMap[T]{
		Next: &next,
		Api:  api,
	}
}

func (c *CommandMap[T]) PreviousArea() func(...string) error {
	return func(...string) error {
		if c.Previous != nil {
			return c.request(*c.Previous)
		} else {
			fmt.Println("you're on the first page, consider using command: `map` (map forward) to display next 20 locations")
			return nil
		}
	}
}

func (c *CommandMap[T]) NextArea() func(...string) error {
	return func(...string) error {
		if c.Next != nil {
			return c.request(*c.Next)
		} else {
			fmt.Println("you're on the last page, consider using command: `mapb` (map back) to display previous 20 locations")
			return nil
		}
	}
}

func (c *CommandMap[T]) ExploreArea(params ...string) error {
	if len(params) == 0 {
		return errors.New("invalid: no area to explore")
	}
	area := params[0]
	fmt.Println("Exploring pastoria-city-area...")
	response, err := c.Api.GetLocationAreaDetails(area)
	if err != nil {
		return err
	}
	fmt.Println("Found Pokemon:")
	for _, encounter := range response.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}

func (c *CommandMap[T]) request(rawUrl string) error {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		panic(err)
	}
	query := parsedUrl.Query()
	offset, err := strconv.Atoi(query.Get("offset"))
	if err != nil {
		return err
	}
	response, err := c.Api.GetLocationArea(offset)
	if err != nil {
		return err
	}
	c.Next, c.Previous = response.Next, response.Previous
	for _, area := range response.Results {
		fmt.Println(area.Name)
	}

	return nil
}
