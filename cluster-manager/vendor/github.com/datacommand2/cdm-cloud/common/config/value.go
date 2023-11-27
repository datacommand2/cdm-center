package config

import "strconv"

// Value 는 Global Config Property 의 값이다.
type Value string

// Bool 은 Value 를 boolean 으로 변환하는 함수이다.
func (v *Value) Bool() (bool, error) {
	return strconv.ParseBool(string(*v))
}

// Int64 는 Value 를 int64 로 변환하는 함수이다.
func (v *Value) Int64() (int64, error) {
	return strconv.ParseInt(string(*v), 10, 64)
}

// Uint64 는 Value 를 uint64 로 변환하는 함수이다.
func (v *Value) Uint64() (uint64, error) {
	return strconv.ParseUint(string(*v), 10, 64)
}

// String 은 Value 를 string 으로 변환하는 함수이다.
func (v *Value) String() string {
	return string(*v)
}

// Float64 은 Value 를 float64 로 변환하는 함수이다.
func (v *Value) Float64() (float64, error) {
	return strconv.ParseFloat(string(*v), 64)
}
