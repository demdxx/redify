package keypattern

import (
	"testing"
)

func TestPattern(t *testing.T) {
	var (
		p1 = NewPatternFromExpression("vars_{{var1}}${{var2}}")
		p2 = NewPattern(
			NewConstMatcher("vars_"),
			NewVarMatcher("var1", suffixMatcher("$")),
			NewConstMatcher("$"),
			NewVarMatcher("var2", fullMatcher),
		)
		ectx = ExecContext{}
	)
	if p1.String() != p2.String() {
		t.Error("patterns must be equal")
	}
	if !p2.Match("vars_val1$val2", ectx) || !p1.Match("vars_val1$val2", ectx) {
		t.Error("invalid matrching")
	}
	if ectx["var1"] != "val1" {
		t.Error("invalid var1")
	}
	if ectx["var2"] != "val2" {
		t.Error("invalid var2")
	}
	if p2.Format(ExecContext{"var1": "v1", "var2": "v2"}) != "vars_v1$v2" {
		t.Error("patterns format invalid")
	}
	if len(p2.Keys()) != 2 {
		t.Error("patterns number of keys are incorrect")
	}
}
