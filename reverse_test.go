package regrev_test

import (
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/russellrollins/regrev"
)

func TestRegexReverse(t *testing.T) {
	rand.Seed(time.Now().UnixNano() / int64(time.Millisecond))

	rr, err := regrev.NewRegexReverser()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name string
		Reg  *regexp.Regexp
	}{
		{
			Name: "A Single Character Regexp",
			Reg:  regexp.MustCompile(`a`),
		},
		{
			Name: "Uses the + Modifier for Regexp",
			Reg:  regexp.MustCompile(`a+b+`),
		},
		{
			Name: "Every type of Modifier",
			Reg:  regexp.MustCompile(`abc{2,5}d?e*f+`),
		},
		{
			Name: "Do some pretty funky stuff with {}'s",
			Reg:  regexp.MustCompile(`{356{3}{{{{4}}{5}`),
		},
		{
			Name: "Dead simple set",
			Reg:  regexp.MustCompile(`[123ABC]`),
		},
		{
			Name: "Dead simple set's evil twim",
			Reg:  regexp.MustCompile(`[^123ABC]`),
		},
		{
			Name: "Get that set range-y",
			Reg:  regexp.MustCompile(`[A-D][EFG]b`),
		},
		{
			Name: "will recursive group solving just work?",
			Reg:  regexp.MustCompile(`(a+b+)?(abc{2,5}){2,4}`),
		},
		{
			Name: "Escaping pain",
			Reg:  regexp.MustCompile(`[*\-\\]`),
		},
		{
			Name: "Escaping pain 2",
			Reg:  regexp.MustCompile(`[*-\\]`),
		},
		{
			Name: "Groups and modifiers, together at last",
			Reg:  regexp.MustCompile("[ABC]?[DEF]{4,5}[GHI]+"),
		},
		{
			Name: "The world is your oyster, special characters start",
			Reg:  regexp.MustCompile(`a.+b`),
		},
		{
			Name: "Make me an ip address! Special characters with a purpose",
			Reg:  regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}`),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			got, err := rr.Reverse(tc.Reg)
			if err != nil {
				t.Fatal(err)
			}

			if !tc.Reg.MatchString(got) {
				t.Errorf("expected reversed string `%s` to match regexp %s", got, tc.Reg.String())
			}
		})
	}
}

// TODO: FUZZ TESTERRRRRRR
// This seems like the kind of project that would really benefit from this. Generate many many valid regexps
// Throw them in, see if they produce a matching string.
