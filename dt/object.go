package dt

// Object .
type Object struct {
	ObjType  uint8
	Encoding uint8
	Lru      uint64
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
