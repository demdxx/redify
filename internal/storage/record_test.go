package storage

import (
	"encoding/json"
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

func TestDatetypeCasting(t *testing.T) {
	testRecord := Record{
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
		"raw": []Record{
			{"json": map[string]any{}},
			{"json": map[string]any{"name": "John"}},
			{"json": json.RawMessage(`{"name": "Brith", "age": 22}`)},
		},
		"json.field": nil,
	}
	expected := Record{
		"staff": []Record{
			{
				"name":   "George",
				"age":    int64(20),
				"gender": "man",
			},
			{
				"name":   "John",
				"age":    int64(21),
				"gender": "man",
			},
			{
				"name":   "Brith",
				"age":    int64(22),
				"gender": "woman",
			},
		},
		"raw": []Record{
			{"json": json.RawMessage(`{}`)},
			{"json": json.RawMessage(`{"name":"John"}`)},
			{"json": json.RawMessage(`{"name": "Brith", "age": 22}`)},
		},
		"json.field": json.RawMessage(`null`),
	}
	res, err := testRecord.DatatypeCasting(
		DatatypeMapper{Name: "staff.age", Type: "int"},
		DatatypeMapper{Name: "raw.json", Type: "json"},
		DatatypeMapper{Name: "json.field", Type: "json"})
	if assert.NoError(t, err) && !reflect.DeepEqual(res, expected) {
		t.Errorf("Expected %v, got %v", expected, res)
	}
}
