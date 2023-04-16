package litws

import (
	"fmt"
	"sync"
)

type SyncMap[IdType comparable, T any] struct {
	m             map[IdType]T
	mux           sync.RWMutex
	eventHandlers map[string][]func(IdType, T) // internal for lws
}

func NewSyncMap[IdType comparable, T any]() *SyncMap[IdType, T] {
	return &SyncMap[IdType, T]{
		m:             make(map[IdType]T),
		mux:           sync.RWMutex{},
		eventHandlers: map[string][]func(IdType, T){},
	}
}

func (sm *SyncMap[IdType, T]) addEventListener(event string, f func(IdType, T)) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	sm.eventHandlers[event] = append(sm.eventHandlers[event], f)
}

func (sm *SyncMap[IdType, T]) removeEventListener(event string, f func(IdType, T)) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	for i, f2 := range sm.eventHandlers[event] {
		if fmt.Sprintf("%p", f) == fmt.Sprintf("%p", f2) {
			sm.eventHandlers[event] = append(sm.eventHandlers[event][:i], sm.eventHandlers[event][i+1:]...)
			return
		}
	}
}

func (sm *SyncMap[IdType, T]) Get(id IdType) (T, bool) {
	sm.mux.RLock()
	defer sm.mux.RUnlock()
	val, ok := sm.m[id]
	return val, ok
}

func (sm *SyncMap[IdType, T]) Set(id IdType, val T) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	sm.m[id] = val
	for _, f := range sm.eventHandlers["set"] {
		f(id, val)
	}
}

func (sm *SyncMap[IdType, T]) Delete(id IdType) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	delete(sm.m, id)
	for _, f := range sm.eventHandlers["delete"] {
		f(id, nil)
	}
}

func (sm *SyncMap[IdType, T]) Updated(id IdType) {
	sm.mux.RLock()
	defer sm.mux.RUnlock()
	val, ok := sm.m[id]
	if ok {
		for _, f := range sm.eventHandlers["updated"] {
			f(id, val)
		}
	}
}

func (sm *SyncMap[IdType, T]) Len() int {
	sm.mux.RLock()
	defer sm.mux.RUnlock()
	return len(sm.m)
}

func (sm *SyncMap[IdType, T]) Keys() []IdType {
	sm.mux.RLock()
	defer sm.mux.RUnlock()
	keys := make([]IdType, 0, len(sm.m))
	for k := range sm.m {
		keys = append(keys, k)
	}
	return keys
}

func (sm *SyncMap[IdType, T]) Values() []T {
	sm.mux.RLock()
	defer sm.mux.RUnlock()
	values := make([]T, 0, len(sm.m))
	for _, v := range sm.m {
		values = append(values, v)
	}
	return values
}
