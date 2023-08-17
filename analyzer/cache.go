package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	gorediscache "github.com/go-redis/cache/v9"
	goredis "github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache: key is missing")

// TechAndPath is a map with technology and a path
type TechAndPath struct {
	// name of the technology
	Technology string
	// path, that matches to the pattern of the technology
	Path string
}

type MatchingFilesValue struct {
	CommitID string
	Results  []*TechAndPath
}

type Item struct {
	Key    string
	Object interface{}
	// Expiration is the cache expiration time.
	Expiration time.Duration
}

type AnalyzerCache struct {
	client     *goredis.Client
	cache      *gorediscache.Cache
	expiration time.Duration
}

func NewAnalyzerCache(client *goredis.Client, expiration time.Duration) *AnalyzerCache {
	return &AnalyzerCache{
		client:     client,
		cache:      gorediscache.New(&gorediscache.Options{Redis: client}),
		expiration: expiration,
	}
}

func matchingFilesKey(repoURL string) string {
	return fmt.Sprintf("matchingfiles|%s", repoURL)
}

func (c *AnalyzerCache) GetMatchingFiles(repoURL string) (*MatchingFilesValue, error) {
	item := &MatchingFilesValue{}
	return item, c.GetItem(matchingFilesKey(repoURL), item)
}

func (c *AnalyzerCache) SetMatchingFiles(repoURL, commitId string, res []*TechAndPath) error {
	item := &MatchingFilesValue{commitId, res}
	return c.SetItem(matchingFilesKey(repoURL), item, c.expiration, false)
}

func (c *AnalyzerCache) SetItem(key string, item interface{}, expiration time.Duration, delete bool) error {
	if delete {
		return c.Delete(key)
	} else {
		if item == nil {
			return fmt.Errorf("cannot set item to nil for key %s", key)
		}
		return c.Set(&Item{Object: item, Key: key, Expiration: expiration})
	}
}

func (c *AnalyzerCache) GetItem(key string, item interface{}) error {
	if item == nil {
		return fmt.Errorf("cannot get item into a nil for key %s", key)
	}
	return c.Get(key, item)
}

func (c *AnalyzerCache) Set(item *Item) error {
	expiration := item.Expiration
	if expiration == 0 {
		expiration = c.expiration
	}

	val, err := c.marshal(item.Object)
	if err != nil {
		return err
	}

	return c.cache.Set(&gorediscache.Item{
		Key:   item.Key,
		Value: val,
		TTL:   expiration,
	})
}

func (c *AnalyzerCache) Get(key string, obj interface{}) error {
	var data []byte
	err := c.cache.Get(context.TODO(), key, &data)
	if err == ErrCacheMiss {
		err = ErrCacheMiss
	}
	if err != nil {
		return err
	}
	return c.unmarshal(data, obj)
}

func (c *AnalyzerCache) Delete(key string) error {
	return c.cache.Delete(context.TODO(), key)
}

func (c *AnalyzerCache) marshal(obj interface{}) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	var w io.Writer = buf
	encoder := json.NewEncoder(w)

	if err := encoder.Encode(obj); err != nil {
		return nil, err
	}
	if flusher, ok := w.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (c *AnalyzerCache) unmarshal(data []byte, obj interface{}) error {
	buf := bytes.NewReader(data)
	var reader io.Reader = buf
	if err := json.NewDecoder(reader).Decode(obj); err != nil {
		return fmt.Errorf("failed to decode cached data: %w", err)
	}
	return nil
}
