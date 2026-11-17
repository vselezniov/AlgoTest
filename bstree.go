package main

import (
	"cmp"
	"iter"
)

// ============================================================================
// 1. ПРИЁМ «ДВОЙНОЙ УКАЗАТЕЛЬ» (**bstNode) — НАВИГАЦИЯ БЕЗ ПОЛЯ PARENT
// ============================================================================
// Обычный указатель *bstNode хранит адрес самого узла.
// Двойной указатель **bstNode (link) хранит адрес ПОЛЯ, где лежит ссылка (t.root, curr.left или curr.right).
// Благодаря этому вставка и удаление меняют связь у родителя одной строкой (*link = newNode),
// не храня в каждом узле лишнюю ссылку на родителя.

type bstNode[K, V any] struct {
	key   K
	value V
	left  *bstNode[K, V]
	right *bstNode[K, V]
}

func (n *bstNode[K, V]) height() int {
	if n == nil {
		return 0
	}
	return 1 + max(n.left.height(), n.right.height())
}

// bstree — классическое двоичное дерево поиска без балансировки.
// Эталонная учебная структура: лаконичная и быстрая на случайных данных,
// но при последовательной вставке вырождается в односвязный список за O(N).
type bstree[K, V any] struct {
	root    *bstNode[K, V]
	size    int
	compare Comparator[K]
}

func newBSTree[K, V any](compare Comparator[K]) *bstree[K, V] {
	if compare == nil {
		panic("bstree: comparator cannot be nil")
	}
	return &bstree[K, V]{compare: compare}
}

// findSlot спускается от корня и возвращает адрес ссылки (**bstNode) на найденный узел
// или на пустой nil-слот, куда ключ должен быть вставлен.
func (t *bstree[K, V]) findSlot(key K) **bstNode[K, V] {
	link := &t.root
	for *link != nil {
		curr := *link
		res := t.compare(key, curr.key)
		if res == 0 {
			return link
		}
		if res < 0 {
			link = &curr.left
		} else {
			link = &curr.right
		}
	}
	return link
}

func (t *bstree[K, V]) minLink(start **bstNode[K, V]) **bstNode[K, V] {
	link := start
	for *link != nil && (*link).left != nil {
		link = &(*link).left
	}
	return link
}

func (t *bstree[K, V]) maxLink(start **bstNode[K, V]) **bstNode[K, V] {
	link := start
	for *link != nil && (*link).right != nil {
		link = &(*link).right
	}
	return link
}

// ============================================================================
// 2. ВСТАВКА, ПОИСК И УДАЛЕНИЕ (АЛГОРИТМ ХИББАРДА)
// ============================================================================

func (t *bstree[K, V]) put(key K, val V, overwrite bool) bool {
	link := t.findSlot(key)
	if *link != nil {
		if overwrite {
			(*link).value = val
		}
		return false
	}
	*link = &bstNode[K, V]{key: key, value: val}
	t.size++
	return true
}

func (t *bstree[K, V]) get(key K) (val V, ok bool) {
	if link := t.findSlot(key); *link != nil {
		return (*link).value, true
	}
	return
}

// unlinkNode вырезает узел по двойному указателю link (3 классических случая):
//  1. Нет левого ребёнка  -->  *link = target.right
//  2. Нет правого ребёнка -->  *link = target.left
//  3. Есть оба ребёнка    -->  находим самый левый узел справа (minNode),
//     вырезаем его из правого поддерева и ставим на место target.
func (t *bstree[K, V]) unlinkNode(link **bstNode[K, V]) {
	target := *link
	if target.left == nil {
		*link = target.right
	} else if target.right == nil {
		*link = target.left
	} else {
		minL := t.minLink(&target.right)
		minNode := *minL
		*minL = minNode.right

		minNode.left = target.left
		minNode.right = target.right
		*link = minNode
	}
	t.size--
}

func (t *bstree[K, V]) deleteKey(key K) bool {
	link := t.findSlot(key)
	if *link == nil {
		return false
	}
	t.unlinkNode(link)
	return true
}

func (t *bstree[K, V]) deleteRange(lo, hi K) {
	if t == nil || t.compare(lo, hi) > 0 {
		return
	}
	t.deleteRangeAt(&t.root, lo, hi)
}

// deleteRangeAt обходит только ветви, пересекающие [lo, hi], и удаляет узлы снизу вверх (post-order),
// чтобы перестановка узла-родителя не сбила обход его ещё не проверенных потомков.
func (t *bstree[K, V]) deleteRangeAt(link **bstNode[K, V], lo, hi K) {
	n := *link
	if n == nil {
		return
	}
	if t.compare(n.key, lo) > 0 {
		t.deleteRangeAt(&n.left, lo, hi)
	}
	if t.compare(n.key, hi) < 0 {
		t.deleteRangeAt(&n.right, lo, hi)
	}
	if t.compare(n.key, lo) >= 0 && t.compare(n.key, hi) <= 0 {
		t.unlinkNode(link)
	}
}

// ============================================================================
// 3. ПОИСК ГРАНИЦ И БЛИЖАЙШИХ СОСЕДЕЙ
// ============================================================================

func (t *bstree[K, V]) min() (k K, v V, ok bool) {
	if t != nil {
		if n := *t.minLink(&t.root); n != nil {
			return n.key, n.value, true
		}
	}
	return
}

func (t *bstree[K, V]) max() (k K, v V, ok bool) {
	if t != nil {
		if n := *t.maxLink(&t.root); n != nil {
			return n.key, n.value, true
		}
	}
	return
}

// floor ищет наибольший ключ <= key, запоминая последнего подходящего кандидата при повороте направо.
func (t *bstree[K, V]) floor(key K) (k K, v V, ok bool) {
	var best *bstNode[K, V]
	for n := t.root; n != nil; {
		c := t.compare(key, n.key)
		switch {
		case c == 0:
			return n.key, n.value, true
		case c > 0:
			best = n
			n = n.right
		default:
			n = n.left
		}
	}
	if best != nil {
		return best.key, best.value, true
	}
	return
}

// ceil ищет наименьший ключ >= key, запоминая последнего подходящего кандидата при повороте налево.
func (t *bstree[K, V]) ceil(key K) (k K, v V, ok bool) {
	var best *bstNode[K, V]
	for n := t.root; n != nil; {
		c := t.compare(key, n.key)
		switch {
		case c == 0:
			return n.key, n.value, true
		case c < 0:
			best = n
			n = n.left
		default:
			n = n.right
		}
	}
	if best != nil {
		return best.key, best.value, true
	}
	return
}

func (t *bstree[K, V]) lower(key K) (k K, v V, ok bool) {
	var best *bstNode[K, V]
	for n := t.root; n != nil; {
		if t.compare(key, n.key) > 0 {
			best = n
			n = n.right
		} else {
			n = n.left
		}
	}
	if best != nil {
		return best.key, best.value, true
	}
	return
}

func (t *bstree[K, V]) higher(key K) (k K, v V, ok bool) {
	var best *bstNode[K, V]
	for n := t.root; n != nil; {
		if t.compare(key, n.key) < 0 {
			best = n
			n = n.left
		} else {
			n = n.right
		}
	}
	if best != nil {
		return best.key, best.value, true
	}
	return
}

func (t *bstree[K, V]) Len() int {
	if t == nil {
		return 0
	}
	return t.size
}

func (t *bstree[K, V]) Clear() {
	if t != nil {
		t.root = nil
		t.size = 0
	}
}

func (t *bstree[K, V]) height() int {
	if t == nil {
		return 0
	}
	return t.root.height()
}

// ============================================================================
// 4. ИТЕРАТИВНЫЕ ОБХОДЫ НА ЯВНОМ СТЕКЕ (ЗАЩИТА ОТ STACK OVERFLOW)
// ============================================================================
// В несбалансированном BST глубина ветви может достигать сотен тысяч уровней.
// Хранение пути в обычном слайсе stack (в динамической памяти/heap) полностью защищает
// от аварийного переполнения стека вызовов при обходе вырожденного дерева.

func (t *bstree[K, V]) walk(reverse bool) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if t == nil {
			return
		}
		var stack []*bstNode[K, V]
		curr := t.root

		for curr != nil || len(stack) > 0 {
			for curr != nil {
				stack = append(stack, curr)
				if reverse {
					curr = curr.right
				} else {
					curr = curr.left
				}
			}

			curr = stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if !yield(curr.key, curr.value) {
				return
			}

			if reverse {
				curr = curr.left
			} else {
				curr = curr.right
			}
		}
	}
}

// walkFrom сразу спускается к границе bound, запоминая в стек только узлы,
// которые лежат по нужную сторону от границы.
func (t *bstree[K, V]) walkFrom(bound K, reverse bool) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if t == nil {
			return
		}
		var stack []*bstNode[K, V]
		for n := t.root; n != nil; {
			cmpRes := t.compare(bound, n.key)
			if !reverse {
				if cmpRes <= 0 {
					stack = append(stack, n)
					n = n.left
				} else {
					n = n.right
				}
			} else {
				if cmpRes >= 0 {
					stack = append(stack, n)
					n = n.right
				} else {
					n = n.left
				}
			}
		}

		for len(stack) > 0 {
			n := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if !yield(n.key, n.value) {
				return
			}

			if !reverse {
				for c := n.right; c != nil; c = c.left {
					stack = append(stack, c)
				}
			} else {
				for c := n.left; c != nil; c = c.right {
					stack = append(stack, c)
				}
			}
		}
	}
}

func (t *bstree[K, V]) walkTo(bound K, reverse bool) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range t.walk(reverse) {
			cmpRes := t.compare(k, bound)
			if (!reverse && cmpRes > 0) || (reverse && cmpRes < 0) {
				return
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// ============================================================================
// 5. КОНСТРУКТОРЫ ВТОРОЙ РЕАЛИЗАЦИИ (НА БАЗЕ КЛАССИЧЕСКОГО BST)
// ============================================================================

func NewSet2[T cmp.Ordered]() *Set[T] {
	return &Set[T]{tree: newBSTree[T, struct{}](cmp.Compare[T])}
}

func NewOrderedSet2[T any](compare Comparator[T]) *Set[T] {
	return &Set[T]{tree: newBSTree[T, struct{}](compare)}
}

func NewMap2[K cmp.Ordered, V any]() *Map[K, V] {
	return &Map[K, V]{tree: newBSTree[K, V](cmp.Compare[K])}
}

func NewOrderedMap2[K, V any](compare Comparator[K]) *Map[K, V] {
	return &Map[K, V]{tree: newBSTree[K, V](compare)}
}

func NewIntervalMap2[K cmp.Ordered, V comparable](defaultVal V) *IntervalMap[K, V] {
	return &IntervalMap[K, V]{defaultVal: defaultVal, m: NewMap2[K, V]()}
}
