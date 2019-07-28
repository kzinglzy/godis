package dt

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func initData(d *Dict) int64 {
	d.Add("a", 1)
	d.Add("b", 2)
	d.Add("c", 1)
	d.Add("d", "helloworld")
	d.Add("", "empty")

	return 5
}

func TestDictCreate(t *testing.T) {
	d := NewDict()
	assert.Equal(t, int64(0), d.Size())
	assert.Equal(t, HtInitialSize-1, d.getMaxSizeMask())
	assert.False(t, d.isRehashing())

	idx, duplicated := d.keyIndexToPopulated("foo")
	assert.Nil(t, duplicated)
	assert.True(t, idx != -1)
}

func TestBaseOperations(t *testing.T) {
	testCases := []struct {
		k string
		v interface{}
	}{
		{
			k: "a",
			v: 1,
		},
		{
			k: "b",
			v: 2,
		},
		{
			k: "c",
			v: "-3",
		},
		{
			k: "d",
			v: "foobar",
		},
		{
			k: "",
			v: 123,
		},
		{
			k: "",
			v: 0,
		},
		{
			k: "a",
			v: "a",
		},
		{
			k: "a",
			v: "a",
		},
		{
			k: "longlonglonglonglonglonglonglonglonglonglonglonglong",
			v: -1000010000100001000,
		},
	}

	d := NewDict()
	for _, tC := range testCases {
		t.Run("", func(t *testing.T) {
			d.Add(tC.k, tC.v)
			assert.Equal(t, d.Get(tC.k).Value, tC.v)
		})
	}

}

func TestDictRehashing(t *testing.T) {
	d := NewDict()
	n := int(math.Pow(2, 20))
	for i := 0; i < n; i++ {
		d.Add(string(i), i)
	}

	assert.False(t, d.isRehashing())
	assert.Equal(t, int64(n), d.Size())

	d.Add("lastElementToRehashing", 1)
	assert.True(t, d.isRehashing())

	preRehashIDX := d.rehashIndex
	d.Get("1") // trigger a step rehashing
	assert.True(t, d.rehashIndex > preRehashIDX)

	d.doRehashing(int(d.hts[0].size - d.rehashIndex))
	assert.False(t, d.isRehashing())
	assert.Equal(t, HtInitialSize, d.hts[1].size)
	assert.Equal(t, int64(0), d.hts[1].used)
	assert.Equal(t, int64(n)+1, d.hts[0].used)
}

func TestDelete(t *testing.T) {
	d := NewDict()
	n := 100000
	for i := 0; i < n; i++ {
		d.Add(string(i), i)
	}

	for i := 0; i < n; i += 2 {
		d.Delete(string(i))
	}

	assert.Equal(t, int64(n/2), d.Size())
}

func TestDictGetRandomKeys(t *testing.T) {
	d := NewDict()
	n := 1000000
	for i := 0; i < n; i++ {
		d.Add(string(i), i)
	}

	j := n / 2
	m := make(map[string]bool)
	duplicated := 0
	for i := 0; i < j; i++ {
		key := d.GetRandomKey().Key
		if _, found := m[key]; found {
			duplicated++
		} else {
			m[key] = true
		}
	}

	d.GetSomeKeys(int64(j))
}
