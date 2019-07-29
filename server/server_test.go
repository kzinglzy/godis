package server

import (
	"testing"

	"github.com/kzinglzy/godis/dt"
	"github.com/stretchr/testify/assert"
)

func init() {
	MakeServer(":6666")
}

func TestDatabase(t *testing.T) {
	db := NewDatabase()

	testCases := []struct {
		k   string
		v   string
		ttl int64
		ref uint
	}{
		{
			k:   "a",
			v:   "1",
			ttl: 10,
			ref: 1,
		},
		{
			k:   "a",
			v:   "2",
			ttl: 10,
			ref: 2,
		},
	}
	for _, tC := range testCases {
		t.Run("", func(t *testing.T) {
			obj := dt.NewObj(dt.ObjString, tC.v)
			db.setKey(tC.k, obj)
			db.setExpire(tC.k, mstime()+tC.ttl*1000)

			v := db.Get(tC.k)
			assert.Equal(t, tC.v, v.Ptr.(string))
			assert.Equal(t, tC.ref, v.RefCount)
			assert.True(t, db.getExpire(tC.k) != -1)
			assert.False(t, db.keyIsExpired(tC.k))
			assert.False(t, db.expireIfNeeded(tC.k))
			assert.True(t, db.ttl(tC.k) > 0)
			assert.Nil(t, db.Get("unknown"))
		})
	}
}
