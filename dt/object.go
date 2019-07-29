package dt

import "time"

// Object .
type Object struct {
	ObjType  uint8
	Encoding uint8
	Lru      int64
	RefCount uint
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
	ObjEncodingLinkedlist
	ObjEncodingSkiplist
)

// NewObj .
func NewObj(t uint8, v interface{}) *Object {
	return &Object{
		ObjType:  t,
		Encoding: ObjEncodingRaw,
		Lru:      time.Now().Unix(),
		RefCount: 1,
		Ptr:      v,
	}
}
