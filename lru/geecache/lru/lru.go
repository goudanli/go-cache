package lru

import "container/list"

type Cache struct {
	maxBytes  int64 //缓存最大上限
	nbytes    int64 //当前已使用的内存
	link      *list.List
	cache     map[string]*list.Element      //*list.Element节点指针
	OnEvicted func(key string, value Value) //回调函数
}

//双向链表的节点，节点的value是实现了Value接口的任意类型
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		link:      list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.link.MoveToFront(ele)
		kv, ok := ele.Value.(*entry) //类型断言
		if ok {
			return kv.value, true
		} else {
			return nil, false
		}
	}
	return
}

func (c *Cache) RemoveOldest() {
	ele := c.link.Back()
	if ele != nil {
		c.link.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.link.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.link.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.link.Len()
}
