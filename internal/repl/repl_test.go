package repl_test

import (
	"testing"

	"github.com/leobel/pokedexcli/internal/repl"
	"github.com/leobel/pokedexcli/internal/termscanner"
)

type MockScanner struct {
	termscanner.PokedexScanner
}

func NewMockScanner() *MockScanner {
	return &MockScanner{}
}

func (c *MockScanner) Scan() bool {
	return true
}

func (c *MockScanner) Text() string {
	return "mock scanner"
}

func (c *MockScanner) Err() error {
	return nil
}

func TestCleanInput(t *testing.T) {
	// arrange
	replCli := repl.NewRepl(NewMockScanner())
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "  hello,    world  ",
			expected: []string{"hello,", "world"},
		},
		{
			input:    "  HeLLo World!",
			expected: []string{"hello", "world!"},
		},
		{
			input:    "    ",
			expected: []string{},
		},
	}

	// act
	for _, c := range cases {
		actual := replCli.CleanInput(c.input)

		if len(c.expected) != len(actual) {
			t.Errorf("Invalid clean input size, expected: %v, but got: %v", len(c.expected), len(actual))
		}

		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if expectedWord != word {
				t.Errorf("Invalid clean input for: %s expected: %s but got: %s", c.input, expectedWord, word)
				t.Fail()
			}
		}
	}
}
