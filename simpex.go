// Package simpex is a simpler and faster alternative to regexp.
//
// Usage is very straightforward. You first compile your pattern into a Simpex,
// which is then used to match against a given text. As a convenience, the
// global Match() function handles both compilation and matching.
package simpex

import (
	"bytes"
	"fmt"
)

const (
	// These special symbols makes compilation and pattern matching a lot
	// easier and faster later on.
	captureStart byte = 2
	captureEnd   byte = 3
	phraseMatch  byte = 29
	wordMatch    byte = 30
	charMatch    byte = 31
)

var (
	matchchars = map[byte]byte{
		'{': captureStart,
		'}': captureEnd,
		'_': charMatch,
		'^': wordMatch,
		'*': phraseMatch,
	}
)

// Match a text against a pattern to see if it matches. This is a convenience
// wrapper for Compile() and Simpex.Match(). If it matches, captures matches
// are returned. It it doesn't, nil is returned.
func Match(pattern []byte, text []byte) ([][]byte, error) {
	sx, err := Compile(pattern)
	if err != nil {
		return nil, err
	}

	return sx.Match(text), nil
}

// Simpex represents a compiled simple expression. It is assumed to be a valid
// pattern, so any construction outside of Compile() is done at one's own risk.
type Simpex []byte

// Compile validates and converts a given pattern into something optimized for
// matching.
func Compile(pattern []byte) (Simpex, error) {
	capturing := false

	// Avoid mutating pattern slice.
	compiled := make([]byte, len(pattern))
	copy(compiled, pattern)

	uncombinable := false

	for i := 0; i < len(compiled); i++ {
		char := compiled[i]

		switch char {
		case captureStart, captureEnd, charMatch, wordMatch, phraseMatch:
			return nil, fmt.Errorf(
				"reserved character '%x' at position %d",
				char, i,
			)

		// These two are only here for all non-symbolic characters to
		// fall under the default case. Their logic follows after the
		// switch (except for the non-capture, uncombinable stuff).
		case '{', '}':
		case '_', '^', '*':
			if uncombinable {
				return nil, fmt.Errorf("invalid combination at position %d", i)
			}
			uncombinable = true

		default:
			uncombinable = false
			continue
		}

		// Determine how many of the same are repeated.
		repeat := bytes.IndexFunc(compiled[i:], isnot(char))

		// Make sure capture symbols are lined up.
		if repeat%2 != 0 && char == '{' {
			if capturing {
				return nil, fmt.Errorf("unclosed capture at position %d", i)
			}
			capturing = true
		} else if repeat%2 != 0 && char == '}' {
			if !capturing {
				return nil, fmt.Errorf("unopened capture at position %d", i)
			}
			capturing = false
		}

		// Consolidate escaped characters.
		if repeat > 1 || repeat < 0 {
			if repeat < 0 {
				repeat = len(compiled) - i
			}

			sequence := bytes.Repeat([]byte{char}, repeat/2)

			// For '{' we want the matching symbol before.
			if repeat%2 != 0 && char == '{' {
				sequence = append([]byte{matchchars[char]}, sequence...)
			} else if repeat%2 != 0 {
				sequence = append(sequence, matchchars[char])
			}

			compiled = append(
				append(compiled[:i], sequence...),
				compiled[i+repeat:]...,
			)

			i += repeat/2 + repeat%2 - 1

			continue
		}

		// Replace the symbol with a matching character.
		compiled[i] = matchchars[char]
	}

	if capturing {
		return nil, fmt.Errorf("unclosed capture at position %d", len(compiled)-1)
	}

	return Simpex(compiled), nil
}

// Match a text against a pattern to see if it matches. If it does, captured
// matches are returned. If it doesn't, nil is returned.
func (sx Simpex) Match(text []byte) [][]byte {
	captures := [][]byte{}

	var capture []byte

	for len(sx) > 0 {
		char := sx[0]

		switch char {
		case captureStart:
			capture = []byte{}
			sx = sx[1:]

		case captureEnd:
			captures = append(captures, capture)
			capture = nil
			sx = sx[1:]

		case charMatch:
			if len(text) == 0 {
				return nil
			}

			if capture != nil {
				capture = append(capture, text[0])
			}

			sx = sx[1:]
			text = text[1:]

		case wordMatch:
			if len(text) == 0 || isnotalphanum(rune(text[0])) {
				return nil
			}

			// Default to matching the whole word.
			edge := bytes.IndexFunc(text, isnotalphanum)
			if edge < 1 {
				edge = len(text)
			}

			// The end of the word is matched by static alphanums.
			if len(sx) > 1 && isalphanum(rune(sx[1])) {
				start := 1
				end := bytes.IndexFunc(sx[start:], isnotalphanum) + start
				if end-start < 0 {
					end = len(sx)
				}

				edge = bytes.Index(text, sx[start:end])
				if edge < 0 {
					return nil
				}
			}

			if capture != nil {
				capture = append(capture, text[:edge]...)
			}

			sx = sx[1:]
			text = text[edge:]

		case phraseMatch:
			if len(text) == 0 {
				return nil
			}

			// Default to a very greedy match.
			edge := len(text)

			// Find the beginning of the next following non-symbol
			// subtext, from where we'll match this phrase.
			start := bytes.IndexFunc(sx, isnotsymbol)

			// With no following non-symbols we make this phrase a
			// greedy one, matching as much as possible.
			if start >= 0 {
				end := bytes.IndexFunc(sx[start:], issymbol) + start
				if end-start < 0 {
					end = len(sx)
				}

				edge = bytes.Index(text, sx[start:end])
				if edge < 0 {
					return nil
				}
			}

			if capture != nil {
				capture = append(capture, text[:edge]...)
			}

			sx = sx[1:]
			text = text[edge:]

		default:
			// Either there's no more text to match or the text
			// doesn't match, so we fail the operation.
			if len(text) == 0 || char != text[0] {
				return nil
			}

			if capture != nil {
				capture = append(capture, text[0])
			}

			sx = sx[1:]
			text = text[1:]
		}
	}

	if len(sx) > 0 || len(text) > 0 {
		return nil
	}

	return captures
}

func isalphanum(r rune) bool {
	return (r >= '0' && r <= '9') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z')
}

func isnotalphanum(r rune) bool {
	return !isalphanum(r)
}

func issymbol(r rune) bool {
	return r == rune(captureStart) ||
		r == rune(captureEnd) ||
		r == rune(charMatch) ||
		r == rune(wordMatch) ||
		r == rune(phraseMatch)
}

func isnotsymbol(r rune) bool {
	return !issymbol(r)
}

func isnot(b byte) func(r rune) bool {
	r := rune(b)
	return func(rr rune) bool {
		return r != rr
	}
}
