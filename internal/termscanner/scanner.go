package termscanner

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/term"
)

type TermScanner struct {
	fd      uintptr
	prompt  string
	cmd     string
	history []string
	index   int
}

func New(prompt string, f *os.File) *TermScanner {
	ts := &TermScanner{
		fd:     f.Fd(),
		prompt: prompt,
		index:  -1,
	}
	return ts
}

func (ts *TermScanner) Scan() bool {
	result := func() bool {
		fmt.Print(ts.prompt)
		var buf = bytes.Buffer{}
		oldState, err := term.MakeRaw(int(ts.fd))
		if err != nil {
			panic(err)
		}
		defer term.Restore(int(ts.fd), oldState)
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

			case seq[0] == '\r' || seq[0] == '\n':
				// Enter pressed
				command := buf.String()
				if command != "" {
					ts.cmd = command
					ts.history = append(ts.history, command)
					ts.index = len(ts.history)
					buf.Reset()
					fmt.Print("\r\n")
					return true
				} else {
					// new prompt line
					fmt.Print("\r\n")
					ts.redrawLine(&buf)
				}

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
