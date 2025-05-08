package termscanner

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"slices"

	"golang.org/x/term"
)

// Term is the subset of x/term you need.
type Term interface {
	IsTerminal(fd int) bool
	MakeRaw(fd int) (*term.State, error)
	Restore(fd int, state *term.State) error
}

// RealTerm implements Term using x/term.
type RealTerm struct{}

func (RealTerm) IsTerminal(fd int) bool                  { return term.IsTerminal(fd) }
func (RealTerm) MakeRaw(fd int) (*term.State, error)     { return term.MakeRaw(fd) }
func (RealTerm) Restore(fd int, state *term.State) error { return term.Restore(fd, state) }

type PokedexScanner interface {
	Scan() bool
	Text() string
	Err() error
}

type TermScanner struct {
	fd      uintptr
	prompt  string
	cmd     string
	history []string
	index   int
	term    Term
}

func New(prompt string, f *os.File, term Term) *TermScanner {
	ts := &TermScanner{
		fd:     f.Fd(),
		prompt: prompt,
		index:  -1,
		term:   term,
	}
	return ts
}

func (ts *TermScanner) Scan() bool {
	result := func() bool {
		fmt.Print(ts.prompt)
		var buf = bytes.Buffer{}
		oldState, err := ts.term.MakeRaw(int(ts.fd))
		if err != nil {
			panic(err)
		}
		defer ts.term.Restore(int(ts.fd), oldState)
		for {
			b := make([]byte, 3)
			n, err := os.Stdin.Read(b)
			if err != nil {
				panic(err)
			}
			if b[0] == 0x03 { // Ctrl+C
				buf.Reset()
				fmt.Print("\r\n")
				return false
			}
			// get first 3 bytes
			seq := b[:n]

			switch {
			case bytes.Equal(seq, []byte{0x1b, 0x5b, 0x41}): // ESC [ A
				// Up arrow
				if len(ts.history) == 0 {
					continue
				}
				if ts.index > 0 {
					ts.index--
					buf.Reset()
					buf.WriteString(ts.history[ts.index])
					ts.redrawLine(&buf)
				}

			case bytes.Equal(seq, []byte{0x1b, 0x5b, 0x42}): // ESC [ B
				// Down arrow
				if len(ts.history) == 0 {
					continue
				}
				if ts.index < len(ts.history)-1 {
					ts.index++
					buf.Reset()
					buf.WriteString(ts.history[ts.index])
					ts.redrawLine(&buf)
				}

			case slices.Index(seq, '\r') >= 0 || slices.Index(seq, '\n') >= 0:
				index := int(math.Max(float64(slices.Index(seq, '\r')), float64(slices.Index(seq, '\n'))))
				buf.Write(seq[0:index])
				// Enter pressed
				command := buf.String()
				ts.cmd = command
				if command != "" {
					ts.cmd = command
					if size := len(ts.history); size == 0 || ts.history[size-1] != command {
						ts.history = append(ts.history, command)
					}
					ts.index = len(ts.history)
				}
				buf.Reset()
				fmt.Print("\r\n")
				return true

			case seq[0] == 127:
				// Backspace
				if buf.Len() > 0 {
					buf.Truncate(buf.Len() - 1)
					ts.redrawLine(&buf)
				}

			default:
				// print characters type
				buf.Write(seq)
				fmt.Printf("%s", seq)
			}
		}
	}()

	return result
}

func (ts *TermScanner) Text() string {
	return ts.cmd
}

func (ts *TermScanner) Err() error {
	return nil
}

func (ts *TermScanner) redrawLine(buf *bytes.Buffer) {
	// \r = return to line start, \x1b[2K = clear entire line
	fmt.Printf("\r\x1b[2K%s", ts.prompt)
	fmt.Print(buf.String())
}
