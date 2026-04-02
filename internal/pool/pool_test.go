package pool

import (
	"sync"
	"testing"
)

// --- Тестовые типы ---

// mockObject — простой объект с счётчиком сбросов.
type mockObject struct {
	Value      string
	ResetCount int
}

func (m *mockObject) Reset() {
	m.Value = ""
	m.ResetCount++
}

// multiFieldObject — объект с несколькими полями для проверки полного сброса.
type multiFieldObject struct {
	Name  string
	Score int
	Tags  []string
}

func (o *multiFieldObject) Reset() {
	o.Name = ""
	o.Score = 0
	o.Tags = o.Tags[:0]
}

// --- Тесты ---

// TestNew проверяет, что конструктор возвращает непустой пул.
func TestNew(t *testing.T) {
	p := New(func() *mockObject {
		return &mockObject{}
	})
	if p == nil {
		t.Fatal("New() вернул nil")
	}
}

// TestGetReturnsObject проверяет, что Get возвращает объект (не nil).
func TestGetReturnsObject(t *testing.T) {
	p := New(func() *mockObject {
		return &mockObject{Value: "initial"}
	})

	obj := p.Get()
	if obj == nil {
		t.Fatal("Get() вернул nil")
	}
}

// TestGetUsesFactory проверяет, что Get вызывает фабрику при пустом пуле.
func TestGetUsesFactory(t *testing.T) {
	callCount := 0
	p := New(func() *mockObject {
		callCount++
		return &mockObject{}
	})

	p.Get()
	p.Get()

	if callCount < 2 {
		t.Errorf("ожидали минимум 2 вызова фабрики, получили %d", callCount)
	}
}

// TestPutCallsReset проверяет, что Put вызывает Reset на объекте.
func TestPutCallsReset(t *testing.T) {
	p := New(func() *mockObject {
		return &mockObject{}
	})

	obj := p.Get()
	obj.Value = "dirty"

	p.Put(obj)

	if obj.ResetCount != 1 {
		t.Errorf("ожидали 1 вызов Reset, получили %d", obj.ResetCount)
	}
	if obj.Value != "" {
		t.Errorf("ожидали пустой Value после Reset, получили %q", obj.Value)
	}
}

// TestGetAfterPutReturnsCleanObject проверяет, что объект после Put
// возвращается чистым (Reset был вызван до возврата в пул).
func TestGetAfterPutReturnsCleanObject(t *testing.T) {
	p := New(func() *mockObject {
		return &mockObject{}
	})

	obj := p.Get()
	obj.Value = "some data"
	p.Put(obj)

	// Следующий Get может вернуть тот же объект из пула
	reused := p.Get()
	if reused.Value != "" {
		t.Errorf("объект не был сброшен: Value = %q", reused.Value)
	}
}

// TestMultiFieldReset проверяет сброс объекта с несколькими полями.
func TestMultiFieldReset(t *testing.T) {
	p := New(func() *multiFieldObject {
		return &multiFieldObject{}
	})

	obj := p.Get()
	obj.Name = "Alice"
	obj.Score = 42
	obj.Tags = append(obj.Tags, "go", "pool")
	p.Put(obj)

	reused := p.Get()
	if reused.Name != "" || reused.Score != 0 || len(reused.Tags) != 0 {
		t.Errorf(
			"неполный сброс: Name=%q Score=%d Tags=%v",
			reused.Name, reused.Score, reused.Tags,
		)
	}
}

// TestConcurrentGetPut проверяет корректность при конкурентном использовании.
func TestConcurrentGetPut(t *testing.T) {
	p := New(func() *mockObject {
		return &mockObject{}
	})

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			obj := p.Get()
			obj.Value = "work"
			p.Put(obj)
		}()
	}

	wg.Wait()
	// Если гонки данных или паники не случилось — тест пройден.
}
