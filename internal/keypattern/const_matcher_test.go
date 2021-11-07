package keypattern

import (
	"testing"
)

func TestConstMatcher(t *testing.T) {
	var (
		v    = NewConstMatcher("var_")
		ectx = ExecContext{}
	)
	t.Run("positive", func(t *testing.T) {
		if off := v.Match("var_item", ectx); off != 4 {
			t.Errorf("invalid offset response %d != %d", off, 5)
		}
	})
	t.Run("negative", func(t *testing.T) {
		if off := v.Match("var1_item", ectx); off != -1 {
			t.Errorf("invalid negative offset response %d != %d", off, -1)
		}
	})
}
