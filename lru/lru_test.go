
package lru

import (
	"container/list"
	"testing"
)

func newTestCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cacheMap: make(map[int]*list.Element),
		order:    list.New(),
	}
}

func assertGet(
	t *testing.T,
	cache *LRUCache,
	key int,
	wantValue int,
	wantOK bool,
) {
	t.Helper()

	gotValue, gotOK := cache.Get(key)

	if gotValue != wantValue || gotOK != wantOK {
		t.Fatalf(
			"Get(%d) = (%d, %v), want (%d, %v)",
			key,
			gotValue,
			gotOK,
			wantValue,
			wantOK,
		)
	}
}

// --------------------------------------------------
// Basic Put / Get
// --------------------------------------------------

func TestLRU_PutAndGet(t *testing.T) {
	cache := newTestCache(2)

	cache.Put(1, 100)

	assertGet(t, cache, 1, 100, true)
}

func TestLRU_GetMissingKey(t *testing.T) {
	cache := newTestCache(2)

	assertGet(t, cache, 1, 0, false)

	cache.Put(1, 100)

	assertGet(t, cache, 2, 0, false)
}

// --------------------------------------------------
// Capacity / Eviction
// --------------------------------------------------

func TestLRU_EvictsLeastRecentlyUsed(t *testing.T) {
	cache := newTestCache(2)

	cache.Put(1, 100)
	cache.Put(2, 200)

	// 2 = MRU
	// 1 = LRU

	cache.Put(3, 300)

	// 1 should be evicted.
	assertGet(t, cache, 1, 0, false)

	assertGet(t, cache, 2, 200, true)
	assertGet(t, cache, 3, 300, true)
}

func TestLRU_GetUpdatesRecency(t *testing.T) {
	cache := newTestCache(2)

	cache.Put(1, 100)
	cache.Put(2, 200)

	// Access 1.
	// Order should now be:
	//
	// 1 -> MRU
	// 2 -> LRU
	cache.Get(1)

	cache.Put(3, 300)

	// 2 should be evicted, not 1.
	assertGet(t, cache, 2, 0, false)
	assertGet(t, cache, 1, 100, true)
	assertGet(t, cache, 3, 300, true)
}

func TestLRU_RepeatedGetUpdatesRecency(t *testing.T) {
	cache := newTestCache(3)

	cache.Put(1, 100)
	cache.Put(2, 200)
	cache.Put(3, 300)

	cache.Get(1)
	cache.Get(2)
	cache.Get(1)

	// Recency:
	//
	// 1 -> MRU
	// 2
	// 3 -> LRU

	cache.Put(4, 400)

	assertGet(t, cache, 3, 0, false)

	assertGet(t, cache, 1, 100, true)
	assertGet(t, cache, 2, 200, true)
	assertGet(t, cache, 4, 400, true)
}

// --------------------------------------------------
// Updating existing keys
// --------------------------------------------------

func TestLRU_UpdateExistingKey(t *testing.T) {
	cache := newTestCache(2)

	cache.Put(1, 100)
	cache.Put(2, 200)

	cache.Put(1, 999)

	assertGet(t, cache, 1, 999, true)
	assertGet(t, cache, 2, 200, true)

	// Updating an existing key should not increase
	// the number of entries.
	if cache.order.Len() != 2 {
		t.Fatalf(
			"cache size after update = %d, want 2",
			cache.order.Len(),
		)
	}

	if len(cache.cacheMap) != 2 {
		t.Fatalf(
			"map size after update = %d, want 2",
			len(cache.cacheMap),
		)
	}
}

func TestLRU_UpdateMakesKeyMostRecentlyUsed(t *testing.T) {
	cache := newTestCache(2)

	cache.Put(1, 100)
	cache.Put(2, 200)

	// Updating 1 should make it MRU.
	cache.Put(1, 999)

	// Adding 3 should evict 2.
	cache.Put(3, 300)

	assertGet(t, cache, 1, 999, true)
	assertGet(t, cache, 2, 0, false)
	assertGet(t, cache, 3, 300, true)
}

// --------------------------------------------------
// Zero values
// --------------------------------------------------

func TestLRU_ZeroValue(t *testing.T) {
	cache := newTestCache(2)

	cache.Put(1, 0)

	assertGet(t, cache, 1, 0, true)
}

func TestLRU_ZeroAndMissingKey(t *testing.T) {
	cache := newTestCache(2)

	cache.Put(1, 0)

	value, ok := cache.Get(1)

	if !ok {
		t.Fatal("expected key 1 to exist")
	}

	if value != 0 {
		t.Fatalf("Get(1) = %d, want 0", value)
	}

	// Important: missing key also returns 0,
	// so the boolean must be checked.
	value, ok = cache.Get(999)

	if ok {
		t.Fatal("expected missing key to return false")
	}

	if value != 0 {
		t.Fatalf("missing key returned value %d, want 0", value)
	}
}

// --------------------------------------------------
// Negative keys / values
// --------------------------------------------------

func TestLRU_NegativeKeyAndValue(t *testing.T) {
	cache := newTestCache(2)

	cache.Put(-1, -100)

	assertGet(t, cache, -1, -100, true)
}

func TestLRU_MultipleNegativeKeys(t *testing.T) {
	cache := newTestCache(2)

	cache.Put(-1, -100)
	cache.Put(-2, -200)

	// Access -1, making it MRU.
	assertGet(t, cache, -1, -100, true)

	// -1 = MRU
	// -2 = LRU

	cache.Put(-3, -300)

	// -2 should be evicted.
	assertGet(t, cache, -2, 0, false)

	assertGet(t, cache, -1, -100, true)
	assertGet(t, cache, -3, -300, true)
}

// --------------------------------------------------
// Capacity = 1
// --------------------------------------------------

func TestLRU_CapacityOne(t *testing.T) {
	cache := newTestCache(1)

	cache.Put(1, 100)

	assertGet(t, cache, 1, 100, true)

	cache.Put(2, 200)

	assertGet(t, cache, 1, 0, false)
	assertGet(t, cache, 2, 200, true)
}

func TestLRU_CapacityOneUpdate(t *testing.T) {
	cache := newTestCache(1)

	cache.Put(1, 100)
	cache.Put(1, 999)

	assertGet(t, cache, 1, 999, true)

	if cache.order.Len() != 1 {
		t.Fatalf(
			"cache size = %d, want 1",
			cache.order.Len(),
		)
	}
}

// --------------------------------------------------
// Capacity = 0
// --------------------------------------------------

func TestLRU_CapacityZero(t *testing.T) {
	cache := newTestCache(0)

	cache.Put(1, 100)

	assertGet(t, cache, 1, 0, false)

	if cache.order.Len() != 0 {
		t.Fatalf(
			"cache list size = %d, want 0",
			cache.order.Len(),
		)
	}

	if len(cache.cacheMap) != 0 {
		t.Fatalf(
			"cache map size = %d, want 0",
			len(cache.cacheMap),
		)
	}
}

// --------------------------------------------------
// Multiple evictions
// --------------------------------------------------

func TestLRU_MultipleEvictions(t *testing.T) {
	cache := newTestCache(2)

	cache.Put(1, 100)
	cache.Put(2, 200)
	cache.Put(3, 300)
	cache.Put(4, 400)
	cache.Put(5, 500)

	assertGet(t, cache, 1, 0, false)
	assertGet(t, cache, 2, 0, false)
	assertGet(t, cache, 3, 0, false)

	assertGet(t, cache, 4, 400, true)
	assertGet(t, cache, 5, 500, true)
}

// --------------------------------------------------
// Complex access pattern
// --------------------------------------------------

func TestLRU_ComplexAccessPattern(t *testing.T) {
	cache := newTestCache(3)

	cache.Put(1, 100)
	cache.Put(2, 200)
	cache.Put(3, 300)

	// Initial:
	//
	// 3 -> MRU
	// 2
	// 1 -> LRU

	cache.Get(1)

	// Now:
	//
	// 1 -> MRU
	// 3
	// 2 -> LRU

	cache.Put(4, 400)

	// 2 evicted.
	assertGet(t, cache, 2, 0, false)

	// Access 3.
	cache.Get(3)

	// Now:
	//
	// 3 -> MRU
	// 4
	// 1 -> LRU

	cache.Put(5, 500)

	// 1 evicted.
	assertGet(t, cache, 1, 0, false)

	assertGet(t, cache, 3, 300, true)
	assertGet(t, cache, 4, 400, true)
	assertGet(t, cache, 5, 500, true)
}

// --------------------------------------------------
// Internal consistency
// --------------------------------------------------

func TestLRU_MapAndListStayInSync(t *testing.T) {
	cache := newTestCache(3)

	cache.Put(1, 100)
	cache.Put(2, 200)
	cache.Put(3, 300)

	cache.Get(1)
	cache.Put(4, 400)
	cache.Put(2, 222)

	if cache.order.Len() != len(cache.cacheMap) {
		t.Fatalf(
			"list size = %d, map size = %d",
			cache.order.Len(),
			len(cache.cacheMap),
		)
	}

	if cache.order.Len() > cache.capacity {
		t.Fatalf(
			"cache size = %d exceeds capacity %d",
			cache.order.Len(),
			cache.capacity,
		)
	}

	// Every map entry should point to an element
	// that actually exists in the list.
	for key, element := range cache.cacheMap {
		if element == nil {
			t.Fatalf("map entry for key %d is nil", key)
		}

		found := false

		for e := cache.order.Front(); e != nil; e = e.Next() {
			if e == element {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf(
				"map entry for key %d points to element not present in list",
				key,
			)
		}
	}
}

// --------------------------------------------------
// Exact order verification
// --------------------------------------------------

func TestLRU_Order(t *testing.T) {
	cache := newTestCache(3)

	cache.Put(1, 100)
	cache.Put(2, 200)
	cache.Put(3, 300)

	// Expected:
	// 3 -> 2 -> 1
	assertOrder(t, cache, []int{3, 2, 1})

	cache.Get(1)

	// Expected:
	// 1 -> 3 -> 2
	assertOrder(t, cache, []int{1, 3, 2})

	cache.Get(2)

	// Expected:
	// 2 -> 1 -> 3
	assertOrder(t, cache, []int{2, 1, 3})

	cache.Put(4, 400)

	// 3 should be evicted.
	//
	// Expected:
	// 4 -> 2 -> 1
	assertOrder(t, cache, []int{4, 2, 1})
}

func assertOrder(t *testing.T, cache *LRUCache, want []int) {
	t.Helper()

	var got []int

	for e := cache.order.Front(); e != nil; e = e.Next() {
		got = append(got, e.Value.(entry).key)
	}

	if len(got) != len(want) {
		t.Fatalf(
			"order = %v, want %v",
			got,
			want,
		)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf(
				"order = %v, want %v",
				got,
				want,
			)
		}
	}
}
