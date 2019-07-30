package dt

import (
	"hash/fnv"
	"math/rand"
	"time"
)

const (
	HtInitialSize int64 = 4
)

// Dict .
type Dict struct {
	hts         [2]*hashtable
	rehashIndex int64
}

type hashtable struct {
	table    []*Entry
	size     int64
	sizemask int64
	used     int64
}

type Entry struct {
	Key   string
	Value interface{}
	next  *Entry
}

func NewDict() *Dict {
	hts := [2]*hashtable{newHashTable(HtInitialSize), newHashTable(HtInitialSize)}
	return &Dict{
		hts:         hts,
		rehashIndex: -1,
	}
}

func newHashTable(size int64) *hashtable {
	return &hashtable{
		table:    make([]*Entry, size, size),
		size:     size,
		sizemask: size - 1,
		used:     0,
	}
}

func (dt *Dict) Add(key string, value interface{}) {
	if dt.IsRehashing() {
		dt.doRehashing(1)
	}

	idx, old := dt.keyIndexToPopulated(key)
	if old != nil {
		old.Value = value // overwrite
		return
	}

	var ht *hashtable
	if dt.IsRehashing() {
		ht = dt.hts[1]
	} else {
		ht = dt.hts[0]
	}

	entry := &Entry{
		Key:   key,
		Value: value,
		next:  nil,
	}

	he := ht.table[idx]
	if he == nil {
		ht.table[idx] = entry
	} else {
		prev := he
		for he != nil {
			prev = he
			he = he.next
		}
		prev.next = entry
	}
	ht.used++
}

func (dt *Dict) Get(key string) *Entry {
	if dt.Used() == 0 {
		return nil
	}

	if dt.IsRehashing() {
		dt.doRehashing(1)
	}

	for _, ht := range dt.hts {
		idx := dt.hash(key) & ht.sizemask
		he := ht.table[idx]

		for he != nil {
			if he.Key == key {
				return he
			}
			he = he.next
		}

		if !dt.IsRehashing() {
			break
		}
	}

	return nil
}

// GetSomeKeys samples the dictionary to return a few keys from random locations
func (dt *Dict) GetSomeKeys(count int64) []*Entry {
	if size := dt.Used(); count > size {
		count = size
	}

	if dt.IsRehashing() {
		dt.doRehashing(int(count))
	}

	tables := 1
	if dt.IsRehashing() {
		tables = 2
	}
	maxsizemask := dt.getMaxSizeMask()
	var emptyLen, n int64
	maxSteps := count * 10

	randIdx := rand.Int63() & maxsizemask
	var keys []*Entry
	for n < count && maxSteps > 0 {

		for j := 0; j < tables; j++ {

			// there are no populated buckets
			if tables == 2 && j == 0 && randIdx < dt.rehashIndex {
				if randIdx >= dt.hts[1].size {
					randIdx = dt.rehashIndex
				} else {
					continue
				}
			}

			if randIdx >= dt.hts[j].size {
				continue
			}

			he := dt.hts[j].table[randIdx]
			if he == nil {
				// Count contiguous empty buckets
				emptyLen++
				if emptyLen >= 5 && emptyLen > count {
					// jump to other locations
					randIdx = rand.Int63() & maxsizemask
					emptyLen = 0
				}
			} else {
				emptyLen = 0

				for he != nil {
					keys = append(keys, he)
					he = he.next
					n++
					if n == count {
						return keys
					}
				}
			}
		}

		randIdx = (randIdx + 1) & maxsizemask
		maxSteps--
	}

	return keys
}

func (dt *Dict) GetRandomKey() *Entry {
	if dt.Used() == 0 {
		return nil
	}

	if dt.IsRehashing() {
		dt.doRehashing(1)
	}

	var entry *Entry
	if dt.IsRehashing() {
		for entry == nil {
			s0 := dt.hts[0].size
			s1 := dt.hts[1].size

			// FIXME: lack of Randomness
			randIdx := dt.rehashIndex + rand.Int63()%(s1+s0-dt.rehashIndex)
			if randIdx >= s0 {
				entry = dt.hts[1].table[randIdx-s0]
			} else {
				entry = dt.hts[0].table[randIdx]
			}
		}
	} else {
		for entry == nil {
			randIdx := rand.Int63() & dt.hts[0].sizemask
			entry = dt.hts[0].table[randIdx]
		}
	}

	len := 0
	origin := entry
	for entry != nil {
		len++
		entry = entry.next
	}
	entry = origin
	n := rand.Int() % len
	for n > 0 {
		entry = entry.next
		n--
	}
	return entry
}

func (dt *Dict) Delete(key string) interface{} {
	if dt.Used() == 0 {
		return nil
	}

	if dt.IsRehashing() {
		dt.doRehashing(1)
	}

	for _, ht := range dt.hts {
		idx := dt.hash(key) & ht.sizemask
		he := ht.table[idx]

		var prevHe *Entry
		for he != nil {
			if he.Key == key {
				if prevHe != nil {
					prevHe.next = he.next
				} else {
					ht.table[idx] = he.next
				}
				ht.used--
				return he.Value
			}
			prevHe = he
			he = he.next
		}

		if !dt.IsRehashing() {
			break
		}
	}

	return nil
}

func (dt *Dict) Used() int64 {
	return dt.hts[0].used + dt.hts[1].used
}

func (dt *Dict) Size() int64 {
	return dt.hts[0].size + dt.hts[1].size
}

func (dt *Dict) keyIndexToPopulated(key string) (int64, *Entry) {
	var idx int64

	if !dt.expandIfNeed() {
		return -1, nil
	}

	for _, ht := range dt.hts {
		idx = dt.hash(key) & ht.sizemask
		he := ht.table[idx]

		for he != nil {
			if he.Key == key {
				return idx, he
			}
			he = he.next
		}

		if !dt.IsRehashing() {
			break
		}
	}

	return idx, nil
}

func (dt *Dict) expandIfNeed() bool {
	if dt.IsRehashing() {
		return true // expand succeed
	}

	if dt.hts[0].size == 0 {
		return dt.expandDict(HtInitialSize)
	}

	if dt.hts[0].used >= dt.hts[0].size {
		return dt.expandDict(dt.hts[0].used * 2)
	}

	return true
}

func (dt *Dict) expandDict(size int64) bool {
	if dt.IsRehashing() || dt.hts[0].used > size {
		return false
	}

	realsize := dt.nextPower(size) // capability is a power of two
	nht := newHashTable(realsize)

	dt.hts[1] = nht
	dt.rehashIndex = 0 // start rehashing
	return true
}

// Performs N steps of incremental rehashing
func (dt *Dict) doRehashing(n int) bool {
	if !dt.IsRehashing() {
		return false // not need to rehashing
	}

	emptyVisits := n * 10 // Max number of empty buckets to visit

	for n > 0 && dt.hts[0].used != 0 {

		for dt.hts[0].table[dt.rehashIndex] == nil {
			dt.rehashIndex++
			emptyVisits--
			if emptyVisits == 0 {
				return true
			}
		}

		// Move all the keys in this bucket from the old to the new HT
		de := dt.hts[0].table[dt.rehashIndex]
		for de != nil {
			nextde := de.next

			idx := dt.hash(de.Key) & dt.hts[1].sizemask
			de.next = dt.hts[1].table[idx]
			dt.hts[1].table[idx] = de
			dt.hts[0].used--
			dt.hts[1].used++

			de = nextde
		}

		dt.hts[0].table[dt.rehashIndex] = nil
		dt.rehashIndex++
		n--
	}

	// Check if we already rehashed the whole table
	if dt.hts[0].used == 0 {
		dt.hts[0] = dt.hts[1]
		dt.hts[1] = newHashTable(HtInitialSize)
		dt.rehashIndex = -1
		return false
	}
	// More to rehash...
	return true
}

// RehashingMillseconds Rehashing for an amount of time between ms milliseconds and ms+1 milliseconds
func (dt *Dict) RehashingMillseconds(ms int64) int {
	start := time.Now().UnixNano() / int64(time.Millisecond)
	rehashes := 0
	for dt.doRehashing(100) {
		rehashes += 100
		if time.Now().UnixNano()/int64(time.Millisecond)-start > ms {
			break
		}
	}
	return rehashes
}

func (dt *Dict) nextPower(size int64) int64 {
	i := HtInitialSize
	for {
		if i >= size {
			return i
		}
		i *= 2
	}
}

func (dt *Dict) IsRehashing() bool {
	return dt.rehashIndex != -1
}

func (dt *Dict) hash(k string) int64 {
	h := fnv.New64()
	h.Write([]byte(k))
	return int64(h.Sum64())
}

func (dt *Dict) getMaxSizeMask() int64 {
	maxsizemask := dt.hts[0].sizemask
	if dt.IsRehashing() {
		if ms := dt.hts[1].sizemask; ms > maxsizemask {
			maxsizemask = ms
		}
	}
	return maxsizemask
}
