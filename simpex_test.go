package simpex_test

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/tobiassjosten/go-simpex"
)

func TestMatch(t *testing.T) {
	tcs := map[string]struct {
		pattern []byte
		text    []byte
		matches [][]byte
	}{
		"mismatch long pattern": {
			pattern: []byte("Lorem ipsum dolor sit amet."),
			text:    []byte("Lorem ipsum."),
		},
		"mismatch long text": {
			pattern: []byte("Lorem ipsum."),
			text:    []byte("Lorem ipsum dolor sit amet."),
		},

		"exact match simple": {
			pattern: []byte("Lorem ipsum dolor sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"exact match capture": {
			pattern: []byte("{Lorem} ipsum dolor sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{[]byte("Lorem")},
		},
		"exact match escaped capture simple": {
			pattern: []byte("{{Lorem}} ipsum dolor sit amet."),
			text:    []byte("{Lorem} ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"exact match escaped capture capture one": {
			pattern: []byte("{{{Lorem}}} ipsum dolor sit amet."),
			text:    []byte("{Lorem} ipsum dolor sit amet."),
			matches: [][]byte{[]byte("{Lorem}")},
		},
		"exact match escaped capture capture two": {
			pattern: []byte("{{{Lorem}} ipsum} dolor sit amet."),
			text:    []byte("{Lorem} ipsum dolor sit amet."),
			matches: [][]byte{[]byte("{Lorem} ipsum")},
		},

		"character match simple": {
			pattern: []byte("Lorem ipsum do_or sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"character match capture": {
			pattern: []byte("Lorem ipsum do{_}or sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{{'l'}},
		},
		"character match escaped one": {
			pattern: []byte("Lorem ipsum do__or sit amet."),
			text:    []byte("Lorem ipsum do_or sit amet."),
			matches: [][]byte{},
		},
		"character match escaped two": {
			pattern: []byte("Lorem ipsum do___or sit amet."),
			text:    []byte("Lorem ipsum do_lor sit amet."),
			matches: [][]byte{},
		},

		"word match simple": {
			pattern: []byte("Lorem ^ dolor sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"word match capture": {
			pattern: []byte("Lorem {^} dolor sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{[]byte("ipsum")},
		},
		"word match prefix": {
			pattern: []byte("Lorem ^sum dolor sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"word match prefix capture": {
			pattern: []byte("Lorem {^sum} dolor sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{[]byte("ipsum")},
		},
		"word match suffix": {
			pattern: []byte("Lorem ip^ dolor sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"word match suffix capture": {
			pattern: []byte("Lorem {ip^} dolor sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{[]byte("ipsum")},
		},
		"word match escaped one": {
			pattern: []byte("Lorem ^^ dolor sit amet."),
			text:    []byte("Lorem ^ dolor sit amet."),
			matches: [][]byte{},
		},
		"word match escaped two": {
			pattern: []byte("Lorem ^^^ dolor sit amet."),
			text:    []byte("Lorem ^ipsum dolor sit amet."),
			matches: [][]byte{},
		},

		"phrase match simple one": {
			pattern: []byte("Lorem ipsum dolor *."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"phrase match simple two": {
			pattern: []byte("Lorem ipsum dolor * lol."),
			text:    []byte("Lorem ipsum dolor sit amet lol."),
			matches: [][]byte{},
		},
		"phrase match capture one": {
			pattern: []byte("Lorem ipsum dolor {*}."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{[]byte("sit amet")},
		},
		"phrase match capture two": {
			pattern: []byte("Lorem ipsum dolor {*} lol."),
			text:    []byte("Lorem ipsum dolor sit amet lol."),
			matches: [][]byte{[]byte("sit amet")},
		},
		"phrase match prefix one": {
			pattern: []byte("Lorem ipsum dolor *et."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"phrase match prefix two": {
			pattern: []byte("Lorem ipsum dolor *et lol."),
			text:    []byte("Lorem ipsum dolor sit amet lol."),
			matches: [][]byte{},
		},
		"phrase match capture prefix one": {
			pattern: []byte("Lorem ipsum dolor {*et}."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{[]byte("sit amet")},
		},
		"phrase match capture prefix two": {
			pattern: []byte("Lorem ipsum dolor {*et} lol."),
			text:    []byte("Lorem ipsum dolor sit amet lol."),
			matches: [][]byte{[]byte("sit amet")},
		},
		"phrase match suffix one": {
			pattern: []byte("Lorem ipsum dolor si*."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"phrase match suffix two": {
			pattern: []byte("Lorem ipsum dolor si* lol."),
			text:    []byte("Lorem ipsum dolor sit amet lol."),
			matches: [][]byte{},
		},
		"phrase match capture suffix one": {
			pattern: []byte("Lorem ipsum dolor {si*}."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{[]byte("sit amet")},
		},
		"phrase match capture suffix two": {
			pattern: []byte("Lorem ipsum dolor {si*} lol."),
			text:    []byte("Lorem ipsum dolor sit amet lol."),
			matches: [][]byte{[]byte("sit amet")},
		},
		"phrase match escaped one": {
			pattern: []byte("Lorem ipsum dolor **."),
			text:    []byte("Lorem ipsum dolor *."),
			matches: [][]byte{},
		},
		"phrase match escaped two": {
			pattern: []byte("Lorem ipsum dolor ***."),
			text:    []byte("Lorem ipsum dolor *sit amet."),
			matches: [][]byte{},
		},

		"combination match simple": {
			pattern: []byte("Lorem ^ do_or *."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"combination match escaped": {
			pattern: []byte("Lorem ^^ do__or **."),
			text:    []byte("Lorem ^ do_or *."),
			matches: [][]byte{},
		},
		"combination match capture": {
			pattern: []byte("{Lorem} {^} do{_}or {*}."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{
				[]byte("Lorem"),
				[]byte("ipsum"),
				{'l'},
				[]byte("sit amet"),
			},
		},
		"combination match capture escaped": {
			pattern: []byte("{{{Lorem}}} {^^} do{__}or {**}."),
			text:    []byte("{Lorem} ^ do_or *."),
			matches: [][]byte{
				[]byte("{Lorem}"),
				[]byte("^"),
				{'_'},
				[]byte("*"),
			},
		},

		"learn fitness from maric begin": {
			pattern: []byte("* bows to you and commences the lesson in ^."),
			text:    []byte("Maric, a filthy ratman bows to you and commences the lesson in Fitness."),
			matches: [][]byte{},
		},
		"learn fitness from maric continue": {
			pattern: []byte("* continues your training in ^."),
			text:    []byte("Maric, a filthy ratman continues your training in Fitness."),
			matches: [][]byte{},
		},
		"learn fitness from maric finish": {
			pattern: []byte("* bows to you - the lesson in ^ is over."),
			text:    []byte("Maric, a filthy ratman bows to you - the lesson in Fitness is over."),
			matches: [][]byte{},
		},
		"learn x y from z input capture": {
			pattern: []byte("learn {^} {^ from *}"),
			text:    []byte("learn 15 fitness from maric"),
			matches: [][]byte{[]byte("15"), []byte("fitness from maric")},
		},
		"learn x y from z input incomplete": {
			pattern: []byte("learn {^} {^ from *}"),
			text:    []byte("learn"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			matches := simpex.Match(tc.pattern, tc.text)
			if !reflect.DeepEqual(tc.matches, matches) {
				t.Fatalf(
					"simpex.Match(%q, %q) = %q, want %q",
					tc.pattern, tc.text, matches, tc.matches,
				)
			}
		})
	}
}

var (
	benchresult1 [][]byte
	benchresult2 [][][]byte
	benchmarks   = map[string][][]byte{
		"exact match": {
			[]byte("Lorem ipsum dolor sit amet."),
			[]byte("Lorem ipsum dolor sit amet."),
			[]byte("Lorem ipsum dolor sit amet."),
		},
		"character match": {
			[]byte("Lorem ipsum dolor sit amet."),
			[]byte("Lorem ipsum do_or sit amet."),
			[]byte("Lorem ipsum do.or sit amet."),
		},
		"word match": {
			[]byte("Lorem ipsum dolor sit amet."),
			[]byte("Lorem ^ dolor sit amet."),
			[]byte("Lorem [a-zA-Z0-9]+ dolor sit amet."),
		},
		"phrase match": {
			[]byte("Lorem ipsum dolor sit amet."),
			[]byte("Lorem ipsum dolor * amet."),
			[]byte("Lorem ipsum dolor .+ amet."),
		},
		"all specials": {
			[]byte("Lorem ipsum dolor sit amet."),
			[]byte("{Lorem} {^} do{_}or {*}."),
			[]byte("(Lorem) ([a-zA-Z0-9]+) do(.)or (.+)."),
		},
	}
)

func BenchmarkMatch(b *testing.B) {
	var r1 [][]byte
	var r2 [][][]byte

	for name, benchmark := range benchmarks {
		b.Run(fmt.Sprintf("%s simpex", name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r1 = simpex.Match(benchmark[1], benchmark[0])
			}
		})

		b.Run(fmt.Sprintf("%s regexp", name), func(b *testing.B) {
			pattern, _ := regexp.Compile(string(benchmark[2]))
			for i := 0; i < b.N; i++ {
				r2 = pattern.FindAllSubmatch(benchmark[0], -1)
			}
		})
	}

	benchresult1 = r1
	benchresult2 = r2
}
