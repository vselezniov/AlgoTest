package main

import (
	"cmp"
	"iter"
)

// entry — пара «ключ-значение», которая хранится внутри узла.
type entry[K, V any] struct {
	key K
	val V
}

// node23 — узел 2-3 дерева.
// Вмещает 1 или 2 ключа (и на миг 0 ключей во время балансировки при удалении).
type node23[K, V any] struct {
	n          int           // текущее количество ключей в узле (0, 1 или 2)
	e1, e2     entry[K, V]   // первый (меньший) и второй (больший) ключи
	c0, c1, c2 *node23[K, V] // левый, средний и правый потомки
}

// isLeaf возвращает true, если у узла нет детей (в 2-3 дереве дети либо все, либо ни одного).
func (n *node23[K, V]) isLeaf() bool { return n.c0 == nil }

// child возвращает потомка по номеру ветви: 0 (левый), 1 (средний) или 2 (правый).
func (n *node23[K, V]) child(idx int) *node23[K, V] {
	switch idx {
	case 0:
		return n.c0
	case 1:
		return n.c1
	default:
		return n.c2
	}
}

// entry возвращает пару ключ-значение по индексу (0 — первая, 1 — вторая).
func (n *node23[K, V]) entry(idx int) entry[K, V] {
	if idx == 0 {
		return n.e1
	}
	return n.e2
}

// setEntry записывает новую пару ключ-значение в слот 0 или 1.
func (n *node23[K, V]) setEntry(idx int, e entry[K, V]) {
	if idx == 0 {
		n.e1 = e
	} else {
		n.e2 = e
	}
}

// eraseLeafKey удаляет ключ из листа по индексу (0 или 1) и сдвигает оставшийся ключ.
func (n *node23[K, V]) eraseLeafKey(idx int) {
	if idx == 0 {
		n.e1 = n.e2 // если удалили первый ключ, второй встаёт на его место
	}
	n.e2 = entry[K, V]{} // зануляем второй слот, чтобы сборщик мусора освободил память
	n.n--
}

// tree23 — идеально сбалансированное поисковое дерево.
// Все его листья всегда находятся на одной глубине, поэтому поиск занимает O(log N).
type tree23[K, V any] struct {
	root    *node23[K, V] // корень дерева
	size    int           // общее количество ключей во всём дереве
	compare Comparator[K] // функция сравнения двух ключей
}

// newTree23 создаёт пустое дерево с заданным правилом сравнения.
func newTree23[K, V any](compare Comparator[K]) *tree23[K, V] {
	if compare == nil {
		panic("tree23: comparator cannot be nil")
	}
	return &tree23[K, V]{compare: compare}
}

// Len возвращает количество элементов в дереве.
func (t *tree23[K, V]) Len() int {
	if t == nil {
		return 0
	}
	return t.size
}

// Clear мгновенно очищает дерево.
func (t *tree23[K, V]) Clear() {
	if t != nil {
		t.root = nil
		t.size = 0
	}
}

// height считает высоту дерева. Достаточно идти только по левым детям,
// так как в 2-3 дереве абсолютно все ветви имеют одинаковую длину.
func (t *tree23[K, V]) height() int {
	h := 0
	for n := t.root; n != nil; n = n.c0 {
		h++
	}
	return h
}

// locate проверяет, есть ли key внутри узла n.
// Если ключ найден — возвращает (индекс_ключа, -1).
// Если нет — возвращает (-1, номер_ветки_для_спуска).
func (t *tree23[K, V]) locate(n *node23[K, V], key K) (keyIdx, childIdx int) {
	c1 := t.compare(key, n.e1.key)
	switch {
	case c1 == 0:
		return 0, -1 // совпал с первым ключом
	case c1 < 0:
		return -1, 0 // меньше первого ключа -> идём в левую ветку (c0)
	case n.n == 1:
		return -1, 1 // больше первого ключа, а второго нет -> идём в правую ветку (c1)
	}

	c2 := t.compare(key, n.e2.key)
	switch {
	case c2 == 0:
		return 1, -1 // совпал со вторым ключом
	case c2 < 0:
		return -1, 1 // между первым и вторым -> идём в среднюю ветку (c1)
	default:
		return -1, 2 // больше второго ключа -> идём в крайнюю правую ветку (c2)
	}
}

// get ищет значение по ключу, спускаясь от корня к листьям.
func (t *tree23[K, V]) get(key K) (val V, ok bool) {
	for n := t.root; n != nil; {
		keyIdx, childIdx := t.locate(n, key)
		if keyIdx != -1 {
			return n.entry(keyIdx).val, true
		}
		n = n.child(childIdx)
	}
	return
}

// split — «осколок» от расколотого узла, который всплывает к родителю:
// средний ключ (mid) и новый правый брат (right).
type split[K, V any] struct {
	mid   entry[K, V]
	right *node23[K, V]
}

// put добавляет ключ в дерево (или обновляет значение, если overwrite == true).
// Возвращает true, если был добавлен именно новый ключ.
func (t *tree23[K, V]) put(key K, val V, overwrite bool) bool {
	if t.root == nil {
		t.root = &node23[K, V]{n: 1, e1: entry[K, V]{key, val}}
		t.size = 1
		return true
	}

	up, inserted := t.insertAt(t.root, split[K, V]{mid: entry[K, V]{key, val}}, overwrite)
	if up.right != nil {
		// Корень тоже раскололся! Создаём новый корень этажом выше.
		t.root = &node23[K, V]{n: 1, e1: up.mid, c0: t.root, c1: up.right}
	}
	if inserted {
		t.size++
	}
	return inserted
}

// insertAt рекурсивно спускается до листа, вставляет ключ и при необходимости
// поднимает вверх всплывающий осколок (up).
func (t *tree23[K, V]) insertAt(n *node23[K, V], in split[K, V], overwrite bool) (up split[K, V], inserted bool) {
	keyIdx, childIdx := t.locate(n, in.mid.key)
	if keyIdx != -1 {
		if overwrite {
			n.setEntry(keyIdx, in.mid) // ключ уже есть — просто обновляем значение
		}
		return
	}

	if !n.isLeaf() {
		// Спускаемся глубже; если снизу ничего не всплыло (in.right == nil) — работа закончена.
		in, inserted = t.insertAt(n.child(childIdx), in, overwrite)
		if in.right == nil {
			return split[K, V]{}, inserted
		}
	}

	// В узле есть свободное место (1 ключ) — просто кладём второй ключ рядом.
	if n.n == 1 {
		n.insert(childIdx, in)
		return split[K, V]{}, true
	}

	// Узел полон (уже 2 ключа) — раскалываем его на два узла.
	return n.split(childIdx, in), true
}

// insert укладывает всплывающий пакет in в свободный слот 2-узла (на позицию 0 или 1).
func (n *node23[K, V]) insert(idx int, in split[K, V]) {
	if idx == 0 {
		n.e2, n.e1 = n.e1, in.mid
		n.c2, n.c1 = n.c1, in.right
	} else {
		n.e2, n.c2 = in.mid, in.right
	}
	n.n = 2
}

// split делит переполненный 3-узел (где уже 2 ключа + пришёл 3-й) на два 2-узла.
// Средний из трёх ключей отправляется наверх к родителю.
func (n *node23[K, V]) split(idx int, in split[K, V]) (up split[K, V]) {
	switch idx {
	case 0: // новый ключ самый маленький -> всплывает старый e1
		up.mid = n.e1
		up.right = &node23[K, V]{n: 1, e1: n.e2, c0: n.c1, c1: n.c2}
		n.e1, n.c1 = in.mid, in.right
	case 1: // новый ключ посередине -> он сам и всплывает
		up.mid = in.mid
		up.right = &node23[K, V]{n: 1, e1: n.e2, c0: in.right, c1: n.c2}
	default: // новый ключ самый большой -> всплывает старый e2
		up.mid = n.e2
		up.right = &node23[K, V]{n: 1, e1: in.mid, c0: n.c2, c1: in.right}
	}

	n.n = 1
	n.e2 = entry[K, V]{}
	n.c2 = nil
	return up
}

// minLeafEntry спускается до упора влево в поддереве n и возвращает наименьший элемент.
func (t *tree23[K, V]) minLeafEntry(n *node23[K, V]) entry[K, V] {
	for n.c0 != nil {
		n = n.c0
	}
	return n.e1
}

// deleteKey удаляет ключ из дерева. Если корень опустел — дерево становится на этаж ниже.
func (t *tree23[K, V]) deleteKey(key K) bool {
	if t.root == nil || !t.deleteAt(t.root, key) {
		return false
	}
	if t.root.n == 0 {
		t.root = t.root.c0 // единственный оставшийся потомок становится новым корнем
	}
	t.size--
	return true
}

// deleteAt находит ключ, удаляет его из листа и на обратном пути чинит опустевшие узлы.
func (t *tree23[K, V]) deleteAt(n *node23[K, V], key K) bool {
	if n == nil {
		return false
	}

	keyIdx, childIdx := t.locate(n, key)

	if keyIdx != -1 {
		if n.isLeaf() {
			n.eraseLeafKey(keyIdx) // мы в листе — просто стираем ключ
			return true
		}
		// Мы во внутреннем узле: подменяем ключ на минимум из правого поддерева
		// и продолжаем спуск, чтобы удалить этот дубликат из листа.
		childIdx = keyIdx + 1
		minRight := t.minLeafEntry(n.child(childIdx))
		n.setEntry(keyIdx, minRight)
		key = minRight.key
	}

	if !t.deleteAt(n.child(childIdx), key) {
		return false
	}
	t.fixChild(n, childIdx) // проверяем, не остался ли потомок пустым
	return true
}

// fixChild восстанавливает баланс, если потомок idx остался без ключей (n == 0).
func (t *tree23[K, V]) fixChild(p *node23[K, V], idx int) {
	if p.child(idx).n > 0 {
		return // с потомком всё в порядке
	}
	// 1. Пробуем занять ключ у левого брата, если у него два ключа
	if idx > 0 && p.child(idx-1).n == 2 {
		t.rotateRight(p, idx-1)
		return
	}
	// 2. Пробуем занять ключ у правого брата, если у него два ключа
	if idx < p.n && p.child(idx+1).n == 2 {
		t.rotateLeft(p, idx)
		return
	}
	// 3. Оба брата бедные (по 1 ключу) — склеиваем пустого потомка с братом и ключом из родителя
	t.mergeChildren(p, max(0, idx-1))
}

// rotateRight переносит ключ из богатого левого брата в пустого правого через родителя.
func (t *tree23[K, V]) rotateRight(p *node23[K, V], sepIdx int) {
	left, right := p.child(sepIdx), p.child(sepIdx+1)

	right.e1 = p.entry(sepIdx) // ключ родителя опускается в пустого правого брата
	right.n = 1
	right.c1, right.c0 = right.c0, left.c2 // перекидываем крайнюю ветку от левого к правому

	p.setEntry(sepIdx, left.e2) // старший ключ левого брата поднимается в родителя

	left.e2, left.n, left.c2 = entry[K, V]{}, 1, nil
}

// rotateLeft переносит ключ из богатого правого брата в пустого левого через родителя.
func (t *tree23[K, V]) rotateLeft(p *node23[K, V], sepIdx int) {
	left, right := p.child(sepIdx), p.child(sepIdx+1)

	left.e1 = p.entry(sepIdx) // ключ родителя опускается в пустого левого брата
	left.n = 1
	left.c1 = right.c0 // перекидываем крайнюю левую ветку от правого брата

	p.setEntry(sepIdx, right.e1) // младший ключ правого брата поднимается в родителя

	right.e1, right.e2, right.n = right.e2, entry[K, V]{}, 1
	right.c0, right.c1, right.c2 = right.c1, right.c2, nil
}

// mergeChildren склеивает двух братьев и разделитель sepIdx из родителя в один 3-узел.
func (t *tree23[K, V]) mergeChildren(p *node23[K, V], sepIdx int) {
	left, right := p.child(sepIdx), p.child(sepIdx+1)

	if left.n == 0 {
		left.e1, left.e2 = p.entry(sepIdx), right.e1
		left.c1, left.c2 = right.c0, right.c1
	} else {
		left.e2 = p.entry(sepIdx)
		left.c2 = right.c0
	}
	left.n = 2

	// Убираем использованный ключ-разделитель из родителя
	if sepIdx == 0 {
		p.e1 = p.e2
		p.c1 = p.c2
	}
	p.e2, p.c2 = entry[K, V]{}, nil
	p.n--
}

// deleteRange удаляет все ключи из отрезка [lo, hi] включительно.
func (t *tree23[K, V]) deleteRange(lo, hi K) {
	if t == nil || t.compare(lo, hi) > 0 {
		return
	}
	for k, _, ok := t.ceil(lo); ok && t.compare(k, hi) <= 0; k, _, ok = t.ceil(lo) {
		t.deleteKey(k)
	}
}

// min возвращает самый маленький ключ в дереве.
func (t *tree23[K, V]) min() (k K, v V, ok bool) {
	if t == nil || t.root == nil {
		return
	}
	n := t.root
	for !n.isLeaf() {
		n = n.c0
	}
	return n.e1.key, n.e1.val, true
}

// max возвращает самый большой ключ в дереве.
func (t *tree23[K, V]) max() (k K, v V, ok bool) {
	if t == nil || t.root == nil {
		return
	}
	n := t.root
	for !n.isLeaf() {
		n = n.child(n.n)
	}
	e := n.entry(n.n - 1)
	return e.key, e.val, true
}

// bracket спускается к «щели» для key и возвращает ближайших соседей слева (lo) и справа (hi).
// Если в дереве встретится ключ, равный key, флаг equalGoesLeft решает, куда его отнести:
//   - true  -> ключ считается левым соседом (нужно для floor <= и higher >);
//   - false -> ключ считается правым соседом (нужно для ceil >= и lower <).
func (t *tree23[K, V]) bracket(key K, equalGoesLeft bool) (lo, hi entry[K, V], hasLo, hasHi bool) {
	for n := t.root; n != nil; {
		i := 0
		for ; i < n.n; i++ {
			c := t.compare(n.entry(i).key, key)
			if c > 0 || (c == 0 && !equalGoesLeft) {
				break
			}
		}
		if i > 0 {
			lo, hasLo = n.entry(i-1), true // уточняем ближайшего соседа слева
		}
		if i < n.n {
			hi, hasHi = n.entry(i), true // уточняем ближайшего соседа справа
		}
		n = n.child(i)
	}
	return
}

// floor ищет наибольший ключ <= key.
func (t *tree23[K, V]) floor(key K) (K, V, bool) {
	lo, _, ok, _ := t.bracket(key, true)
	return lo.key, lo.val, ok
}

// ceil ищет наименьший ключ >= key.
func (t *tree23[K, V]) ceil(key K) (K, V, bool) {
	_, hi, _, ok := t.bracket(key, false)
	return hi.key, hi.val, ok
}

// lower ищет наибольший ключ, который строго меньше key (<).
func (t *tree23[K, V]) lower(key K) (K, V, bool) {
	lo, _, ok, _ := t.bracket(key, false)
	return lo.key, lo.val, ok
}

// higher ищет наименьший ключ, который строго больше key (>).
func (t *tree23[K, V]) higher(key K) (K, V, bool) {
	_, hi, _, ok := t.bracket(key, true)
	return hi.key, hi.val, ok
}

// scan — универсальный генератор (iter.Seq2) для обхода дерева от start до stop.
// Если start или stop равны nil — соответствующая граница не ограничена.
func (t *tree23[K, V]) scan(start, stop *K, reverse bool) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if t == nil || t.root == nil {
			return
		}

		// visit проверяет границы [start, stop] для конкретного ключа и отдаёт его в цикл range.
		visit := func(e entry[K, V]) bool {
			if start != nil {
				if (!reverse && t.compare(e.key, *start) < 0) || (reverse && t.compare(e.key, *start) > 0) {
					return true // ключ остался до старта — просто пропускаем его
				}
			}
			if stop != nil {
				if (!reverse && t.compare(e.key, *stop) > 0) || (reverse && t.compare(e.key, *stop) < 0) {
					return false // вышли за границу stop — прерываем весь обход
				}
			}
			return yield(e.key, e.val)
		}

		var dfs func(n *node23[K, V]) bool
		dfs = func(n *node23[K, V]) bool {
			if n == nil {
				return true
			}

			if !reverse {
				// Прямой порядок: левая ветка -> 1-й ключ -> средняя ветка -> [2-й ключ -> правая ветка]
				if start == nil || t.compare(n.e1.key, *start) >= 0 {
					if !dfs(n.c0) {
						return false
					}
				}
				if !visit(n.e1) {
					return false
				}
				if n.n == 1 {
					return dfs(n.c1)
				}
				if start == nil || t.compare(n.e2.key, *start) >= 0 {
					if !dfs(n.c1) {
						return false
					}
				}
				if !visit(n.e2) {
					return false
				}
				return dfs(n.c2)
			}

			// Обратный порядок: [правая ветка -> 2-й ключ] -> средняя ветка -> 1-й ключ -> левая ветка
			if n.n == 2 {
				if start == nil || t.compare(n.e2.key, *start) <= 0 {
					if !dfs(n.c2) {
						return false
					}
				}
				if !visit(n.e2) {
					return false
				}
			}
			if start == nil || t.compare(n.e1.key, *start) <= 0 {
				if !dfs(n.c1) {
					return false
				}
			}
			if !visit(n.e1) {
				return false
			}
			return dfs(n.c0)
		}

		dfs(t.root)
	}
}

// walk обходит все элементы дерева (по возрастанию или по убыванию при reverse == true).
func (t *tree23[K, V]) walk(reverse bool) iter.Seq2[K, V] {
	return t.scan(nil, nil, reverse)
}

// walkFrom начинает обход с границы bound и идёт до конца дерева.
func (t *tree23[K, V]) walkFrom(bound K, reverse bool) iter.Seq2[K, V] {
	return t.scan(&bound, nil, reverse)
}

// walkTo идёт с самого начала дерева и останавливается на границе bound включительно.
func (t *tree23[K, V]) walkTo(bound K, reverse bool) iter.Seq2[K, V] {
	return t.scan(nil, &bound, reverse)
}

// ============================================================================
// ПУБЛИЧНЫЕ КОНСТРУКТОРЫ КОЛЛЕКЦИЙ НА БАЗЕ 2-3 ДЕРЕВА
// ============================================================================

// NewSet создаёт упорядоченное множество для стандартных сравнимых типов (int, string и др.).
func NewSet[T cmp.Ordered]() *Set[T] {
	return &Set[T]{tree: newTree23[T, struct{}](cmp.Compare[T])}
}

// NewOrderedSet создаёт упорядоченное множество со своей функцией сравнения.
func NewOrderedSet[T any](compare Comparator[T]) *Set[T] {
	return &Set[T]{tree: newTree23[T, struct{}](compare)}
}

// NewMap создаёт упорядоченный словарь для стандартных сравнимых ключей.
func NewMap[K cmp.Ordered, V any]() *Map[K, V] {
	return &Map[K, V]{tree: newTree23[K, V](cmp.Compare[K])}
}

// NewOrderedMap создаёт упорядоченный словарь со своей функцией сравнения ключей.
func NewOrderedMap[K, V any](compare Comparator[K]) *Map[K, V] {
	return &Map[K, V]{tree: newTree23[K, V](compare)}
}

// NewIntervalMap создаёт карту полуинтервалов со значением фона defaultVal.
func NewIntervalMap[K cmp.Ordered, V comparable](defaultVal V) *IntervalMap[K, V] {
	return &IntervalMap[K, V]{defaultVal: defaultVal, m: NewMap[K, V]()}
}
