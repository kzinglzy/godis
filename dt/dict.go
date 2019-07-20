package dt

// Dict .
type Dict struct {
	dictType    dicttype
	ht          hashtable
	rehashIndex int
}

type hashtable struct {
	table entry
	size  int
	used  int
}

type entry struct {
	key   interface{}
	value interface{}
	next  *entry
}

type dicttype struct {
	hashFunc func(interface{}) int64
}

// dict add
// dict get
// dict delete
// dict rehash
// dict getsomekeys
