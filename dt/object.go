package dt

import "time"

// Object .
type Object struct {
	ObjType  uint8
	Encoding uint8
	Lru      int64
	Ptr      interface{}
}

// actual Redis Object
const (
	ObjString = iota
	ObjList
	ObjZSet
	ObjHash
)

// objects encoding
const (
	ObjEncodingRaw = iota
	ObjEncodingInt
	ObjEncodingHt
	ObjEncodingList
	ObjEncodingSkiplist
)

// NewObj .
func NewObj(t uint8, v interface{}) *Object {
	return &Object{
		ObjType:  t,
		Encoding: ObjEncodingRaw,
		Lru:      time.Now().Unix(),
		Ptr:      v,
	}
}

// NewList .
func NewList(t uint8, v []string) *Object {
	obj := NewObj(t, v)
	obj.Encoding = ObjEncodingList
	return obj
}
