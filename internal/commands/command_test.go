package commands_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/leobel/pokedexcli/internal/commands"
	"github.com/leobel/pokedexcli/internal/pokeapi"
	"github.com/leobel/pokedexcli/internal/pokecache"
	"github.com/leobel/pokedexcli/internal/repl"
)

// --- Mock Cache ---
type mockCache struct {
	store map[string][]byte
}

func newMockCache() *mockCache {
	return &mockCache{store: map[string][]byte{}}
}
func (c *mockCache) Get(key string) ([]byte, bool) {
	v, ok := c.store[key]
	return v, ok
}
func (c *mockCache) Add(key string, val []byte) {
	c.store[key] = val
}
func (c *mockCache) Stop() {
	c.store = map[string][]byte{}
}

// --- Mock API ---
type mockApi struct {
	baseUrl                 string
	config                  pokeapi.Config
	getPokemonResponse      *pokeapi.Pokemon
	getLocationDetailsError error
	locationDetailsResp     *pokeapi.LocationAreaDetailsResponse
	locationAreaResponses   map[int]*pokeapi.LocationAreaResponse
}

func newMockApi(url string, config pokeapi.Config) *mockApi {
	return &mockApi{
		baseUrl:               url,
		config:                config,
		locationAreaResponses: map[int]*pokeapi.LocationAreaResponse{},
	}
}

func (api *mockApi) GetBaseUrl() string {
	return api.baseUrl
}

func (api *mockApi) GetConfig() pokeapi.Config {
	return api.config
}

func (m *mockApi) GetPokemon(name string, cache pokecache.Cache) (*pokeapi.Pokemon, error) {
	return m.getPokemonResponse, nil
}

func (m *mockApi) GetLocationAreaDetails(area string, cache pokecache.Cache) (*pokeapi.LocationAreaDetailsResponse, error) {
	return m.locationDetailsResp, m.getLocationDetailsError
}

func (m *mockApi) GetLocationArea(offset int, cache pokecache.Cache) (*pokeapi.LocationAreaResponse, error) {
	if resp, ok := m.locationAreaResponses[offset]; ok {
		return resp, nil
	}
	return nil, errors.New("not found")
}

type AlwaysCatch struct{}

func (c AlwaysCatch) TryToCatch(pokeapi.Pokemon) bool {
	return true
}

// --- Helpers to capture stdout ---
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestCommandExit(t *testing.T) {
	cache := newMockCache()
	cmd := commands.NewCommandExit(cache)

	err := cmd.Exit()
	if err == nil || !strings.Contains(err.Error(), "Closing the Pokedex") {
		t.Fatalf("Exit() error = %v; want closing message", err)
	}
	// Stop should clear store
	cache.store["x"] = []byte("y")
	cmd.Exit()
	if len(cache.store) != 0 {
		t.Errorf("Exit did not clear cache")
	}
}

func TestCommandHelp(t *testing.T) {
	// prepare a fake command map
	cmds := map[string]repl.CliCommand{
		"a": {Name: "a", Description: "descA"},
		"b": {Name: "b", Description: "descB"},
	}
	h := commands.NewCommandHelp(&cmds)

	out := captureStdout(func() {
		if err := h.Help(); err != nil {
			t.Fatal(err)
		}
	})

	// check ordering and content
	if !strings.Contains(out, "Welcome to the Pokedex!") {
		t.Error("Help output missing header")
	}
	// because keys sorted ["a","b"]
	if !strings.Contains(out, "a: descA\nb: descB") {
		t.Errorf("Help output wrong ordering or content: %q", out)
	}
}

func TestCommandMapNextPrevious(t *testing.T) {
	cache := newMockCache()
	api := newMockApi("url", pokeapi.Config{})

	// simulate two pages
	api.locationAreaResponses[0] = &pokeapi.LocationAreaResponse{
		Next:     ptrString("url?offset=1"),
		Previous: nil,
		Results: []struct {
			Name string `json:"name"`
			Url  string `json:"url"`
		}{
			{
				Name: "foo",
			},
		},
	}
	api.locationAreaResponses[1] = &pokeapi.LocationAreaResponse{
		Next:     ptrString("url?offset=2"),
		Previous: ptrString("url?offset=0"),
		Results: []struct {
			Name string `json:"name"`
			Url  string `json:"url"`
		}{
			{
				Name: "bar",
			},
		},
	}

	// build and test forward
	cm := commands.NewCommandMap(api, cache)
	out := captureStdout(func() {
		if err := cm.NextArea()(); err != nil {
			t.Fatal(err)
		}
	})

	// check ordering and content
	if !strings.Contains(out, "foo") {
		t.Error("Map output missing foo area")
	}

	out = captureStdout(func() {
		if err := cm.NextArea()(); err != nil {
			t.Fatal(err)
		}
	})
	// check ordering and content
	if !strings.Contains(out, "bar") {
		t.Error("Map output missing bar area")
	}

	out = captureStdout(func() {
		if err := cm.PreviousArea()(); err != nil {
			t.Fatal(err)
		}
	})

	// check ordering and content
	if !strings.Contains(out, "foo") {
		t.Error("Map output missing foo area")
	}

}

func TestCommandMapExplore(t *testing.T) {
	cache := newMockCache()
	api := newMockApi("url", pokeapi.Config{})
	api.locationDetailsResp = &pokeapi.LocationAreaDetailsResponse{
		PokemonEncounters: []pokeapi.PokemonEncounters{
			{
				Pokemon: struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				}{
					Name: "Pikachu",
				},
			},
		},
	}

	cm := commands.NewCommandMap(api, cache)

	out := captureStdout(func() {
		if err := cm.ExploreArea("some-area"); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "Pikachu") {
		t.Error("ExploreArea did not list Pikachu")
	}
	// error on no param
	if err := cm.ExploreArea(); err == nil {
		t.Error("ExploreArea should error on missing param")
	}
}

func TestCommandPokedex_ShowInspectCatch(t *testing.T) {
	cache := newMockCache()
	api := newMockApi("url", pokeapi.Config{})
	api.getPokemonResponse = &pokeapi.Pokemon{
		Name:           "Pikachu",
		BaseExperience: 112,
		Height:         4,
		Weight:         60,
		Stats: []struct {
			BaseStat int `json:"base_stat"`
			Effort   int `json:"effort"`
			Stat     struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"stat"`
		}{
			{
				BaseStat: 50,
				Stat: struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				}{
					Name: "hp",
				},
			},
		},
		Types: []struct {
			Slot int `json:"slot"`
			Type struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"type"`
		}{
			{
				Type: struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				}{
					Name: "electric",
				},
			},
		},
	}

	cp := commands.NewCommandPokedex(api, cache, commands.WithPokemonCatcher(AlwaysCatch{}))

	// show empty first
	out0 := captureStdout(func() {
		cp.ShowPokemons()
		if err := cp.ShowPokemons(); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out0, "Your Pokedex:") {
		t.Error("ShowPokemons missing header")
	}

	// catch success (lambda small so success guaranteed)
	out1 := captureStdout(func() {
		if err := cp.CatchPokemon("Pikachu"); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out1, "Pikachu was caught") {
		t.Error("CatchPokemon did not report catch")
	}

	// inspect caught
	out2 := captureStdout(func() {
		if err := cp.InspectPokemon("Pikachu"); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out2, "Name: Pikachu") {
		t.Error("InspectPokemon did not print details")
	}

	// inspect missing
	out3 := captureStdout(func() {
		if err := cp.InspectPokemon("Missing"); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out3, "not caught") {
		t.Error("InspectPokemon should warn on missing")
	}
}

// helper to get *string
func ptrString(s string) *string { return &s }
