package simpex

import (
	"bytes"
)

// Match a text against a pattern to see if it matches. If it does, captured
// matches are returned. If it doesn't, nil is returned.
func Match(pattern []byte, text []byte) [][]byte {
	captures := [][]byte{}

	var capture []byte

	tick := func() {
		if capture != nil {
			capture = append(capture, text[0])
		}

		pattern = pattern[1:]
		text = text[1:]
	}

	for len(pattern) > 0 && len(text) > 0 {
		p := pattern[0]

		switch p {
		case '{':
			if bytes.IndexFunc(pattern, isnot(p))%2 == 0 {
				if p != text[0] {
					return nil
				}

				pattern = pattern[1:]
				tick()
				continue
			}

			capture = []byte{}
			pattern = pattern[1:]

		case '}':
			if len(pattern) > 1 && p == pattern[1] {
				if p != text[0] {
					return nil
				}

				pattern = pattern[1:]
				tick()
				continue
			}

			captures = append(captures, capture)
			capture = nil
			pattern = pattern[1:]

		case '^':
			if len(pattern) > 1 && p == pattern[1] {
				if p != text[0] {
					return nil
				}

				pattern = pattern[1:]
				tick()
				continue
			}

			edge := bytes.IndexFunc(text, isnotalphanum)
			if edge < 1 {
				return nil
			}

			edge -= bytes.IndexFunc(pattern[1:], isnotalphanum)

			if capture != nil {
				capture = append(capture, text[:edge]...)
			}

			pattern = pattern[1:]
			text = text[edge:]

		case '*':
			if len(pattern) > 1 && p == pattern[1] {
				if p != text[0] {
					return nil
				}

				pattern = pattern[1:]
				tick()
				continue
			}

			if len(pattern) == 1 {
				if capture != nil {
					return nil
				}

				return captures
			}

			if capture != nil && len(pattern) == 2 && pattern[1] == '}' {
				capture = append(capture, text...)
				captures = append(captures, capture)

				return captures
			}

			start := bytes.IndexFunc(pattern, isnotspecial)

			end := bytes.IndexFunc(pattern[start:], isspecial) - 1
			if end < 0 {
				end = len(pattern[start:]) - 1
			}

			segment := pattern[start : start+end+1]

			edge := bytes.Index(text, segment)
			if edge < 0 {
				return nil
			}

			if capture != nil {
				capture = append(capture, text[:edge]...)
			}

			pattern = pattern[1:]
			text = text[edge:]

		case '_':
			if len(pattern) > 1 && p == pattern[1] {
				if p != text[0] {
					return nil
				}

				pattern = pattern[1:]
				tick()
				continue
			}

			pattern[0] = text[0]
			tick()

		default:
			if p != text[0] {
				return nil
			}

			tick()
		}
	}

	if len(pattern) > 0 || len(text) > 0 {
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

func isspecial(r rune) bool {
	return r == '{' || r == '}' || r == '_' || r == '^' || r == '*'
}

func isnotspecial(r rune) bool {
	return !isspecial(r)
}

func isnot(b byte) func(r rune) bool {
	r := rune(b)
	return func(rr rune) bool {
		return r != rr
	}
}
