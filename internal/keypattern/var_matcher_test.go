package keypattern

import (
	"testing"
)

func TestMatchers(t *testing.T) {
	var tests = []struct {
		m      FMatcher
		s      string
		rs     string
		offset int
	}{
		{
			m:      suffixMatcher("_$_"),
			s:      "value_$_next",
			rs:     "value",
			offset: 5,
		},
		{
			m:      fullMatcher,
			s:      "value",
			rs:     "value",
			offset: 5,
		},
	}
	for _, test := range tests {
		r, off := test.m(test.s)
		if r != test.rs {
			t.Errorf("invalid value response %s != %s", r, test.rs)
		}
		if off != test.offset {
			t.Errorf("invalid offset response %d != %d", off, test.offset)
		}
	}
}

func TestVarMatcher(t *testing.T) {
	v := NewVarMatcher("var", suffixMatcher("_$_"))
	ectx := ExecContext{}
	off := v.Match("value_$_next", ectx)
	if off != 5 {
		t.Errorf("invalid offset response %d != %d", off, 5)
	}
	if ectx["var"] != "value" {
		t.Errorf("invalid context state %v != %v", ectx["var"], "value")
	}
}
