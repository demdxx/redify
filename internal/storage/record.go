package storage

import (
	"encoding/json"
	"strings"

	"github.com/demdxx/gocast/v2"
	"github.com/pkg/errors"
)

var (
	ErrStructureOfTheRecordCannotBeReorganized = errors.New("structure of the record cannot be reorganized. All fields with the same prefix must be represented as a key with separator `.` in the key name and have the same size")
	ErrUnsupportedDatatype                     = errors.New("unsupported data type")
	ErrUnsupportedDatetypeConversion           = errors.New("unsupported data type conversion")
)

type DatatypeMapper struct {
	Name string
	Type string
}

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

// DatatypeCasting convert values to the specified types
func (r Record) DatatypeCasting(mappers ...DatatypeMapper) (_ Record, err error) {
	for _, mapper := range mappers {
		_, ok := r[mapper.Name]
		mType := strings.ToLower(mapper.Type)
		if ok {
			names := []string{mapper.Name}
			err = r.fieldCasting(r, names, names, mType)
		} else {
			names := strings.Split(mapper.Name, ".")
			err = r.fieldCasting(r, names, names, mType)
		}
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r Record) fieldCasting(record Record, field, completeField []string, dtype string) error {
	val := record[field[0]]
	if val == nil {
		if dtype == "json" {
			record[field[0]] = json.RawMessage("null")
		}
		return nil
	}
	if len(field) == 1 {
		var err error
		switch dtype {
		case "json":
			switch val.(type) {
			case json.RawMessage:
			case string, []byte:
				data := gocast.Str(val)
				if data != "" {
					// Validate JSON
					if err = json.Unmarshal([]byte(data), &[]any{nil}[0]); err == nil {
						record[field[0]] = json.RawMessage(data)
					}
				}
			default:
				var data []byte
				if data, err = json.Marshal(val); err == nil {
					record[field[0]] = json.RawMessage(data)
				}
			}
		case "string":
			record[field[0]], err = gocast.TryStr(val)
		case "int":
			record[field[0]], err = gocast.TryNumber[int64](val)
		case "float":
			record[field[0]], err = gocast.TryNumber[float64](val)
		case "bool":
			// TODO: add boolean type check
			record[field[0]] = gocast.Bool(val)
		default:
			err = errors.Wrap(ErrUnsupportedDatatype, dtype)
		}
		return err
	}
	rec, _ := val.([]Record)
	if rec == nil {
		return errors.Wrap(ErrUnsupportedDatetypeConversion, strings.Join(completeField, "."))
	}
	for i := range rec {
		if err := r.fieldCasting(rec[i], field[1:], completeField, dtype); err != nil {
			return err
		}
	}
	return nil
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
