package streets

import (
	"container/list"
	"sync"
)

type ThreadSafeDeque[T comparable] struct {
	deque *list.List
	lock  *sync.RWMutex
}

func NewThreadSafeDeque[T comparable]() *ThreadSafeDeque[T] {
	return &ThreadSafeDeque[T]{
		deque: list.New(),
		lock:  &sync.RWMutex{},
	}
}

func (d *ThreadSafeDeque[T]) PushFront(value T) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.deque.PushFront(value)
}

func (d *ThreadSafeDeque[T]) PushBack(value T) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.deque.PushBack(value)
}

func (d *ThreadSafeDeque[T]) PopFront() T {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.deque.Len() == 0 {
		return *new(T)
	}

	element := d.deque.Front()
	d.deque.Remove(element)

	return element.Value.(T)
}

func (d *ThreadSafeDeque[T]) PopBack() T {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.deque.Len() == 0 {
		return *new(T)
	}

	element := d.deque.Back()
	d.deque.Remove(element)

	return element.Value.(T)
}

func (d *ThreadSafeDeque[T]) Len() int {
	d.lock.RLock()
	defer d.lock.RUnlock()

	return d.deque.Len()
}

func (d *ThreadSafeDeque[T]) At(i int) T {
	d.lock.RLock()
	defer d.lock.RUnlock()

	element := d.deque.Front()
	for j := 0; j < i; j++ {
		element = element.Next()
	}
	return element.Value.(T)
}

func (d *ThreadSafeDeque[T]) Back() T {
	d.lock.RLock()
	defer d.lock.RUnlock()

	return d.deque.Back().Value.(T)
}

func (d *ThreadSafeDeque[T]) Exists(t T) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()

	for element := d.deque.Front(); element != nil; element = element.Next() {
		if element.Value.(T) == t {
			return true
		}
	}
	return false
}
