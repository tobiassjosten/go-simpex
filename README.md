# simpex

A simpler and faster alternative to regular expressions. Sprung from the [Nogfx MUD client](https://github.com/tobiassjosten/nogfx), this Go library can help you match and extract subsets from text.

Table of contents:

- [Installation](#installation)
- [Quick start](#quick-start)
- [Usage](#usage)
- [Limitations](#limitations)
- [Contribute](#contribute)

## Installation

1. Download the module:

    go get -u github.com/gin-gonic/gin

2. Import it in your project:

    import "github.com/tobiassjosten/go-simpex"

## Quick start

```
package main

import (
  "fmt"
  "github.com/tobiassjosten/go-simpex"
)

func main() {
  matches := simpex.Match("Hello {^}!", "Hello world!")
  if matches != nil {
    fmt.Printf("Howdy %s!\n", matches[0])
  }
}
```

## Usage

Simpex matches patterns against the full texts given, never partially. A pattern of `two` wouldn't match the text `one two three`. In regexp speak, patterns are anchored at both ends and simpex `two` would be the equivalent of regepx `^two$`.

Simpex can match single characters, words, and phrases using the symbols `_`, `^`, and `*` respectively. In order to match those symbols, they can be escaped by doubling them, like `__`, `^^`, and `**`.

- A character is represented by a byte, not a rune. Non-ASCII characters might therefor need

- A word is represented by alphanumeric characters (`[a-zA-Z0-9]+` in regexp).

- A phrase is represented by anything that would fulfill the other parts of the patter – greedily or otherwise.

Simpex can also capture substrings, using the `{` and `}` symbols. Again, escaping them is simply a matter of repeating, like `{{` and `}}`.

There's one main function, `Match()`, which returns a string slice of captures. A `nil` return value signified a non-match.

The following examples might make it easier to understand.

```
package main

import (
  "fmt"
  "github.com/tobiassjosten/go-simpex"
)

func main() {
  var matches [][]byte

  // Match a character.
  simpex.Match("Hello w_rld!", "Hello world!") != nil

  // Match an underscore.
  simpex.Match("snake__case", "snake_case") != nil

  // Match a word.
  simpex.Match("Hello ^!", "Hello world!") != nil

  // Match a caret.
  simpex.Match("Look up! ^^", "Look up! ^") != nil

  // Match a phrase.
  simpex.Match("*!", "Hello world!") != nil

  // Match a star.
  simpex.Match("It's a star! **", "It's a star! *") != nil

  // Capture substrings and print: "Howdy world! I wonder, how are you?"
  matches := simpex.Match("Hello {^}, {*}{_}", "Hello world, how are you?")
  if matches != nil {
    fmt.Printf("Howdy %s! I wonder, %s%s\n", matches[0], matches[1], matches[2])
  }
}
```

## Limitations

- The module deals with bytes and byte slices, meaning it doesn't support wide runes or other non-ASCII characters for its `_` symbol.

- The matching algorithm can probably be improved a whole lot. It's developed for use with short texts meant for human reading, so anything outside of that could potentially reveal flaws I haven't bumped into.

- I'm sure there are many other limitations to this. I originally built it for my own needs, it works perfectly for that, and I haven't given too much thought to anything outside of my narrow use case.

## Contribute

Feel free to [create a ticket](https://github.com/tobiassjosten/go-simpex/issues/new) if you want to discuss or suggest something. I'd love your input and will happily work with you to cover your use cases and explore your ideas for improvements.

Changes can be suggested directly by [creating a pull request](https://github.com/tobiassjosten/go-simpex/compare) but I'd recommend starting an issue first, so you don't end up wasting your time with something I end up rejecting.

There's an extensive test suite, along with a benchmark, which you can use to make sure that your change works and is performant. You can run them as you would any other Go test/benchmark:

    go test ./...

    go test ./.. -bench=.

### Contributors

- [Tobias Sjösten](https://github.com/tobiassjosten)
