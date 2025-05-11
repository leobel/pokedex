package commands

import (
	"errors"

	"github.com/leobel/pokedexcli/internal/pokecache"
)

type CommandExit[T pokecache.Cache] struct {
	Cache pokecache.Cache
}

func NewCommandExit[T pokecache.Cache](cache T) *CommandExit[T] {
	return &CommandExit[T]{cache}
}

func (c *CommandExit[T]) Exit(...string) error {
	c.Cache.Stop()
	return errors.New("Closing the Pokedex... Goodbye!")
}
