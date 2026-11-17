package main

import (
	"cmp"
	"fmt"
	"iter"
)

// ============================================================================
// 1. ПОРЯДОК СРАВНЕНИЯ И ИНТЕРФЕЙС ДЕРЕВА (ПАТТЕРН «МОСТ» / BRIDGE)
// ============================================================================
// Идея: мы отделяем пользовательские коллекции (Set, Map, IntervalMap) от того,
// как именно данные лежат в памяти (в 2-3 дереве или обычном BST).

// Comparator задаёт строгий линейный порядок:
//
//	< 0, если a < b
//	  0, если a == b
//	> 0, если a > b
type Comparator[K any] func(a, b K) int

// Reverse меняет знак результата на противоположный, разворачивая сортировку.
func (c Comparator[K]) Reverse() Comparator[K] {
	return func(a, b K) int { return -c(a, b) }
}

// tree — минимальный контракт, который обязано уметь любое поисковое дерево.
type tree[K, V any] interface {
	put(key K, val V, overwrite bool) bool
	get(key K) (V, bool)
	deleteKey(key K) bool
	deleteRange(lo, hi K)

	min() (K, V, bool)
	max() (K, V, bool)
	floor(key K) (K, V, bool)  // <= key
	ceil(key K) (K, V, bool)   // >= key
	lower(key K) (K, V, bool)  // <  key
	higher(key K) (K, V, bool) // >  key

	Len() int
	Clear()
	height() int

	walk(reverse bool) iter.Seq2[K, V]
	walkFrom(bound K, reverse bool) iter.Seq2[K, V]
	walkTo(bound K, reverse bool) iter.Seq2[K, V]
}

// ============================================================================
// 2. МНОЖЕСТВО: Set[T]
// ============================================================================
// Трюк Go: struct{} занимает 0 байт памяти. Поэтому Set — это просто Map,
// у которого вместо значений хранится пустота.

type Set[T any] struct {
	tree[T, struct{}]
}

func (s *Set[T]) Add(val T) bool { return s.put(val, struct{}{}, false) }
func (s *Set[T]) DeleteRange(lo, hi T) {
	if s != nil {
		s.deleteRange(lo, hi)
	}
}

func (s *Set[T]) Contains(val T) bool {
	if s == nil {
		return false
	}
	_, ok := s.get(val)
	return ok
}

func (s *Set[T]) Delete(val T) bool {
	if s == nil {
		return false
	}
	return s.deleteKey(val)
}

func (s *Set[T]) Min() (val T, ok bool) {
	if s != nil {
		val, _, ok = s.min()
	}
	return
}

func (s *Set[T]) Max() (val T, ok bool) {
	if s != nil {
		val, _, ok = s.max()
	}
	return
}

func (s *Set[T]) Floor(target T) (val T, ok bool) {
	if s != nil {
		val, _, ok = s.floor(target)
	}
	return
}

func (s *Set[T]) Ceiling(target T) (val T, ok bool) {
	if s != nil {
		val, _, ok = s.ceil(target)
	}
	return
}

func (s *Set[T]) Lower(target T) (val T, ok bool) {
	if s != nil {
		val, _, ok = s.lower(target)
	}
	return
}

func (s *Set[T]) Higher(target T) (val T, ok bool) {
	if s != nil {
		val, _, ok = s.higher(target)
	}
	return
}

func (s *Set[T]) All() iter.Seq[T]              { return keysOnly(s.walk(false)) }
func (s *Set[T]) Backward() iter.Seq[T]         { return keysOnly(s.walk(true)) }
func (s *Set[T]) From(lo T) iter.Seq[T]         { return keysOnly(s.walkFrom(lo, false)) }
func (s *Set[T]) BackwardFrom(hi T) iter.Seq[T] { return keysOnly(s.walkFrom(hi, true)) }
func (s *Set[T]) To(hi T) iter.Seq[T]           { return keysOnly(s.walkTo(hi, false)) }
func (s *Set[T]) BackwardTo(lo T) iter.Seq[T]   { return keysOnly(s.walkTo(lo, true)) }

// keysOnly превращает генератор пар (K, V) в генератор одних ключей K.
func keysOnly[K, V any](seq iter.Seq2[K, V]) iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range seq {
			if !yield(k) {
				return
			}
		}
	}
}

// ============================================================================
// 3. СЛОВАРЬ: Map[K, V]
// ============================================================================

type Map[K, V any] struct {
	tree[K, V]
}

func (m *Map[K, V]) Set(key K, val V) { m.put(key, val, true) }
func (m *Map[K, V]) DeleteRange(lo, hi K) {
	if m != nil {
		m.deleteRange(lo, hi)
	}
}
func (m *Map[K, V]) All() iter.Seq2[K, V]              { return m.walk(false) }
func (m *Map[K, V]) Backward() iter.Seq2[K, V]         { return m.walk(true) }
func (m *Map[K, V]) From(lo K) iter.Seq2[K, V]         { return m.walkFrom(lo, false) }
func (m *Map[K, V]) BackwardFrom(hi K) iter.Seq2[K, V] { return m.walkFrom(hi, true) }
func (m *Map[K, V]) To(hi K) iter.Seq2[K, V]           { return m.walkTo(hi, false) }
func (m *Map[K, V]) BackwardTo(lo K) iter.Seq2[K, V]   { return m.walkTo(lo, true) }

func (m *Map[K, V]) Get(key K) (val V, ok bool) {
	if m != nil {
		return m.get(key)
	}
	return
}

func (m *Map[K, V]) Delete(key K) bool {
	if m == nil {
		return false
	}
	return m.deleteKey(key)
}

func (m *Map[K, V]) Min() (k K, v V, ok bool) {
	if m != nil {
		return m.min()
	}
	return
}

func (m *Map[K, V]) Max() (k K, v V, ok bool) {
	if m != nil {
		return m.max()
	}
	return
}

func (m *Map[K, V]) Floor(key K) (k K, v V, ok bool) {
	if m != nil {
		return m.floor(key)
	}
	return
}

func (m *Map[K, V]) Ceiling(key K) (k K, v V, ok bool) {
	if m != nil {
		return m.ceil(key)
	}
	return
}

func (m *Map[K, V]) Lower(key K) (k K, v V, ok bool) {
	if m != nil {
		return m.lower(key)
	}
	return
}

func (m *Map[K, V]) Higher(key K) (k K, v V, ok bool) {
	if m != nil {
		return m.higher(key)
	}
	return
}

// ============================================================================
// 4. КАРТА ПОЛУИНТЕРВАЛОВ: IntervalMap[K, V] (СЖАТИЕ ОСИ ЧЕРЕЗ ТОЧКИ СМЕНЫ)
// ============================================================================
// Вместо хранения миллионов точек мы храним в дереве только ТОЧКИ ПЕРЕКЛЮЧЕНИЯ:
//
//   Ось:    ---[ фон "." ]---9---[ "Анна" ]---14---[ "Борис" ]---18---[ фон "." ]--->
//   Дерево:                  9:"Анна"         14:"Борис"         18:"."

type Interval[K cmp.Ordered, V comparable] struct {
	From  K
	To    K
	Value V
}

type IntervalMap[K cmp.Ordered, V comparable] struct {
	defaultVal V
	m          *Map[K, V]
}

func (im *IntervalMap[K, V]) Clear() {
	if im != nil {
		im.m.Clear()
	}
}

// Assign заливает полуинтервал [from, to) значением val за 4 шага:
// 1. Узнаём valAfter — какое значение действовало в точке to (чтобы восстановить хвост).
// 2. Сносим все старые границы внутри [from, to) вызовом DeleteRange.
// 3. Ставим границу в from, ТОЛЬКО если слева от неё было другое значение (иначе отрезки склеятся!).
// 4. Ставим границу в to, ТОЛЬКО если справа от неё (valAfter) другое значение.
func (im *IntervalMap[K, V]) Assign(from, to K, val V) {
	if !(from < to) {
		return
	}
	valAfter := im.At(to)
	im.m.DeleteRange(from, to)
	if im.At(from) != val {
		im.m.Set(from, val)
	}
	if valAfter != val {
		im.m.Set(to, valAfter)
	}
}

// At(point) — это просто поиск ближайшей точки смены слева или точно в point (Floor).
func (im *IntervalMap[K, V]) At(point K) V {
	if _, v, ok := im.m.Floor(point); ok {
		return v
	}
	return im.defaultVal
}

func (im *IntervalMap[K, V]) IntervalAt(point K) (val V, from K, to K, hasFrom bool, hasTo bool) {
	if k, v, ok := im.m.Floor(point); ok {
		val, from, hasFrom = v, k, true
	} else {
		val = im.defaultVal
	}
	if k, _, ok := im.m.Higher(point); ok {
		to, hasTo = k, true
	}
	return
}

func (im *IntervalMap[K, V]) IsUniform(from, to K, expected V) bool {
	if !(from < to) {
		return true
	}
	if im.At(from) != expected {
		return false
	}
	if nextK, _, ok := im.m.Higher(from); ok {
		return nextK >= to
	}
	return true
}

func (im *IntervalMap[K, V]) QueryRange(from, to K) iter.Seq[Interval[K, V]] {
	return func(yield func(Interval[K, V]) bool) {
		if !(from < to) {
			return
		}
		currStart := from
		currVal := im.At(from)

		for k, v := range im.m.From(from) {
			if k >= to {
				break
			}
			if k > currStart {
				if !yield(Interval[K, V]{From: currStart, To: k, Value: currVal}) {
					return
				}
			}
			currStart = k
			currVal = v
		}
		if currStart < to {
			yield(Interval[K, V]{From: currStart, To: to, Value: currVal})
		}
	}
}

func (im *IntervalMap[K, V]) All() iter.Seq[Interval[K, V]] {
	return func(yield func(Interval[K, V]) bool) {
		var (
			prevK   K
			prevV   V
			hasPrev bool
		)
		for k, v := range im.m.All() {
			if hasPrev {
				if !yield(Interval[K, V]{From: prevK, To: k, Value: prevV}) {
					return
				}
			}
			prevK = k
			prevV = v
			hasPrev = true
		}
	}
}

// validate проверяет каноничность сжатия:
// • в дереве не может быть ровно 1 узел (любой конечный отрезок имеет начало и конец);
// • два соседних узла не могут иметь одинаковое значение (иначе они обязаны склеиться);
// • самый правый узел всегда возвращает ось в фон defaultVal.
func (im *IntervalMap[K, V]) validate() error {
	if im.m.Len() == 1 {
		k, v, _ := im.m.Min()
		return fmt.Errorf("ошибка: в дереве ровно 1 узел (%v -> %v), должно быть 0 или >= 2", k, v)
	}

	prevVal := im.defaultVal
	var prevKey K
	hasPrev := false

	for k, v := range im.m.All() {
		if v == prevVal {
			if !hasPrev {
				return fmt.Errorf("ошибка: первый узел (%v) равен фону (%v)", k, v)
			}
			return fmt.Errorf("ошибка: узел (%v) повторяет соседа (%v) со значением (%v)", k, prevKey, v)
		}
		prevKey, prevVal = k, v
		hasPrev = true
	}

	if hasPrev && prevVal != im.defaultVal {
		return fmt.Errorf("ошибка: последний узел (%v) должен быть фоном (%v), получено (%v)", prevKey, im.defaultVal, prevVal)
	}
	return nil
}
