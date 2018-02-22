package sidecar

import (
	"github.com/opentracing/opentracing-go"
	"time"
	"sync"
)

type SecurityContext struct {
	Token string
	PlainContext string
}
type AuthMap struct {
	sync.RWMutex
	values map[uint64]SecurityContext
}

type TimeMap struct {
	sync.RWMutex
	values map[uint64]time.Time
}

type Sidecar struct {
	Tracer opentracing.Tracer
	Auths  *AuthMap
	Times  *TimeMap
}

func (t *TimeMap) Get(key uint64) time.Time {
	t.RLock()
	defer t.RUnlock()
	return t.values[key]
}

func (t *TimeMap) Delete(key uint64) {
	t.Lock()
	defer t.Unlock()
	delete(t.values, key)
}

func (t *TimeMap) Put(key uint64, value time.Time) {
	t.Lock()
	defer t.Unlock()
	t.values[key] = value
}

func (t *AuthMap) Get(key uint64) SecurityContext {
	t.RLock()
	defer t.RUnlock()
	return t.values[key]
}

func (t *AuthMap) Put(key uint64, value SecurityContext) {
	t.Lock()
	defer t.Unlock()
	t.values[key] = value
}

func (t *AuthMap) Delete(key uint64) {
	t.Lock()
	defer t.Unlock()
	delete(t.values, key)
}

func NewTimeMap() *TimeMap {
	return &TimeMap{
		values: make(map[uint64]time.Time),
	}
}

func NewAuthsMap() *AuthMap {
	return &AuthMap{
		values: make(map[uint64]SecurityContext),
	}
}