package commands

import (
	"errors"

	"github.com/leobel/pokedexcli/internal/pokecache"
)

type CommandExit struct {
	Cache pokecache.Cache
}

func NewCommandExit(cache pokecache.Cache) *CommandExit {
	return &CommandExit{cache}
}

func (c *CommandExit) Exit(...string) error {
	c.Cache.Stop()
	return errors.New("Closing the Pokedex... Goodbye!")
}
