package cache

import (
	"bytes"
	"encoding/gob"
	"testing"
)

type testCacheVal string

func (cacheVal testCacheVal) Size() int {
	return len(cacheVal)
}

func TestLRUCache_InitCapacity(t *testing.T) {
	lru := NewLRUCache(100)

	// Test if the cache capacity allocation works
	if lru.capacity != 100 {
		t.Errorf("Size expected %d, but got %d", 100, lru.capacity)
	}
}

func TestLRUCache_Get(t *testing.T) {
	lru := NewLRUCache(100)

	lru.Set("key", testCacheVal("value"))

	val, ok := lru.Get("key")
	if !ok || val != testCacheVal("value") {
		t.Error("Get failed")
	}

	_, ok = lru.Get("missing")
	if ok {
		t.Error("Get returned ok for missing key")
	}
}

func TestLRUCache_Set(t *testing.T) {
	lru := NewLRUCache(100)

	lru.Set("key", testCacheVal("value"))

	val, ok := lru.Get("key")
	if !ok || val != testCacheVal("value") {
		t.Error("Failed to set and get key")
	}
}

func TestLRUCache_Delete(t *testing.T) {
	lru := NewLRUCache(100)

	lru.Set("key", testCacheVal("value"))

	ok := lru.Delete("key")
	if !ok {
		t.Error("Failed to delete existing key")
	}

	ok = lru.Delete("key")
	if ok {
		t.Error("Delete returned true for missing key")
	}
}

func TestLRUCache_SetIfAbsent(t *testing.T) {

	lru := NewLRUCache(100)

	// Set new key
	lru.SetIfAbsent("newkey", testCacheVal("newValue"))

	if val, ok := lru.Get("newkey"); !ok || val != testCacheVal("newValue") {
		t.Error("Failed to set new key")
	}

	// Update existing key
	lru.Set("existingkey", testCacheVal("oldValue"))

	lru.SetIfAbsent("existingkey", testCacheVal("newValue"))

	if val, ok := lru.Get("existingkey"); !ok || val != testCacheVal("oldValue") {
		t.Error("Failed to update existing key")
	}

}

func TestLRUCache_Clear(t *testing.T) {

	lru := NewLRUCache(200)

	lru.Set("key1", testCacheVal("value1"))
	lru.Set("key2", testCacheVal("value2"))

	lru.Clear()

	if lru.list.Len() != 0 {
		t.Error("Failed to clear cache")
	}

	if _, ok := lru.Get("key1"); ok {
		t.Error("Key 1 still exists after clear")
	}

	if _, ok := lru.Get("key2"); ok {
		t.Error("Key 2 still exists after clear")
	}

}

func TestLRUCache_CheckCapacity(t *testing.T) {

	lru := NewLRUCache(200)

	lru.Set("key1", testCacheVal("value1"))
	lru.Set("key2", testCacheVal("value2"))

	// Over capacity
	lru.Set("key3", testCacheVal("value3"))

	capacity := testCacheVal("value1").Size() + testCacheVal("value2").Size()
	lru.SetCapacity(uint64(capacity))

	if lru.list.Len() != 2 {
		t.Errorf("Failed to evict when over capacity, expected %d, but got %d", 2, lru.list.Len())
	}

	if _, ok := lru.Get("key1"); ok {
		t.Error("Key 1 was not evicted")
	}

}

func TestLRUCache_SetCapacity(t *testing.T) {

	lru := NewLRUCache(200)

	lru.Set("key1", testCacheVal("value1"))
	lru.Set("key2", testCacheVal("value2"))

	capacity := testCacheVal("value1").Size()
	lru.SetCapacity(uint64(capacity))

	if lru.list.Len() != 1 {
		t.Errorf("Failed to set capacity, expected %d, but got %d", 1, lru.list.Len())
	}

}

func TestLRUCache_Stats(t *testing.T) {

	lru := NewLRUCache(100)

	lru.Set("key", testCacheVal("value"))

	length, _, capacity, _ := lru.Stats()

	if length != 1 {
		t.Error("Stats returned incorrect length")
	}

	if capacity != 100 {
		t.Error("Stats returned incorrect capacity")
	}

}

func TestLRUCache_Keys(t *testing.T) {

	v1 := testCacheVal("value1")
	v2 := testCacheVal("value2")
	lru := NewLRUCache(uint64(v1.Size()) + uint64(v2.Size()))

	lru.Set("key1", v1)
	lru.Set("key2", v2)

	keys := lru.Keys()

	if len(keys) != 2 || keys[0] != "key2" || keys[1] != "key1" {
		t.Errorf("Keys returned incorrect results, expected %v, but got %v", []string{"key2", "key1"}, keys)
	}

}

func TestLRUCache_SaveAndLoad(t *testing.T) {

	v1 := testCacheVal("value1")
	v2 := testCacheVal("value2")
	lru := NewLRUCache(uint64(v1.Size()) + uint64(v2.Size()))

	lru.Set("key1", v1)
	lru.Set("key2", v2)

	var testObj testCacheVal
	gob.Register(testObj)

	// Save to buffer
	var buf bytes.Buffer
	if err := lru.SaveItems(&buf); err != nil {
		t.Fatal(err)
	}

	// Load into new cache
	lru2 := NewLRUCache(uint64(v1.Size()) + uint64(v2.Size()))
	if err := lru2.LoadItems(&buf); err != nil {
		t.Fatal(err)
	}

	if lru2.list.Len() != 2 {
		t.Error("Failed to load items")
	}

	if val, ok := lru2.Get("key1"); !ok || val != v1 {
		t.Error("Loaded item does not match")
	}

}

func TestLRUCache_Items(t *testing.T) {

	v1 := testCacheVal("value1")
	v2 := testCacheVal("value2")
	lru := NewLRUCache(uint64(v1.Size()) + uint64(v2.Size()))

	lru.Set("key1", v1)
	lru.Set("key2", v2)

	var testObj testCacheVal
	gob.Register(testObj)

	items := lru.Items()

	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	if items[0].Key != "key2" || items[0].Value != v2 {
		t.Errorf("Unexpected item 0: %v", items[0])
	}

}
