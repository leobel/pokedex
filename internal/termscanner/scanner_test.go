package termscanner_test

import (
	"os"
	"testing"
	"time"

	"github.com/leobel/pokedexcli/internal/termscanner"
	"golang.org/x/term"
)

type fakeTerm struct{}

func (fakeTerm) IsTerminal(fd int) bool                  { return true }
func (fakeTerm) MakeRaw(fd int) (*term.State, error)     { return nil, nil }
func (fakeTerm) Restore(fd int, state *term.State) error { return nil }

type InputBuilder struct {
	w        *os.File
	r        *os.File
	oldStdin *os.File
	ts       *termscanner.TermScanner
}

func CreateInputBuilder() InputBuilder {
	r, w, _ := os.Pipe()
	oldStdin := os.Stdin
	os.Stdin = r
	return InputBuilder{
		w:        w,
		r:        r,
		oldStdin: oldStdin,
		ts:       termscanner.New("Pokedex > ", os.Stdin, fakeTerm{}),
	}
}

func (c InputBuilder) Close() {
	os.Stdin = c.oldStdin
	c.w.Close()
}

// helper to feed input into os.Stdin for the duration of a Scan
func (c InputBuilder) withInput(input []byte, fn func(ts *termscanner.TermScanner)) InputBuilder {
	c.w.Write(input)
	fn(c.ts)

	return c
}

func TestTermScannerSimpleText(t *testing.T) {
	builder := CreateInputBuilder()
	defer func() { builder.Close() }()

	// simulate user typing "help<Enter>"
	input := []byte("help\r")
	builder.withInput(input, func(ts *termscanner.TermScanner) {
		ok := ts.Scan()
		if !ok {
			t.Fatal("expected Scan() to return true")
		}

		expected := "help"
		actual := ts.Text()
		if actual != expected {
			t.Errorf("Text() = %s; want %s", actual, expected)
		}
	})

}

func TestTermScannerEmptyLine(t *testing.T) {
	builder := CreateInputBuilder()
	defer func() { builder.Close() }()

	// simulate just Enter—Scan should return false (no command) but not block
	input := []byte("\r")
	builder.withInput(input, func(ts *termscanner.TermScanner) {
		ok := ts.Scan()
		if !ok || ts.Text() != "" {
			t.Fatal("expected Scan() to return true for empty input")
		}
	})
}

func TestTermScannerHistoryUpDown(t *testing.T) {
	builder := CreateInputBuilder()
	defer func() { builder.Close() }()

	// first command
	builder.withInput([]byte("one\r"), func(ts *termscanner.TermScanner) {
		if !ts.Scan() {
			t.Fatal("first Scan() failed")
		}
		if ts.Text() != "one" {
			t.Fatalf("expected first Text() == 'one', actual: %s", ts.Text())
		}
	})

	// second command
	builder.withInput([]byte("two\r"), func(ts *termscanner.TermScanner) {
		if !ts.Scan() {
			t.Fatal("second Scan() failed")
		}
		if ts.Text() != "two" {
			t.Fatalf("expected second Text() == 'two', actual: %s", ts.Text())
		}
	})

	// now press Up (ESC [ A), then Enter
	up := []byte{0x1b, 0x5b, 0x41, '\r'} // arrow up
	builder.withInput(up, func(ts *termscanner.TermScanner) {
		if !ts.Scan() {
			t.Fatal("Scan() after Up failed")
		}
		// history[1] was "two", history[0] was "one"; Up from fresh goes to "two"
		if actual, expected := ts.Text(), "two"; actual != expected {
			t.Errorf("after Up, Text() = %s; expected %s", actual, expected)
		}
	})

	// Down arrow then Enter should go back to empty/new → Scan returns false
	down := []byte{0x1b, 0x5b, 0x42, '\r'}
	builder.withInput(down, func(ts *termscanner.TermScanner) {
		if !ts.Scan() {
			t.Fatal("Scan() after Up failed")
		}

		// history[1] was "two", history[0] was "one"; Down from "two" goes to ""
		if actual, expected := ts.Text(), ""; actual != expected {
			t.Errorf("after Up, Text() = %s; expected %s", actual, expected)
		}
	})
}

func TestTermScannerCtrlC(t *testing.T) {
	builder := CreateInputBuilder()
	defer func() { builder.Close() }()
	// simulate Ctrl+C (0x03)
	input := []byte{0x03}
	done := make(chan struct{})
	go builder.withInput(input, func(ts *termscanner.TermScanner) {
		if ok := ts.Scan(); ok {
			t.Error("expected Scan() to return false on Ctrl+C")
		}
		close(done)
	})

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Scan() did not return after Ctrl+C")
	}
}
