package service

import "sync"

// 缓存获取到的数据，对于Data这个map而言，key是dbname，value为从这个db获取到的表数据
type cache struct {
	Data  map[string]*dbFields
	mutex *sync.RWMutex
}

type dbFields struct {
	Columns []string
	Rows    [][]string
}

var databaseDataCache = newCache()

var once = sync.Once{}

func newCache() *cache {
	return &cache{
		Data:  make(map[string]*dbFields),
		mutex: &sync.RWMutex{},
	}
}

func (c *cache) Columns(dbName string, columns []string) {
	c.Data[dbName].Columns = columns
}

func (c *cache) MergeSlice(dbName string, m []string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Data[dbName].Rows = append(c.Data[dbName].Rows, m)
}

func (c *cache) Range(dbName string, f func(index int, data []string) bool) {
	for index, value := range c.Data[dbName].Rows {
		c.mutex.RLock()
		if f(index, value) {
			c.mutex.RUnlock()
			continue
		} else {
			c.mutex.RUnlock()
			break
		}
	}
}

func (c *cache) Clear() {
	databaseDataCache.mutex.Lock()
	defer databaseDataCache.mutex.Unlock()
	c.Data = make(map[string]*dbFields)
}

func (c *cache) GetIndexData(dbName string, index int) []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.Data[dbName].Rows[index]
}
