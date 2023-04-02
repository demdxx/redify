package storage

import (
	"errors"
	"strings"

	"github.com/demdxx/gocast/v2"
)

var ErrStructureOfTheRecordCannotBeReorganized = errors.New("structure of the record cannot be reorganized. All fields with the same prefix must be represented as a key with separator `.` in the key name and have the same size")

// Record is a map of key-value pairs
type Record map[string]any

// Get value by key
func (r Record) Get(key string) any {
	if r == nil {
		return nil
	}
	return r[key]
}

// ReorganizeNested record according to the key pattern
// If some fields represented as a key with separator `.` in the key name
// it need to combine with the other fields with the same prefix
// Example:
//
//	{
//		"staff.name": ["George", "John", "Brith"],
//		"staff.age": ["20", "21", "22"],
//		"staff.gender": ["man", "man", "woman"],
//	}
//
// After reorganize:
//
//		{
//			"staff": [
//				{
//					"name": "George",
//					"age": "20",
//				  "gender": "man",
//				},
//				{
//					"name": "John",
//					"age": "21",
//	 		  "gender": "man",
//				},
//				{
//					"name": "Brith",
//					"age": "22",
//	       "gender": "woman",
//				},
//			],
//		}
func (r Record) ReorganizeNested() (Record, error) {
	res := make(Record, len(r))
	for key, value := range r {
		path := strings.Split(key, ".")
		if len(path) == 1 {
			res[key] = value
		} else {
			if gocast.IsSlice(value) {
				if res[path[0]] == nil {
					res[path[0]] = make([]Record, 0, 5)
				} else if !gocast.IsSlice(res[path[0]]) {
					return nil, ErrStructureOfTheRecordCannotBeReorganized
				}
				for i, v := range gocast.Cast[[]any](value) {
					if i > len(res[path[0]].([]Record))-1 {
						res[path[0]] = append(res[path[0]].([]Record), make(Record, 5))
					}
					aggregateWithPrefix(res[path[0]].([]Record)[i], path[1:], v)
				}
			} else {
				aggregateWithPrefix(res, path, value)
			}
		}
	}
	return res, nil
}

func aggregateWithPrefix(res Record, path []string, value any) {
	if len(path) == 1 {
		res[path[0]] = value
		return
	}
	if res[path[0]] == nil {
		res[path[0]] = make(Record, 5)
	}
	aggregateWithPrefix(res[path[0]].(Record), path[1:], value)
}
