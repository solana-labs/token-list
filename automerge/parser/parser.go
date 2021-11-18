package parser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tailscale/hujson"
	"regexp"
	"strings"
)

// A wild custom JSON parser appears.
// May the gods have mercy!

var ErrUnexpectedClosingBracket = errors.New("unexpected closing bracket")

func NormalizeWhatever(s string) ([]Token, error) {
	// First, strip all control characters from the start and end.
	s = strings.Trim(s, "\t\n {},[]")

	// Put back the outer braces.
	s = fmt.Sprintf("{\n%s", s)

	// Fix up truncated [] or {}
	if len(s) > 2 && s[len(s)-1] == ':' {
        s = s + " null"
    }

	// Then, iterate over the lines of s (assuming that the JSON is at least
	// properly line-delimited) and deterministically add the closing braces
	// we just stripped.
	queue := make([]rune, 0)

	for _, l := range strings.Split(s, "\n") {
		for _, c := range l {
			if c == '{' {
				queue = append(queue, '}')
			}
			if c == '[' {
				queue = append(queue, ']')
			}
			if c == ']' {
				if len(queue) == 0 {
                    return nil, ErrUnexpectedClosingBracket
                }
				if queue[len(queue)-1] == ']' {
					queue = queue[:len(queue)-1]
                } else {
                    return nil, ErrUnexpectedClosingBracket
				}
			}
			if c == '}' {
				if len(queue) == 0 {
                    return nil, ErrUnexpectedClosingBracket
                }
				if queue[len(queue)-1] == '}' {
					queue = queue[:len(queue)-1]
                } else {
                    return nil, ErrUnexpectedClosingBracket
				}
			}
		}
	}

	if len(queue) > 0 {
		for i := len(queue); i > 0; i-- {
            s += "\n" + string(queue[i-1]) + ","
		}
    }

	// Put trailing commas everywhere
	s = regexp.MustCompile(`(?m)["}]$`).ReplaceAllString(s, "$0,")

	// Remove last trailing comma
	s = strings.TrimRight(s, ",")

	// Figure out whether to parse multiple objects
	var multi bool
	if strings.Count(s, `"chainId"`) > 1 {
		multi = true
		s = fmt.Sprintf("[\n%s\n]", s)
	}

	// Preprocess using custom JSON parser that ignores trailing commas
	ast, err := hujson.Parse([]byte(s))
	if err != nil {
		return nil, fmt.Errorf("failed to normalize JSON: %v", err)
	}
	ast.Standardize()
	b := ast.Pack()

	var tt []Token
	dec := json.NewDecoder(bytes.NewBuffer(b))
	dec.DisallowUnknownFields()
	if multi {
		if err := dec.Decode(&tt); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %v", err)
		}
	} else {
		var t Token
		if err := dec.Decode(&t); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %v", err)
		}
		tt = []Token{t}
	}

	return tt, nil
}
