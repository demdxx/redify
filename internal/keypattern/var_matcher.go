package keypattern

import "strings"

// FMatcher defines matching function
type FMatcher func(s string) (string, int)

func suffixMatcher(suffix string) FMatcher {
	return func(s string) (string, int) {
		size := len(suffix)
		for i := size; i < len(s)-size; i++ {
			if strings.HasSuffix(s[:i], suffix) {
				return s[:i-size], i - size
			}
		}
		return "", -1
	}
}

func fullMatcher(s string) (string, int) {
	return s, len(s)
}

type varMatcher struct {
	name string
	f    FMatcher
}

// NewVarMatcher for the specific variable in the pattern
func NewVarMatcher(name string, matcher FMatcher) Matcher {
	return &varMatcher{name: name, f: matcher}
}

func (m *varMatcher) String() string {
	return "{{" + m.name + "}}"
}

// Match variable and put into the context
func (m *varMatcher) Match(s string, ectx ExecContext) int {
	val, offset := m.f(s)
	if offset <= 0 {
		return -1
	}
	ectx[m.name] = val
	return offset
}
