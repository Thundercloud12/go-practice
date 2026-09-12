package lru
import(
	"fmt"
	"container/list"
)
func Abs(x int64) int64 {
    if x < 0 {
        return -x
    }
    return x
}   

type entry struct {
    key   int
    value int
}

type LRUCache struct{
	capacity int
	cacheMap map[int]*list.Element
	order *list.List
}

func (c *LRUCache) Get(key int)(int,bool){
	m,ok:=c.cacheMap[key]
	if !ok {
		return 0,false
	}
	c.order.MoveToFront(m)

	return m.Value.(entry).value,true
}


func (c *LRUCache) Put(key int, value int) {

    if e, ok := c.cacheMap[key]; ok {
        e.Value = entry{
            key:   key,
            value: value,
        }

        c.order.MoveToFront(e)
        return
    }

    e := c.order.PushFront(entry{
        key:   key,
        value: value,
    })

    c.cacheMap[key] = e

 
    if c.order.Len() > c.capacity {
        last := c.order.Back()
        keyToRemove := last.Value.(entry).key

        delete(c.cacheMap, keyToRemove)
        c.order.Remove(last)
    }
}

func lru() {
	cache := LRUCache{
		capacity: 2,
		cacheMap: make(map[int]*list.Element),
		order:    list.New(),
	}

	cache.Put(1, 100)
	cache.Put(2, 200)

	value, ok := cache.Get(1)
	fmt.Println(value, ok) // 100 true

	cache.Put(3, 300)

	value, ok = cache.Get(2)
	fmt.Println(value, ok) // 0 false

	value, ok = cache.Get(1)
	fmt.Println(value, ok) // 100 true

	value, ok = cache.Get(3)
	fmt.Println(value, ok) // 300 true
}