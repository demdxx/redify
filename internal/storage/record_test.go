package storage

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecordReorganizeNested(t *testing.T) {
	testRecord := Record{
		"staff.name":   []string{"George", "John", "Brith"},
		"staff.age":    []string{"20", "21", "22"},
		"staff.gender": []string{"man", "man", "woman"},
	}
	expected := Record{
		"staff": []Record{
			{
				"name":   "George",
				"age":    "20",
				"gender": "man",
			},
			{
				"name":   "John",
				"age":    "21",
				"gender": "man",
			},
			{
				"name":   "Brith",
				"age":    "22",
				"gender": "woman",
			},
		},
	}
	res, err := testRecord.ReorganizeNested()
	if assert.NoError(t, err) && !reflect.DeepEqual(res, expected) {
		t.Errorf("Expected %v, got %v", expected, res)
	}
}
