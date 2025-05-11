package pokeapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leobel/pokedexcli/internal/pokeapi"
)

type MockCache struct {
	store map[string][]byte
}

func NewMockCache() *MockCache {
	return &MockCache{store: make(map[string][]byte)}
}

func (c *MockCache) Get(key string) ([]byte, bool) {
	val, ok := c.store[key]
	return val, ok
}

func (c *MockCache) Add(key string, val []byte) {
	c.store[key] = val
}

func (c *MockCache) Stop() {

}

func TestNewPokeApi(t *testing.T) {
	// arrange
	baseUrl := "https://pokeapi.co/api/v2"
	type ctr struct {
		baseUrl string
		cache   *MockCache
		opts    []pokeapi.Option
	}
	cache := NewMockCache()
	cases := []struct {
		actual   ctr
		expected pokeapi.PokeApi[*MockCache]
	}{
		{
			actual: ctr{
				baseUrl: baseUrl,
				cache:   cache,
				opts:    []pokeapi.Option{},
			},
			expected: pokeapi.PokeApi[*MockCache]{
				BaseUrl: baseUrl,
				Config:  pokeapi.Config{Limit: 20},
				Cache:   cache,
			},
		},
		{
			actual: ctr{
				baseUrl: baseUrl,
				cache:   cache,
				opts:    []pokeapi.Option{pokeapi.WithLimit(10)},
			},
			expected: pokeapi.PokeApi[*MockCache]{
				BaseUrl: baseUrl,
				Config:  pokeapi.Config{Limit: 10},
				Cache:   cache,
			},
		},
	}

	for _, c := range cases {
		// act
		api := pokeapi.NewPokeApi(c.actual.baseUrl, c.actual.cache, c.actual.opts...)

		// assert
		if c.expected != api {
			t.Errorf("Invalid PokeApi created for params: %v", c.actual)
			t.Fail()
		}
	}
}

func TestGetLocationAreaFromApi(t *testing.T) {
	// arrange
	data, _ := json.Marshal(pokeapi.LocationAreaResponse{Count: 1})
	cache := NewMockCache()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	api := pokeapi.NewPokeApi(ts.URL, cache, pokeapi.WithLimit(20))

	// act
	r, err := api.GetLocationArea(0)

	// assert
	if err != nil || r.Count != 1 {
		t.Errorf("unexpected: %v, err: %v", r, err)
	}
}

func TestGetLocationAreaFromCache(t *testing.T) {
	// arrange
	data, _ := json.Marshal(pokeapi.LocationAreaResponse{Count: 1})
	cache := NewMockCache()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	api := pokeapi.NewPokeApi(ts.URL, cache, pokeapi.WithLimit(20))

	cacheData, _ := json.Marshal(pokeapi.LocationAreaResponse{Count: 10})
	cache.Add(fmt.Sprintf("%s/location-area?offset=%d&limit=%d", api.BaseUrl, 0, api.Config.Limit), cacheData)

	// act
	r, err := api.GetLocationArea(0)

	// assert
	if err != nil || r.Count != 10 {
		t.Errorf("unexpected: %v, err: %v", r, err)
	}
}

func TestGetLocationAreaDetailsFromApi(t *testing.T) {
	// arrange
	name := "kanto-route-1"
	data, _ := json.Marshal(pokeapi.LocationAreaDetailsResponse{Name: name})
	cache := NewMockCache()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	api := pokeapi.NewPokeApi(ts.URL, cache, pokeapi.WithLimit(20))

	// act
	d, err := api.GetLocationAreaDetails(name)

	// assert
	if err != nil || d.Name != name {
		t.Errorf("unexpected: %v, err: %v", d, err)
	}
}

func TestGetLocationAreaDetailsFromCache(t *testing.T) {
	// arrange
	name := "kanto-route-1"
	data, _ := json.Marshal(pokeapi.LocationAreaDetailsResponse{Name: name})
	cache := NewMockCache()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	api := pokeapi.NewPokeApi(ts.URL, cache, pokeapi.WithLimit(20))

	cacheName := "solaceon-ruins-b3f-e"
	cacheData, _ := json.Marshal(pokeapi.LocationAreaDetailsResponse{Name: cacheName})
	cache.Add(fmt.Sprintf("%s/location-area/%s", api.BaseUrl, cacheName), cacheData)

	// act
	d, err := api.GetLocationAreaDetails(cacheName)

	// assert
	if err != nil || d.Name != cacheName {
		t.Errorf("unexpected: %v, err: %v", d, err)
	}
}

func TestGetPokemonFromApi(t *testing.T) {
	name := "pikachu"
	data, _ := json.Marshal(pokeapi.Pokemon{Name: name})
	cache := NewMockCache()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	api := pokeapi.NewPokeApi(ts.URL, cache)

	p, err := api.GetPokemon(name)
	if err != nil || p.Name != name {
		t.Errorf("unexpected result: %v, err: %v", p, err)
	}
}

func TestGetPokemonFromCache(t *testing.T) {
	name := "pikachu"
	data, _ := json.Marshal(pokeapi.Pokemon{Name: name})
	cache := NewMockCache()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	api := pokeapi.NewPokeApi(ts.URL, cache)

	cacheName := "nosepass"
	cacheData, _ := json.Marshal(pokeapi.Pokemon{Name: cacheName})
	cache.Add(fmt.Sprintf("%s/pokemon/%s", api.BaseUrl, cacheName), cacheData)

	p, err := api.GetPokemon(cacheName)
	if err != nil || p.Name != cacheName {
		t.Errorf("unexpected result: %v, err: %v", p, err)
	}
}

func TestApiHttpStatusError(t *testing.T) {
	cache := NewMockCache()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	api := pokeapi.NewPokeApi(ts.URL, cache)

	t.Run("returns GetPokemon error", func(t *testing.T) {
		_, err := api.GetPokemon("pikachu")
		if err == nil || len(cache.store) > 0 {
			t.Errorf("expected error and no item added to cache: %v, err: %v", len(cache.store), err)
		}
	})
	t.Run("returns GetLocationArea error", func(t *testing.T) {
		_, err := api.GetLocationArea(0)
		if err == nil || len(cache.store) > 0 {
			t.Errorf("expected error and no item added to cache: %v, err: %v", len(cache.store), err)
		}
	})

	t.Run("returns GetLocationAreaDetails error", func(t *testing.T) {
		_, err := api.GetLocationAreaDetails("canalave-city-area")
		if err == nil || len(cache.store) > 0 {
			t.Errorf("expected error and no item added to cache: %v, err: %v", len(cache.store), err)
		}
	})
}

func TestApiParseBodyError(t *testing.T) {
	cache := NewMockCache()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{ invalid json "))
	}))
	defer ts.Close()

	api := pokeapi.NewPokeApi(ts.URL, cache)

	t.Run("returns GetPokemon error", func(t *testing.T) {
		_, err := api.GetPokemon("pikachu")
		if err == nil || len(cache.store) > 0 {
			t.Errorf("expected error and no item added to cache: %v, err: %v", len(cache.store), err)
		}
	})
	t.Run("returns GetLocationArea error", func(t *testing.T) {
		_, err := api.GetLocationArea(0)
		if err == nil || len(cache.store) > 0 {
			t.Errorf("expected error and no item added to cache: %v, err: %v", len(cache.store), err)
		}
	})

	t.Run("returns GetLocationAreaDetails error", func(t *testing.T) {
		_, err := api.GetLocationAreaDetails("canalave-city-area")
		if err == nil || len(cache.store) > 0 {
			t.Errorf("expected error and no item added to cache: %v, err: %v", len(cache.store), err)
		}
	})
}

func TestApiHttpReadBodyError(t *testing.T) {
	cache := NewMockCache()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush() // Send headers
		conn, _, _ := w.(http.Hijacker).Hijack()
		conn.Close() // Close mid-body read
	}))
	defer ts.Close()

	api := pokeapi.NewPokeApi(ts.URL, cache)

	t.Run("returns GetPokemon error", func(t *testing.T) {
		_, err := api.GetPokemon("pikachu")
		if err == nil || len(cache.store) > 0 {
			t.Errorf("expected error and no item added to cache: %v, err: %v", len(cache.store), err)
		}
	})
	t.Run("returns GetLocationArea error", func(t *testing.T) {
		_, err := api.GetLocationArea(0)
		if err == nil || len(cache.store) > 0 {
			t.Errorf("expected error and no item added to cache: %v, err: %v", len(cache.store), err)
		}
	})

	t.Run("returns GetLocationAreaDetails error", func(t *testing.T) {
		_, err := api.GetLocationAreaDetails("canalave-city-area")
		if err == nil || len(cache.store) > 0 {
			t.Errorf("expected error and no item added to cache: %v, err: %v", len(cache.store), err)
		}
	})
}

func TestApiNetworkError(t *testing.T) {
	cache := NewMockCache()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	api := pokeapi.NewPokeApi("http://invalid.localhost/", cache)

	t.Run("returns GetPokemon error", func(t *testing.T) {
		_, err := api.GetPokemon("pikachu")
		if err == nil || len(cache.store) > 0 {
			t.Errorf("expected error and no item added to cache: %v, err: %v", len(cache.store), err)
		}
	})
	t.Run("returns GetLocationArea error", func(t *testing.T) {
		_, err := api.GetLocationArea(0)
		if err == nil || len(cache.store) > 0 {
			t.Errorf("expected error and no item added to cache: %v, err: %v", len(cache.store), err)
		}
	})

	t.Run("returns GetLocationAreaDetails error", func(t *testing.T) {
		_, err := api.GetLocationAreaDetails("canalave-city-area")
		if err == nil || len(cache.store) > 0 {
			t.Errorf("expected error and no item added to cache: %v, err: %v", len(cache.store), err)
		}
	})
}
