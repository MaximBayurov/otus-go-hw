package hw04lrucache

type Key string

// cacheItem представляет элемент кэша.
type cacheItem struct {
	key   Key
	value any
}

type Cache interface {
	// Set добавляет/обновляет значение в кэше по ключу.
	// Возвращаемое значение - флаг, присутствовал ли элемент в кэше.
	Set(key Key, value interface{}) bool
	// Get получение элемента по ключу
	Get(key Key) (interface{}, bool)
	// Clear очистка кэша
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func (c *lruCache) Set(key Key, value any) bool {
	val, isExists := c.items[key]
	if isExists {
		val.Value.(*cacheItem).value = value
		c.queue.MoveToFront(val)

		return true
	}
	if c.queue.Len() == c.capacity {
		oldest := c.queue.Back()
		oldestItem := oldest.Value.(*cacheItem)
		delete(c.items, oldestItem.key)
		c.queue.Remove(oldest)
	}
	c.items[key] = c.queue.PushFront(&cacheItem{
		key:   key,
		value: value,
	})
	return false
}

func (c *lruCache) Get(key Key) (any, bool) {
	val, isExists := c.items[key]
	if isExists {
		c.queue.MoveToFront(val)
		return val.Value.(*cacheItem).value, true
	}
	return nil, false
}

func (c *lruCache) Clear() {
	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
