package keypattern

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/demdxx/gocast"
)

type ValueType int

const (
	ConstValueType ValueType = 1
	VarValueType   ValueType = 2
	varReplacer              = "$%^REPLACE^%$"
)

var patternRegex = regexp.MustCompile(`\{\{([^\}]*)\}\}`)

type ValueGetter interface {
	Get(key string) interface{}
}

type ExecContext map[string]string

func (c ExecContext) Get(key string) interface{} {
	if c == nil {
		return nil
	}
	return c[key]
}

type Matcher interface {
	Match(s string, ectx ExecContext) (offset int)
}

// Pattern matcher type
type Pattern struct {
	matchers []Matcher
}

// NewPattern object by matchers
func NewPattern(matchers ...Matcher) *Pattern {
	return &Pattern{
		matchers: matchers,
	}
}

// NewPatternFromExpression parse expression
// {{varname}} - variable definition in the pattern
// Example:
//   posts_{{id}}
func NewPatternFromExpression(expression string) *Pattern {
	var (
		matchers []Matcher
		vars     = ExtractVars(expression)
		varIdx   = 0
	)
	for _, vr := range vars {
		expression = strings.ReplaceAll(expression, vr[0], varReplacer)
	}
	consts := strings.Split(expression, varReplacer)
	for i := 0; i < len(consts); i++ {
		if consts[i] != "" {
			matchers = append(matchers, NewConstMatcher(consts[i]))
		}
		if varIdx < len(vars) {
			var expMatcher FMatcher = fullMatcher
			if i < len(consts)-1 && consts[i+1] != "" {
				expMatcher = suffixMatcher(consts[i+1])
			}
			matchers = append(matchers, NewVarMatcher(vars[varIdx][1], expMatcher))
			varIdx += 1
		}
	}
	return &Pattern{
		matchers: matchers,
	}
}

func (p *Pattern) String() string {
	var buf bytes.Buffer
	for _, mt := range p.matchers {
		switch m := mt.(type) {
		case *varMatcher:
			buf.WriteString(m.name)
		case *constMatcher:
			buf.WriteString(m.val)
		default:
			buf.WriteString(fmt.Sprintf("%s", mt))
		}
	}
	return buf.String()
}

func (p *Pattern) Format(vals ValueGetter) string {
	var buf bytes.Buffer
	for _, mt := range p.matchers {
		switch m := mt.(type) {
		case *varMatcher:
			buf.WriteString(gocast.ToString(vals.Get(m.name)))
		case *constMatcher:
			buf.WriteString(m.val)
		default:
			buf.WriteString(fmt.Sprintf("%s", mt))
		}
	}
	return buf.String()
}

// Match input string with the pattern
func (p *Pattern) Match(s string, ectx ExecContext) bool {
	offset := 0
	for _, m := range p.matchers {
		off := m.Match(s[offset:], ectx)
		if off < 1 {
			return false
		}
		offset += off
	}
	return true
}

// Keys of the pattern
func (p *Pattern) Keys() []string {
	keys := make([]string, 0, len(p.matchers))
	for _, mt := range p.matchers {
		switch m := mt.(type) {
		case *varMatcher:
			keys = append(keys, m.name)
		}
	}
	return keys
}

// ExtractVars from the string expression
// Return data: [["{{key1}}", "key1"], ["{{key2}}", "key2"]]
func ExtractVars(expression string) [][]string {
	matches := patternRegex.FindAllStringSubmatch(expression, -1)
	keys := map[string]bool{}
	newMatches := make([][]string, 0, len(matches))
	for _, v := range matches {
		if !keys[v[0]] {
			keys[v[0]] = true
			newMatches = append(newMatches, v)
		}
	}
	return newMatches
}
