## Welcome to Pokedex!

Pokedex is a REPL that uses the [PokeApi](https://pokeapi.co/docs/v2) to fetch data about Pokemon. 
This project is the result of completing Boot.dev course: [Build a Pokedex in Go](https://www.boot.dev/courses/build-pokedex-cli-golang)

```cli
go run .
Pokedex > help
Welcome to the Pokedex!
Usage:

catch: Trying to catch a Pokemon by name
exit: Exit the Pokedex
explore: List of all the Pokemons located in a specific area
help: Displays this help message
inspect: Show name, height, weight, stats and type(s) of Pokemon
map: Display next 20 location areas of the Pokemon world
mapb: Display previous 20 location areas of the Pokemon world
pokedex: Show all Pokemon you've caught so far
Up/Down keys: Use it to navigate between previous and next commands
Pokedex > 
```

### Testing
```cli
go test ./...
```
