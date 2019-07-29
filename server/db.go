package server

import (
	"github.com/kzinglzy/godis/dt"
)

// Database implements the kv service
type Database struct {
	store   *dt.Dict
	expires *dt.Dict
}

// NewDatabase .
func NewDatabase() *Database {
	return &Database{
		store:   dt.NewDict(),
		expires: dt.NewDict(),
	}
}

// Get Lookups a key, and as a side effect, if needed,
// expires the key if its TTL is reached.
func (db *Database) Get(key string) *dt.Object {
	return db.lookupKey(key, true)
}

func (db *Database) lookupKey(key string, touch bool) *dt.Object {
	if db.expireIfNeeded(key) {
		return nil
	}

	de := db.store.Get(key)
	if de == nil {
		return nil
	}

	val := de.Value.(*dt.Object)
	if !godisServer.isSaving() && touch {
		db.updateLRU(val)
	}
	return val
}

func (db *Database) expireIfNeeded(key string) bool {
	if !db.keyIsExpired(key) {
		return false
	}

	db.deleteKey(key)
	return true
}

func (db *Database) setKey(key string, obj *dt.Object) {
	old := db.lookupKey(key, true)
	if old == nil {
		db.store.Add(key, obj)
	} else {
		old.Ptr = obj.Ptr
		old.RefCount++
		db.expires.Delete(key)
	}
}

func (db *Database) deleteKey(key string) {
	db.expires.Delete(key)
	db.store.Delete(key)
}

func (db *Database) keyIsExpired(key string) bool {
	when := db.getExpire(key)
	if when < 0 {
		return false
	}

	return mstime() > when
}

func (db *Database) getExpire(key string) int64 {
	de := db.expires.Get(key)
	if de == nil {
		return -1
	}
	return de.Value.(int64)
}

func (db *Database) setExpire(key string, when int64) {
	db.expires.Add(key, when)
}

func (db *Database) ttl(key string) int64 {
	if db.lookupKey(key, false) == nil {
		return -2
	}

	var ttl int64
	expire := db.getExpire(key)
	if expire != -1 {
		ttl = expire - mstime()
		if ttl < 0 {
			ttl = 0
		}
		return (ttl + 500) / 1000
	}

	return -1
}

func (db *Database) updateLRU(o *dt.Object) {
	// TODO memory Policy
}
