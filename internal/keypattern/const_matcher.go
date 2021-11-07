package keypattern

import "strings"

type constMatcher struct {
	val string
}

// NewConstMatcher defines constant part of the pattern
func NewConstMatcher(cval string) Matcher {
	return &constMatcher{val: cval}
}

func (m *constMatcher) String() string {
	return m.val
}

// Match variable const expression
func (m *constMatcher) Match(s string, _ ExecContext) (offset int) {
	if strings.HasPrefix(s, m.val) {
		return len(m.val)
	}
	return -1
}
