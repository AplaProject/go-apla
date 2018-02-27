package ini

import (
	"bytes"
	"fmt"
	"io"
)

// manages all the key/value defined in the .ini file format
type Section struct {
	//Name of the section
	Name string
	//key values
	keyValues map[string]Key
}

// construct a new section with section name
func NewSection(name string) *Section {
	return &Section{Name: name,
		keyValues: make(map[string]Key)}
}

// add key/value to the section and overwrite the old one
func (section *Section) Add(key, value string) {
	section.keyValues[key] = newNormalKey(key, value)
}

// check if the key is in the section
//
// return true if the section contains the key
func (section *Section) HasKey(key string) bool {
	_, ok := section.keyValues[key]
	return ok
}

// Get all the keys in the section
//
// return: all keys in the section
func (section *Section) Keys() []Key {
	r := make([]Key, 0)
	for _, v := range section.keyValues {
		r = append(r, v)
	}
	return r
}

// Get the key.
//
// This method can be called even if the key is not in the
// section.
func (section *Section) Key(key string) Key {
	if v, ok := section.keyValues[key]; ok {
		return v
	}
	return newNonExistKey(key)
}

// Get value of key as string
func (section *Section) GetValue(key string) (string, error) {
	return section.Key(key).Value()
}

// Get value of key and if the key does not exist, return the defValue
func (section *Section) GetValueWithDefault(key string, defValue string) string {
	return section.Key(key).ValueWithDefault(defValue)
}

// Get the value of key as bool, it will return true if the value of the key is one
// of following( case insensitive):
//  - true
//  - yes
//  - t
//  - y
//  - 1
func (section *Section) GetBool(key string) (bool, error) {
	return section.Key(key).Bool()
}

// Get the value of key as bool and if the key does not exist, return the
// default value
func (section *Section) GetBoolWithDefault(key string, defValue bool) bool {
	return section.Key(key).BoolWithDefault(defValue)
}

// Get the value of the key as int
func (section *Section) GetInt(key string) (int, error) {
	return section.Key(key).Int()
}

// Get the value of the key as int and if the key does not exist return
// the default value
func (section *Section) GetIntWithDefault(key string, defValue int) int {
	return section.Key(key).IntWithDefault(defValue)
}

// Get the value of the key as uint
func (section *Section) GetUint(key string) (uint, error) {
	return section.Key(key).Uint()
}

// Get the value of the key as int and if the key does not exist return
// the default value
func (section *Section) GetUintWithDefault(key string, defValue uint) uint {
	return section.Key(key).UintWithDefault(defValue)
}

// Get the value of the key as int64
func (section *Section) GetInt64(key string) (int64, error) {
	return section.Key(key).Int64()
}

// Get the value of the key as int64 and if the key does not exist return
// the default value
func (section *Section) GetInt64WithDefault(key string, defValue int64) int64 {
	return section.Key(key).Int64WithDefault(defValue)
}

// Get the value of the key as uint64
func (section *Section) GetUint64(key string) (uint64, error) {
	return section.Key(key).Uint64()
}

// Get the value of the key as uint64 and if the key does not exist return
// the default value
func (section *Section) GetUint64WithDefault(key string, defValue uint64) uint64 {
	return section.Key(key).Uint64WithDefault(defValue)
}

// Get the value of the key as float32
func (section *Section) GetFloat32(key string) (float32, error) {
	return section.Key(key).Float32()
}

// Get the value of the key as float32 and if the key does not exist return
// the default value
func (section *Section) GetFloat32WithDefault(key string, defValue float32) float32 {
	return section.Key(key).Float32WithDefault(defValue)
}

// Get the value of the key as float64
func (section *Section) GetFloat64(key string) (float64, error) {
	return section.Key(key).Float64()
}

// Get the value of the key as float64 and if the key does not exist return
// the default value
func (section *Section) GetFloat64WithDefault(key string, defValue float64) float64 {
	return section.Key(key).Float64WithDefault(defValue)
}

// convert the section content to the .ini section format, so the section content will
// be converted to following format:
//
//  [sectionx]
//  key1 = value1
//  key2 = value2
//
func (section *Section) String() string {
	buf := bytes.NewBuffer(make([]byte, 0))
	section.Write(buf)
	return buf.String()
}

// write the section content to the writer with .ini section format.
func (section *Section) Write(writer io.Writer) error {
	_, err := fmt.Fprintf(writer, "[%s]\n", section.Name)
	if err != nil {
		return err
	}
	for _, v := range section.keyValues {
		_, err = fmt.Fprintf(writer, "%s\n", v.String())
		if err != nil {
			return err
		}
	}
	return nil
}
