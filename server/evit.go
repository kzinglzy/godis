package server

import (
	"log"

	"github.com/kzinglzy/godis/dt"
)

type EvPoolEntry struct {
	idle int64
	key  string
}

var EvPool [EvPoolSize]EvPoolEntry

func init() {
	for i := 0; i < EvPoolSize; i++ {
		EvPool[i] = EvPoolEntry{}
	}
}

func freeMemoryIfNeed() bool {
	if godisServer.maxmemoryPolicy == MaxmemoryNoEviction {
		return false
	}
	toFree := maxMemoryToFree()
	if toFree <= 0 {
		return false
	}

	log.Printf("start free memory %d", toFree)
	var freed int64
	for freed < toFree {
		bestkey := ""

		if godisServer.memPolicyLru() {
			for bestkey == "" {
				populateEvictionPool(godisServer.db.store)
				for k := EvPoolSize - 1; k >= 0; k-- {
					key := EvPool[k].key
					if key == "" {
						continue
					}
					EvPool[k].key = ""
					EvPool[k].idle = 0
					bestkey = key
					break
				}
			}
		} else if godisServer.memPolicyRandom() {
			if e := godisServer.db.store.RandomEntry(); e != nil {
				bestkey = e.Key
			}
		}

		preMem := usedmemory()
		if bestkey != "" {
			log.Printf("evict key %s to free memory", bestkey)
			godisServer.db.deleteKey(bestkey)
		}
		freed += preMem - usedmemory()
	}
	return true
}

func maxMemoryToFree() int64 {
	used := usedmemory()
	if godisServer.maxmemory == 0 || used <= godisServer.maxmemory {
		return 0
	}
	return used - godisServer.maxmemory
}

func populateEvictionPool(sampledict *dt.Dict) {
	entries := sampledict.SomeEntries(MaxmemorySamples)
	for _, e := range entries {
		o := e.Value.(*dt.Object)
		idle := mstime() - o.Lru

		// find the first empty bucket or the first populated bucket
		// that has an idle time greater than our idle time.
		var k int
		var found, replaced bool
		for i, e := range EvPool {
			if e.key == "" {
				k = i
				found = true
				break
			}

			if e.idle > idle {
				k = i
				found = true
				replaced = true
				break
			}
		}

		if !found {
			continue
		} else if replaced {
			if EvPool[EvPoolSize-1].key == "" {
				// insert at k shifting all the elements from k to end to the right
				shiftEvPoolToRight(k)
			} else {
				// shift all elements on the left of k (included) to the
				// left, so we discard the element with smaller idle time
				shiftEvPoolToLeft(k)
			}
		}

		EvPool[k].idle = idle
		EvPool[k].key = e.Key
	}
}

func shiftEvPoolToRight(k int) {
	i := k + 1
	prev := EvPool[k]
	for i < EvPoolSize {
		cur := EvPool[i]
		EvPool[i] = prev
		prev = cur

		i++
	}
}

func shiftEvPoolToLeft(k int) {
	i := k - 1
	prev := EvPool[k]
	for i >= 0 {
		cur := EvPool[i]
		EvPool[i] = prev
		prev = cur

		i--
	}
}
