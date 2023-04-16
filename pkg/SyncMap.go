package litws

import (
	"fmt"
	"sort"
	"sync"
)

type IdWithValuesCallback[IdType comparable, T any] func(map[IdType]*T)
type SyncMapSortFn[IdType comparable, T any] func(a, b KeyValue[IdType, T], field string, desc bool) bool
type KeyValue[IdType comparable, T any] struct {
	Id    IdType
	Value *T
}

type SyncMap[IdType comparable, T any] struct {
	m             map[IdType]*T
	mux           sync.RWMutex
	eventHandlers map[string][]IdWithValuesCallback[IdType, T] // internal for lws

	SortFunc     SyncMapSortFn[IdType, T]
	ValueByField func(string, *T) interface{}
}

func NewSyncMap[IdType comparable, T any]() *SyncMap[IdType, T] {
	return &SyncMap[IdType, T]{
		m:             make(map[IdType]*T),
		mux:           sync.RWMutex{},
		eventHandlers: map[string][]IdWithValuesCallback[IdType, T]{},
	}
}

func (sm *SyncMap[IdType, T]) addEventListener(event string, f IdWithValuesCallback[IdType, T]) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	sm.eventHandlers[event] = append(sm.eventHandlers[event], f)
}

func (sm *SyncMap[IdType, T]) removeEventListener(event string, f IdWithValuesCallback[IdType, T]) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	for i, f2 := range sm.eventHandlers[event] {
		if fmt.Sprintf("%p", f) == fmt.Sprintf("%p", f2) {
			sm.eventHandlers[event] = append(sm.eventHandlers[event][:i], sm.eventHandlers[event][i+1:]...)
			return
		}
	}
}

func (sm *SyncMap[IdType, T]) Get(id IdType) (*T, bool) {
	sm.mux.RLock()
	defer sm.mux.RUnlock()
	val, ok := sm.m[id]
	return val, ok
}

func (sm *SyncMap[IdType, T]) Set(id IdType, val *T) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	sm.m[id] = val
	for _, f := range sm.eventHandlers["set"] {
		f(map[IdType]*T{id: val})
	}
}

func (sm *SyncMap[IdType, T]) Delete(id IdType) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	delete(sm.m, id)
	for _, f := range sm.eventHandlers["delete"] {
		f(map[IdType]*T{id: nil})
	}
}

func (sm *SyncMap[IdType, T]) Updated(id IdType) {
	sm.mux.RLock()
	defer sm.mux.RUnlock()
	val, ok := sm.m[id]
	if ok {
		for _, f := range sm.eventHandlers["updated"] {
			f(map[IdType]*T{id: val})
		}
	}
}

func (sm *SyncMap[IdType, T]) MultipleUpdates(ids []IdType) {
	sm.mux.RLock()
	defer sm.mux.RUnlock()
	var vals = make(map[IdType]*T)
	for _, id := range ids {
		val, ok := sm.m[id]
		if ok {
			vals[id] = val
		}
	}
	for _, f := range sm.eventHandlers["updated"] {
		f(vals)
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

func (sm *SyncMap[IdType, T]) Values() []*T {
	sm.mux.RLock()
	defer sm.mux.RUnlock()
	values := make([]*T, 0, len(sm.m))
	for _, v := range sm.m {
		values = append(values, v)
	}
	return values
}

func (sm *SyncMap[IdType, T]) GetSortedList(orderBy string, orderDesc bool) []KeyValue[IdType, T] {
	sm.mux.RLock()
	defer sm.mux.RUnlock()
	if sm.SortFunc == nil {
		panic("SortFunc is nil")
	}
	var list = make([]KeyValue[IdType, T], 0, len(sm.m))
	for k, v := range sm.m {
		list = append(list, KeyValue[IdType, T]{k, v})
	}
	sort.Slice(list, func(i, j int) bool {
		return sm.SortFunc(list[i], list[j], orderBy, orderDesc)
	})
	return list
}
