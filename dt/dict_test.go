package dt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
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
	require.Equal(t, int64(0), d.size())
	require.Equal(t, HtInitialSize-1, d.getMaxSizeMask())
	require.False(t, d.isRehashing())
	require.True(t, d.keyIndexToPopulated("nonexist") != -1)
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
	}

	d := NewDict()
	for _, tC := range testCases {
		t.Run("", func(t *testing.T) {
			d.Add(tC.k, tC.v)
			// require.Equal(t, d.Get(tC.k).value, tC.v)
			fmt.Println(d.Get(""))
		})
	}
}

// func TestkeyIndexToPopulated(T *testing.T) {

// }

// func TestDictRehashing(t *testing.T) {

// }

// func TestDictGetRandomKeys(t *testing.T) {

// }
