package simpex_test

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/tobiassjosten/go-simpex"
)

func TestCompile(t *testing.T) {
	tcs := map[string]struct {
		pattern []byte
		sx      []byte
		error   bool
	}{
		"escape and handle start/end capture symbols": {
			pattern: []byte("{{{{{Lorem}}} ipsum {{dolor}}}} sit amet."),
			sx:      []byte("\x02{{Lorem}\x03 ipsum {dolor}} sit amet."),
		},

		"separate escape and capture": {
			pattern: []byte("{{{Lorem}} ipsum} dolor sit amet."),
			sx:      []byte("\x02{Lorem} ipsum\x03 dolor sit amet."),
		},

		"handle unopened capture symbols": {
			pattern: []byte("Lorem} ipsum dolor sit amet."),
			error:   true,
		},

		"handle unclosed capture symbols": {
			pattern: []byte("{Lorem ipsum dolor sit amet."),
			error:   true,
		},

		"handle nested capture symbols": {
			pattern: []byte("{Lorem {ipsum} dolor} sit amet."),
			error:   true,
		},

		"escape and handle phrase symbols": {
			pattern: []byte("Lorem * ** ***."),
			sx:      []byte("Lorem \x1d * *\x1d."),
		},

		"escape and handle word symbols": {
			pattern: []byte("Lorem ^ ^^ ^^^."),
			sx:      []byte("Lorem \x1e ^ ^\x1e."),
		},

		"escape and handle captured word symbols": {
			pattern: []byte("{^^}"),
			sx:      []byte("\x02^\x03"),
		},

		"escape and handle character symbols": {
			pattern: []byte("Lorem ip_um d______r s___t amet."),
			sx:      []byte("Lorem ip\x1fum d___r s_\x1ft amet."),
		},

		"escape and handle everything": {
			pattern: []byte("{{{{{}}} {*} {**} {***} {^} {^^} {^^^} {_} {__} {___}"),
			sx:      []byte("\x02{{}\x03 \x02\x1d\x03 \x02*\x03 \x02*\x1d\x03 \x02\x1e\x03 \x02^\x03 \x02^\x1e\x03 \x02\x1f\x03 \x02_\x03 \x02_\x1f\x03"),
		},

		"disallow character word combination": {
			pattern: []byte("_^"),
			error:   true,
		},

		"disallow character phrase combination": {
			pattern: []byte("_*"),
			error:   true,
		},

		"disallow word character combination": {
			pattern: []byte("^_"),
			error:   true,
		},

		"disallow word phrase combination": {
			pattern: []byte("^*"),
			error:   true,
		},

		"disallow phrase character combination": {
			pattern: []byte("*_"),
			error:   true,
		},

		"disallow phrase word combination": {
			pattern: []byte("*^"),
			error:   true,
		},

		"reserved capture start symbol": {
			pattern: []byte("\x02"),
			error:   true,
		},

		"reserved capture end symbol": {
			pattern: []byte("\x03"),
			error:   true,
		},

		"reserved character symbol": {
			pattern: []byte("\x1f"),
			error:   true,
		},

		"reserved word symbol": {
			pattern: []byte("\x1e"),
			error:   true,
		},

		"reserved phrase symbol": {
			pattern: []byte("\x1d"),
			error:   true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			sx, err := simpex.Compile(tc.pattern)

			if tc.error && (err == nil) {
				t.Fatalf("Compile(%q) missing error", tc.pattern)
			} else if !tc.error && (err != nil) {
				t.Fatalf("Compile(%q) unexpected error '%s'", tc.pattern, err)
			}

			if string(tc.sx) != string(sx) {
				t.Fatalf("Compile(%q)\ngot  %q\nwant %q", tc.pattern, sx, tc.sx)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	tcs := map[string]struct {
		pattern []byte
		text    []byte
		matches [][]byte
		error   bool
	}{
		"mismatch longer pattern": {
			pattern: []byte("Lorem ipsum dolor sit amet."),
			text:    []byte("Lorem ipsum."),
		},
		"mismatch longer text": {
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

		"character match empty": {
			pattern: []byte("_"),
			text:    []byte(""),
		},
		"character match single": {
			pattern: []byte("_"),
			text:    []byte("a"),
			matches: [][]byte{},
		},
		"character match capture single": {
			pattern: []byte("{_}"),
			text:    []byte("a"),
			matches: [][]byte{[]byte("a")},
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

		"word match empty": {
			pattern: []byte("^"),
			text:    []byte(""),
		},
		"word match single": {
			pattern: []byte("^"),
			text:    []byte("asdf"),
			matches: [][]byte{},
		},
		"word match capture single": {
			pattern: []byte("{^}"),
			text:    []byte("asdf"),
			matches: [][]byte{[]byte("asdf")},
		},
		"word match prefix": {
			pattern: []byte("^df"),
			text:    []byte("asdf"),
			matches: [][]byte{},
		},
		"word match non-match prefix": {
			pattern: []byte("^df"),
			text:    []byte("asdd"),
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
		"word match mid prefix": {
			pattern: []byte("Lorem ^sum dolor sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"word match capture prefix": {
			pattern: []byte("Lorem {^sum} dolor sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{[]byte("ipsum")},
		},
		"word match suffix": {
			pattern: []byte("Lorem ip^ dolor sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"word match capture suffix": {
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

		"phrase match empty": {
			pattern: []byte("*"),
			text:    []byte(""),
		},
		"phrase match single": {
			pattern: []byte("*"),
			text:    []byte("asdf"),
			matches: [][]byte{},
		},
		"phrase match capture single": {
			pattern: []byte("{*}"),
			text:    []byte("asdf"),
			matches: [][]byte{[]byte("asdf")},
		},
		"phrase match all": {
			pattern: []byte("*"),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"phrase match capture all": {
			pattern: []byte("{*}"),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{[]byte("Lorem ipsum dolor sit amet.")},
		},
		"phrase match beginning": {
			pattern: []byte("* dolor sit amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"phrase match middle": {
			pattern: []byte("Lorem * amet."),
			text:    []byte("Lorem ipsum dolor sit amet."),
			matches: [][]byte{},
		},
		"phrase match end": {
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
		"phrase non-matching following": {
			pattern: []byte("* amet"),
			text:    []byte("asdf"),
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

		"unclosed phrase capture": {
			pattern: []byte("{*"),
			text:    []byte("0"),
			error:   true,
		},
		"unopened phrase capture": {
			pattern: []byte("*}"),
			text:    []byte("0"),
			error:   true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			matches, err := simpex.Match(tc.pattern, tc.text)

			if tc.error && (err == nil) {
				t.Fatalf(
					"Match(%q, %q) missing error",
					tc.pattern, tc.text,
				)
			} else if !tc.error && (err != nil) {
				t.Fatalf(
					"Match(%q, %q) unexpected error '%s'",
					tc.pattern, tc.text, err,
				)
			}

			if tc.matches != nil && matches == nil {
				t.Fatalf(
					"Match(%q, %q) = nil, want %q",
					tc.pattern, tc.text, tc.matches,
				)
			} else if tc.matches == nil && matches != nil {
				t.Fatalf(
					"Match(%q, %q) = %q, want nil",
					tc.pattern, tc.text, matches,
				)
			} else if !reflect.DeepEqual(tc.matches, matches) {
				t.Fatalf(
					"Match(%q, %q) = %q, want %q",
					tc.pattern, tc.text, matches, tc.matches,
				)
			}
		})
	}
}

func FuzzMatch(f *testing.F) {
	f.Add(
		[]byte("{Lorem} {^} do{_}or {*}."),
		[]byte("Lorem ipsum dolor sit amet."),
	)

	f.Fuzz(func(t *testing.T, pattern, text []byte) {
		_, _ = simpex.Match(pattern, text)
	})
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
			sx, _ := simpex.Compile(benchmark[1])
			for i := 0; i < b.N; i++ {
				r1 = sx.Match(benchmark[0])
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
