package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	// arrange
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
		actual := cleanInput(c.input)

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
