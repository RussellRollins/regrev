package regrev

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type RegexReverser struct {
	maxRepeats int
}

type component interface {
	solve() (string, error)
}

// We will solve the regex by recursively splitting and solving components.
// The degenerative case is a set of solvable cases, such as literals and
// ranges. Other components cannot be solved directly, but contain componds
// themselves.

type compound struct {
	rr       *RegexReverser
	compound string
}

type literal struct {
	rr       *RegexReverser
	literal  string
	modifier string
}

type special struct {
	rr       *RegexReverser
	special  string
	modifier string
}

type regRange struct {
	rr       *RegexReverser
	regRange string
	modifier string
}

type group struct {
	compound *compound
	modifier string
}

func NewRegexReverser(options ...func(*RegexReverser) error) (*RegexReverser, error) {
	r := &RegexReverser{
		maxRepeats: 64,
	}

	for _, option := range options {
		err := option(r)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (rr *RegexReverser) Reverse(reg *regexp.Regexp) (string, error) {
	// Create a group that is the whole regex, with no modifier
	comp := &compound{
		rr:       rr,
		compound: reg.String(),
	}

	// Recursively solve the group by splitting it into component groups, solving them.
	return comp.solve()
}

// There are four parts of a compound we must split:
//   1. literals "a" "\(" etc.
//   2. specials "." "\S" etc.
//   3. ranges "[abc]" "[1-9]" etc
//   4. groups, which are themselves compounds, but that we apply modifiers to.
func (c *compound) split() []component {
	splits := []component{}
	escaped := false
	for i := 0; i < len(c.compound); i++ {
		char := c.compound[i]

		// If we're not escaped, and get the escape character, enter escape and contine
		if !escaped && char == '\\' {
			escaped = true
			continue
		}

		if escaped {
			if reserved(char) {
				// If we're escaped, and get a reserved character, that's just a literal
				modifier, skip := extractModifier(c.compound, i+1)
				i += skip
				splits = append(splits, &literal{
					rr:       c.rr,
					literal:  fmt.Sprintf("\\%s", string(char)),
					modifier: modifier,
				})
			} else {
				// If we're escaped, and get a non-reserved character, that's a special
				modifier, skip := extractModifier(c.compound, i+1)
				i += skip
				splits = append(splits, &special{
					rr:       c.rr,
					special:  fmt.Sprintf("\\%s", string(char)),
					modifier: modifier,
				})
			}
			escaped = false
			continue
		}

		// if this is not a reserved character, it's just a literal
		if !reserved(char) {
			modifier, skip := extractModifier(c.compound, i+1)
			i += skip
			splits = append(splits, &literal{
				rr:       c.rr,
				literal:  string(char),
				modifier: modifier,
			})
			continue
		}

		// next, check for the start of a range
		if char == '[' {
			escaped := false
			skip := 0
			for j := i + 1; j < len(c.compound); j++ {
				if escaped {
					escaped = false
					continue
				}

				charJ := c.compound[j]
				if charJ == '\\' {
					escaped = true
					continue
				}

				if charJ == ']' {
					skip = (j - i)
					break
				}
			}
			r := &regRange{
				rr:       c.rr,
				regRange: c.compound[i+1 : i+skip],
			}
			i += skip
			modifier, skip := extractModifier(c.compound, i+1)
			i += skip
			r.modifier = modifier
			splits = append(splits, r)
		}

		// finally, check for the start of a group
		if char == '(' {
			// Note, remove non-caputuring group or named capture group syntax first
		}

		// if it's none of those things, fallthrough. This probably shouldn't happen though.

	}
	return splits
}

// Determines whether a character belongs to the regexp reserved syntax
func reserved(c byte) bool {
	rcs := []byte{'[', '\\', '^', '$', '.', '|', '?', '*', '+', '(', ')'}
	found := false
	for _, rc := range rcs {
		if c == rc {
			found = true
			break
		}
	}
	return found
}

// Takes a string, and the index of character. Determines the modifier applies to that character
// and the length of said modifier.
func extractModifier(s string, i int) (string, int) {
	// After the last character, there can be no modifier.
	if i >= len(s) {
		return "", 0
	}

	char := s[i]
	if char == '?' || char == '*' || char == '+' {
		// Simple modifiers
		return string(char), 1
	} else if char == '{' {
		// TODO revisit this logic it is wrong! (comment is more accurate)
		// { can be a modifier, but only if it contains either a single digit,
		// or two digits seperated by a comma!
		skip := 0
		for j := i + 1; j < len(s); j++ {
			charJ := s[j]

			if charJ == '{' {
				break
			}

			if charJ == '}' {
				skip = (j - i) + 1
				break
			}
		}
		return s[i : i+skip], skip
	}

	return "", 0
}

// To solve a compound, first split it into its components, then solve each one, then rejoin.
func (c *compound) solve() (string, error) {
	splits := c.split()

	resPieces := []string{}
	for _, split := range splits {
		resPiece, err := split.solve()
		if err != nil {
			return "", err
		}
		resPieces = append(resPieces, resPiece)
	}

	return strings.Join(resPieces, ""), nil
}

// To solve a literal, examine its modifier, and repeat it that many times.
func (l *literal) solve() (string, error) {
	repeats, err := l.rr.repeats(l.modifier)
	if err != nil {
		return "", err
	}
	return strings.Repeat(l.literal, repeats), nil
}

// To solve a special, determine which special it is, then apply its special logic.
func (s *special) solve() (string, error) {
	return "", nil
}

// To solve a range, determine:
//   1. Are the | OR components?
//      - split them up
//      - pick one at random
//   2. Is it a range?
//      - ranges, use ASCII modulated characters.
//      - pick one at random
//   3. If not, simple split into component literals
//      - pick one at random
func (r *regRange) solve() (string, error) {
	repeats, err := r.rr.repeats(r.modifier)
	if err != nil {
		return "", err
	}

	components := []string{}
	lowIndex := 0
	escaped := false
	for i := 0; i < len(r.regRange); i++ {
		if escaped {
			escaped = false
			continue
		}

		char := r.regRange[i]
		if char == '\\' {
			escaped = true
			continue
		}
		if char == '|' {
			components = append(components, r.regRange[lowIndex:i])
			lowIndex = i + 1
		}
	}
	components = append(components, r.regRange[lowIndex:len(r.regRange)])

	fmt.Printf("%+v\n", components)

	resPieces := []string{}
	for i := 0; i < repeats; i++ {
		use := components[rand.Intn(len(components))]
		fmt.Println(use)

	}
	return strings.Join(resPieces, ""), nil
}

// A group is a compound, recursively solve its internal compound, the number of times
// dictated by the modifier.
func (g *group) solve() (string, error) {
	repeats, err := g.compound.rr.repeats(g.modifier)
	if err != nil {
		return "", err
	}

	res := []string{}
	for i := 0; i < repeats; i++ {
		solved, err := g.compound.solve()
		if err != nil {
			return "", err
		}
		res = append(res, solved)
	}

	return strings.Join(res, ""), nil
}

func (rr *RegexReverser) repeats(modifier string) (int, error) {
	var repeats int
	switch modifier {
	case "?":
		repeats = rand.Intn(2)
	case "*":
		repeats = rand.Intn(rr.maxRepeats + 1)
	case "+":
		repeats = rand.Intn(rr.maxRepeats) + 1
	case "":
		repeats = 1
	default:
		if modifier[0] == '{' && modifier[len(modifier)-1] == '}' {
			splits := strings.Split(modifier[1:len(modifier)-1], ",")
			if len(splits) == 1 {
				repeat64, err := strconv.ParseInt(splits[0], 0, 0)
				if err != nil {
					return 0, errors.Wrap(err, "invalid count modifier")
				}
				return int(repeat64), nil
			}

			if len(splits) != 2 {
				return 0, errors.Errorf("invalid count modifier %s", modifier)
			}
			min, err := strconv.ParseInt(splits[0], 0, 0)
			if err != nil {
				return 0, errors.Wrap(err, "invalid count modifier")
			}
			max, err := strconv.ParseInt(splits[1], 0, 0)
			if err != nil {
				return 0, errors.Wrap(err, "invalid count modifier")
			}
			repeats = rand.Intn(int(max)-int(min)) + int(min)
		} else {
			return 0, errors.Errorf("regrev doesn't know how to handle that modifier %s", modifier)
		}
	}

	return repeats, nil
}
