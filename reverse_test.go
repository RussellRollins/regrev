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
			Name: "Okay, let's get group-y",
			Reg:  regexp.MustCompile("[A-D|3-8]b"),
		},
		//		{
		//			Name: "Make me an ip address!",
		//			Reg:  regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}`),
		//		},
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
