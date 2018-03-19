package ini

import (
	"fmt"
	"strconv"
	"strings"
)

// represents the <key, value> pair stored in the
// section of the .ini file
//
type Key interface {
	// get name of the key
	Name() string

	// get value of the key
	Value() (string, error)

	//get the value of key and return defValue if
	//the value does not exist
	ValueWithDefault(defValue string) string

	// get the value as bool
	// return true if the value is one of following(case insensitive):
	// - true
	// - yes
	// - T
	// - Y
	// - 1
	// Any other value will return false
	Bool() (bool, error)

	// get the value as bool and return the defValue if the
	// value of the key does not exist
	BoolWithDefault(defValue bool) bool
	// get the value as int
	Int() (int, error)

	// get value as int and return defValue if the
	// value of the key does not exist
	IntWithDefault(defValue int) int

	//get value as uint
	Uint() (uint, error)

	//get value as uint and return defValue if the
	//key does not exist or it is not uint format
	UintWithDefault(defValue uint) uint

	// get the value as int64
	Int64() (int64, error)

	// get the value as int64 and return defValue
	// if the value of the key does not exist
	Int64WithDefault(defValue int64) int64

	// get the value as uint64
	Uint64() (uint64, error)

	// get the value as uint64 and return defValue
	// if the value of the key does not exist
	Uint64WithDefault(defValue uint64) uint64

	// get the value as float32
	Float32() (float32, error)

	// get the value as float32 and return defValue
	// if the value of the key does not exist
	Float32WithDefault(defValue float32) float32

	// get the value as float64
	Float64() (float64, error)

	// get the value as the float64 and return defValue
	// if the value of the key does not exist
	Float64WithDefault(defValue float64) float64

	// return a string as "key=value" format
	// and if no value return empty string
	String() string
}

type nonExistKey struct {
	name string
}

func newNonExistKey(name string) *nonExistKey {
	return &nonExistKey{name: name}
}

func (nek *nonExistKey) Name() string {
	return nek.name
}

func (nek *nonExistKey) Value() (string, error) {
	return "", nek.noSuchKey()
}

func (nek *nonExistKey) ValueWithDefault(defValue string) string {
	return defValue
}

func (nek *nonExistKey) Bool() (bool, error) {
	return false, nek.noSuchKey()
}

func (nek *nonExistKey) BoolWithDefault(defValue bool) bool {
	return defValue
}

func (nek *nonExistKey) Int() (int, error) {
	return 0, nek.noSuchKey()
}

func (nek *nonExistKey) IntWithDefault(defValue int) int {
	return defValue
}

func (nek *nonExistKey) Uint() (uint, error) {
	return 0, nek.noSuchKey()
}

func (nek *nonExistKey) UintWithDefault(defValue uint) uint {
	return defValue
}

func (nek *nonExistKey) Int64() (int64, error) {
	return 0, nek.noSuchKey()
}

func (nek *nonExistKey) Int64WithDefault(defValue int64) int64 {
	return defValue
}

func (nek *nonExistKey) Uint64() (uint64, error) {
	return 0, nek.noSuchKey()
}

func (nek *nonExistKey) Uint64WithDefault(defValue uint64) uint64 {
	return defValue
}

func (nek *nonExistKey) Float32() (float32, error) {
	return .0, nek.noSuchKey()
}

func (nek *nonExistKey) Float32WithDefault(defValue float32) float32 {
	return defValue
}

func (nek *nonExistKey) Float64() (float64, error) {
	return .0, nek.noSuchKey()
}

func (nek *nonExistKey) Float64WithDefault(defValue float64) float64 {
	return defValue
}

func (nek *nonExistKey) String() string {
	return ""
}

func (nek *nonExistKey) noSuchKey() error {
	return fmt.Errorf("no such key:%s", nek.name)
}

type normalKey struct {
	name  string
	value string
}

var trueBoolValue = map[string]bool{"true": true, "t": true, "yes": true, "y": true, "1": true}

func newNormalKey(name, value string) *normalKey {
	return &normalKey{name: name, value: replace_env(value)}
}

func (k *normalKey) Name() string {
	return k.name
}

func (k *normalKey) Value() (string, error) {
	return k.value, nil
}

func (k *normalKey) ValueWithDefault(defValue string) string {
	return k.value
}

func (k *normalKey) Bool() (bool, error) {
	if _, ok := trueBoolValue[strings.ToLower(k.value)]; ok {
		return true, nil
	}
	return false, nil
}

func (k *normalKey) BoolWithDefault(defValue bool) bool {
	v, err := k.Bool()
	if err == nil {
		return v
	}
	return defValue
}

func (k *normalKey) Int() (int, error) {
	return strconv.Atoi(k.value)
}

func (k *normalKey) IntWithDefault(defValue int) int {
	i, err := strconv.Atoi(k.value)
	if err == nil {
		return i
	}
	return defValue
}

func (k *normalKey) Uint() (uint, error) {
	v, err := strconv.ParseUint(k.value, 0, 32)
	return uint(v), err
}

func (k *normalKey) UintWithDefault(defValue uint) uint {
	i, err := k.Uint()
	if err == nil {
		return i
	}
	return defValue

}

func (k *normalKey) Int64() (int64, error) {
	return strconv.ParseInt(k.value, 0, 64)
}

func (k *normalKey) Int64WithDefault(defValue int64) int64 {
	i, err := strconv.ParseInt(k.value, 0, 64)
	if err == nil {
		return i
	}
	return defValue
}

func (k *normalKey) Uint64() (uint64, error) {
	return strconv.ParseUint(k.value, 0, 64)
}

func (k *normalKey) Uint64WithDefault(defValue uint64) uint64 {
	i, err := strconv.ParseUint(k.value, 0, 64)
	if err == nil {
		return i
	}
	return defValue
}

func (k *normalKey) Float32() (float32, error) {
	f, err := strconv.ParseFloat(k.value, 32)
	return float32(f), err
}

func (k *normalKey) Float32WithDefault(defValue float32) float32 {
	f, err := strconv.ParseFloat(k.value, 32)
	if err == nil {
		return float32(f)
	}
	return defValue
}

func (k *normalKey) Float64() (float64, error) {
	return strconv.ParseFloat(k.value, 64)
}

func (k *normalKey) Float64WithDefault(defValue float64) float64 {
	f, err := strconv.ParseFloat(k.value, 64)
	if err == nil {
		return f
	}
	return defValue
}

func (k *normalKey) String() string {
	return fmt.Sprintf("%s=%s", k.name, toEscape(k.value))
}
