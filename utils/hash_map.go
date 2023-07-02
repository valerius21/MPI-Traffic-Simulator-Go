package utils

import "sync"

type HashMap[T comparable, U any] struct {
	m map[T]U
	sync.Mutex
}

func NewMap[T comparable, U any]() HashMap[T, U] {
	m := make(map[T]U)
	return HashMap[T, U]{
		m: m,
	}
}

func (m *HashMap[T, U]) ToList() []U {
	m.Lock()
	defer m.Unlock()

	list := make([]U, 0, len(m.m))

	for _, value := range m.m {
		list = append(list, value)
	}

	return list
}

func (m *HashMap[T, U]) Set(key T, value U) {
	m.Lock()
	defer m.Unlock()

	m.m[key] = value
}

func (m *HashMap[T, U]) Get(key T) (U, bool) {
	m.Lock()
	defer m.Unlock()

	value, ok := m.m[key]
	return value, ok
}

func (m *HashMap[T, U]) GetOrInsert(key T, value U) (val U, loaded bool) {
	m.Lock()
	defer m.Unlock()

	if m.Has(key) {
		val, _ := m.Get(key)
		return val, true
	}

	m.Set(key, value)

	return val, false
}

func (m *HashMap[T, U]) Has(key T) bool {
	m.Lock()
	defer m.Unlock()

	_, ok := m.m[key]
	return ok
}

func (m *HashMap[T, U]) Del(key T) {
	m.Lock()
	defer m.Unlock()

	delete(m.m, key)
}

func (m *HashMap[T, U]) Len() int {
	m.Lock()
	defer m.Unlock()

	return len(m.m)
}
