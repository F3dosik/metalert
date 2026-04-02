package pool

import "sync"

// Resettable — ограничение типов, которые можно хранить в Pool.
// Любой тип T должен реализовывать метод Reset(),
// который сбрасывает состояние объекта перед возвратом в пул.
type Resettable interface {
	Reset()
}

// Pool — типобезопасная обёртка над sync.Pool для объектов типа T.
// T должен реализовывать интерфейс Resettable (иметь метод Reset()).
type Pool[T Resettable] struct {
	pool sync.Pool
}

// New создаёт и возвращает указатель на Pool[T].
// Аргумент fn — фабричная функция, которая создаёт новый объект типа T,
// когда пул пуст и требуется новый экземпляр.
func New[T Resettable](fn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return fn()
			},
		},
	}
}

// Get возвращает объект типа T из пула.
// Если пул пуст, вызывается фабричная функция, переданная в New.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put сбрасывает состояние объекта через Reset() и возвращает его в пул.
func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.pool.Put(v)
}
