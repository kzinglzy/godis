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
		ref uint
	}{
		{
			k:   "a",
			v:   "1",
			ref: 1,
		},
		{
			k:   "a",
			v:   "2",
			ref: 2,
		},
	}
	for _, tC := range testCases {
		t.Run("", func(t *testing.T) {
			obj := dt.NewObj(dt.ObjString, tC.v)
			db.setKey(tC.k, obj)

			v := db.Get(tC.k)
			assert.Equal(t, tC.v, v.Ptr.(string))
			assert.Equal(t, tC.ref, v.RefCount)
		})
	}
}
