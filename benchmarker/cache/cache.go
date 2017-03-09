package cache

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Cache struct {
	l    *sync.RWMutex
	Data map[*http.Request]*CacheData
}

type CacheData struct {
	LastModified string
	Etag         string
	Expires      time.Time
	Resp         *http.Response
}

var cacheRegex = regexp.MustCompile(`([a-zA-Z][a-zA-Z_-]*)\s*(?:=(?:"([^"]*)"|([^ \t",;]*)))?`)

func (c *Cache) Get(key *http.Request) (*CacheData, bool) {
	defer c.l.RUnlock()
	c.l.RLock()
	val, ok := c.Data[key]

	if !ok {
		return nil, false
	}

	if !val.IsValid() {
		delete(c.Data, key)
		return nil, false
	}

	return val, ok
}

func (c *Cache) Set(key *http.Request, val *CacheData) {
	defer c.l.Unlock()
	c.l.Lock()
	if c.Data == nil {
		c.Data = make(map[*http.Request]*CacheData)
	}
	c.Data[key] = val
}

func (c *Cache) Clear() {
	defer c.l.Unlock()
	c.l.Lock()
	c.Data = nil
}

func (c *CacheData) IsValid() bool {
	return time.Now().Before(c.Expires)
}

func NewCache() *Cache {
	return &Cache{
		l:    new(sync.RWMutex),
		Data: make(map[*http.Request]*CacheData),
	}
}

func NewHTTPCache(res *http.Response) (*CacheData, error) {

	header := res.Header.Get("Cache-Control")
	if len(header) == 0 {
		return nil, errors.New("Cache-Control Header Not Found")
	}

	cache := make(map[string]string)

	for _, match := range cacheRegex.Copy().FindAllString(header, -1) {
		if strings.EqualFold(match, "no-store") {
			return nil, errors.New("no-store detected")
		}
		var key, value string
		key = match
		if index := strings.Index(match, "="); index != -1 {
			key, value = match[:index], match[index+1:]
		}
		cache[key] = value
	}

	limit, ok := cache["max-age"]

	if !ok {
		return nil, errors.New("cache age not found")
	}

	t, err := strconv.Atoi(limit)

	if err != nil {
		return nil, err
	}

	return &CacheData{
		LastModified: res.Header.Get("Last-Modified"),
		Etag:         res.Header.Get("ETag"),
		Expires:      time.Now().Add(time.Duration(t) * time.Second),
		Resp:         res,
	}, nil
}
