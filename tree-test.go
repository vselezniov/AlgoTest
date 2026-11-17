package main

// Демонстрация, проверка корректности и стресс-тест.
// Запуск из папки проекта:
//   go run .         — на 2-3 дереве (по умолчанию)
//   go run . bst     — на обычном двоичном дереве (bstree)

import (
	"cmp"
	"fmt"
	"iter"
	"math"
	"math/rand"
	"os"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

// ============================================================================
// 1. ВЫБОР ДЕРЕВА ДЛЯ ТЕСТА (false = 2-3 дерево, true = обычное BST)
// ============================================================================

var useBST = false

func treeName() string {
	if useBST {
		return "обычное BST (bstree)"
	}
	return "2-3 дерево (tree23)"
}

func makeSet[T cmp.Ordered]() *Set[T] {
	if useBST {
		return NewSet2[T]()
	}
	return NewSet[T]()
}

func makeMap[K cmp.Ordered, V any]() *Map[K, V] {
	if useBST {
		return NewMap2[K, V]()
	}
	return NewMap[K, V]()
}

func makeIntervalMap[K cmp.Ordered, V comparable](def V) *IntervalMap[K, V] {
	if useBST {
		return NewIntervalMap2[K, V](def)
	}
	return NewIntervalMap[K, V](def)
}

// ============================================================================
// 2. ПРОВЕРКА КОРРЕКТНОСТИ (Set, Map и IntervalMap)
// ============================================================================

func assert(cond bool, msg string) {
	if !cond {
		panic("FAIL: " + msg)
	}
}

func assertEq[T comparable](got, want []T, label string) {
	assert(len(got) == len(want), fmt.Sprintf("%s: len=%d, want=%d", label, len(got), len(want)))
	for i := range want {
		assert(got[i] == want[i], fmt.Sprintf("%s[%d]: got=%v, want=%v", label, i, got[i], want[i]))
	}
}

func toSlice[T any](seq iter.Seq[T]) (res []T) {
	for v := range seq {
		res = append(res, v)
	}
	return
}

func testCorrectness() {
	fmt.Printf("=== ПРОВЕРКА Set, Map И IntervalMap [%s] ===\n", treeName())

	// 1. Проверка Set (вставка, соседи, 6 итераторов, удаление узлов и отрезков)
	s := makeSet[int]()
	for _, x := range []int{50, 30, 70, 20, 40, 60, 80, 35, 45} {
		assert(s.Add(x), "Add")
	}
	assert(!s.Add(40) && s.Len() == 9, "Duplicate Add / Len")

	v, ok := s.Floor(42)
	assert(ok && v == 40, "Floor")
	v, ok = s.Ceiling(42)
	assert(ok && v == 45, "Ceiling")
	v, ok = s.Lower(40)
	assert(ok && v == 35, "Lower")
	v, ok = s.Higher(40)
	assert(ok && v == 45, "Higher")

	assertEq(toSlice(s.All()), []int{20, 30, 35, 40, 45, 50, 60, 70, 80}, "All")
	assertEq(toSlice(s.Backward()), []int{80, 70, 60, 50, 45, 40, 35, 30, 20}, "Backward")
	assertEq(toSlice(s.From(40)), []int{40, 45, 50, 60, 70, 80}, "From")
	assertEq(toSlice(s.BackwardFrom(42)), []int{40, 35, 30, 20}, "BackwardFrom")
	assertEq(toSlice(s.To(45)), []int{20, 30, 35, 40, 45}, "To")
	assertEq(toSlice(s.BackwardTo(50)), []int{80, 70, 60, 50}, "BackwardTo")

	assert(s.Delete(20) && s.Delete(30) && s.Delete(50), "Delete 0/2 children")
	s.DeleteRange(40, 70)
	assertEq(toSlice(s.All()), []int{35, 80}, "DeleteRange")
	fmt.Println("  [OK] Set: вставка, удаление узлов/диапазонов, соседи и 6 итераторов.")

	// 2. Проверка Map (вставка, перезапись существующего ключа, поиск и удаление)
	m := makeMap[int, string]()
	m.Set(10, "ten")
	m.Set(20, "twenty")
	m.Set(10, "TEN")
	gotVal, found := m.Get(10)
	assert(found && gotVal == "TEN" && m.Len() == 2, "Map Set/Overwrite/Get")
	assert(m.Delete(10) && !m.Delete(10) && m.Len() == 1, "Map Delete")
	fmt.Println("  [OK] Map: вставка, перезапись по ключу, поиск и удаление.")

	// 3. Проверка IntervalMap (склейка отрезков, вырезание дыр, запросы по диапазону)
	im := makeIntervalMap[int, string](".")
	check := func() { assert(im.validate() == nil, fmt.Sprint(im.validate())) }

	im.Assign(10, 10, "A")
	im.Assign(20, 5, "A")
	im.Assign(-50, 50, ".")
	assert(im.m.Len() == 0 && im.IsUniform(-100, 100, "."), "Empty map")
	fmt.Println("  [OK] Вырожденные интервалы (from >= to) и заливка фоном пустого дерева.")

	im.Assign(10, 20, "A")
	im.Assign(5, 10, "A")
	im.Assign(20, 25, "A")
	check()
	assertEq(toSlice(im.All()), []Interval[int, string]{{5, 25, "A"}}, "Merge sides")
	fmt.Println("  [OK] Автоматическая склейка примыкающих отрезков слева и справа.")

	im.Assign(10, 20, ".")
	check()
	assertEq(toSlice(im.All()), []Interval[int, string]{
		{5, 10, "A"}, {10, 20, "."}, {20, 25, "A"},
	}, "Split hole")
	fmt.Println("  [OK] Вырезание фоном посередине (расщепление на два отрезка).")

	im.Assign(10, 20, "B")
	im.Assign(20, 30, "C")
	im.Assign(8, 22, "X")
	check()
	assertEq(toSlice(im.QueryRange(0, 35)), []Interval[int, string]{
		{0, 5, "."}, {5, 8, "A"}, {8, 22, "X"}, {22, 30, "C"}, {30, 35, "."},
	}, "QueryRange")

	val, l, r, hasL, hasR := im.IntervalAt(15)
	assert(val == "X" && hasL && l == 8 && hasR && r == 22, "IntervalAt")

	im.Assign(-100, 100, ".")
	check()
	assert(im.m.Len() == 0, "Full erase")
	fmt.Println("  [OK] Перекрытие нескольких границ, QueryRange, IntervalAt и полная очистка.")
}

// ============================================================================
// 3. ВИЗУАЛИЗАЦИЯ И ДЕМОНСТРАЦИЯ РАСПИСАНИЯ ПЕРЕГОВОРКИ
// ============================================================================

const (
	rst   = "\x1b[0m"
	dimC  = "\x1b[2m"
	boldC = "\x1b[1m"
	grn   = "\x1b[32m"
	red   = "\x1b[31m"
	cyan  = "\x1b[36m"
)

const free = "свободно"

const (
	dayFrom = 8
	dayTo   = 20
	slots   = dayTo - dayFrom
)

var people = []string{"Анна", "Борис", "Вера"}

var palette = map[string]string{
	"Анна":  "\x1b[42m\x1b[30m",
	"Борис": "\x1b[44m\x1b[97m",
	"Вера":  "\x1b[43m\x1b[30m",
}

func pause() { time.Sleep(1500 * time.Millisecond) }

func cell(v string) string {
	if v == free {
		return dimC + "···" + rst
	}
	r, _ := utf8.DecodeRuneInString(v)
	return palette[v] + " " + string(r) + " " + rst
}

func strip(at func(p int) string) string {
	var b strings.Builder
	for p := 0; p < slots; p++ {
		b.WriteString(cell(at(p)))
	}
	return b.String()
}

func ruler() string {
	var b strings.Builder
	for h := dayFrom; h < dayTo; h += 2 {
		fmt.Fprintf(&b, "%-6d", h)
	}
	return b.String()
}

func marks(im *IntervalMap[int, string]) string {
	var b strings.Builder
	for p := 0; p < slots; p++ {
		if _, ok := im.m.Get(dayFrom + p); ok {
			b.WriteString(cyan + "▲" + rst + "  ")
		} else {
			b.WriteString("   ")
		}
	}
	return b.String()
}

func nodeList(im *IntervalMap[int, string]) string {
	var parts []string
	for k, v := range im.m.All() {
		parts = append(parts, fmt.Sprintf("%d→%s", k, v))
	}
	if len(parts) == 0 {
		return "(дерево пустое)"
	}
	return strings.Join(parts, " ")
}

func showState(im *IntervalMap[int, string]) {
	fmt.Println("  " + ruler())
	fmt.Println("  " + strip(func(p int) string { return im.At(dayFrom + p) }))
	fmt.Println("  " + marks(im))
	fmt.Printf("  %sузлов в дереве: %d (часов на оси: %d)%s\n", dimC, im.m.Len(), slots, rst)
	fmt.Println("  " + dimC + nodeList(im) + rst)
}

func demoRoom() {
	fmt.Println(boldC + "═══ 1. Расписание переговорки ═══" + rst)
	fmt.Println("Задача: помнить, кто занял комнату в")
	fmt.Println("каждый час, и менять расписание кусками.")
	fmt.Println("Хитрость: дерево хранит только")
	fmt.Println("ТОЧКИ СМЕНЫ, а не каждый час.")
	fmt.Println()
	var lg strings.Builder
	for _, p := range people {
		lg.WriteString(cell(p) + " " + p + "  ")
	}
	fmt.Println("  " + lg.String())
	fmt.Println("  " + cell(free) + " свободно   " + cyan + "▲" + rst + " точка смены")
	fmt.Println()

	im := makeIntervalMap[int, string](free)
	fmt.Println(boldC + "▶ Сначала пусто" + rst)
	showState(im)
	pause()

	steps := []struct {
		text     string
		from, to int
		who      string
	}{
		{"Анна бронирует 9–12", 9, 12, "Анна"},
		{"Борис бронирует 14–18", 14, 18, "Борис"},
		{"Анна продлевает 12–14: блоки склеились,\n  узел 12 исчез", 12, 14, "Анна"},
		{"Вера перебивает 11–16: подрезала\n  и Анну, и Бориса", 11, 16, "Вера"},
		{"Отмена 13–15: в блоке Веры дыра", 13, 15, free},
		{"Всё снято 8–20: дерево схлопнулось в 0", dayFrom, dayTo, free},
	}
	for _, s := range steps {
		fmt.Printf("\n%s▶ %s%s\n", boldC, s.text, rst)
		fmt.Printf("  %sAssign[%d,%d) = %s%s\n", dimC, s.from, s.to, s.who, rst)
		im.Assign(s.from, s.to, s.who)
		showState(im)
		pause()
	}
}

// ============================================================================
// 4. ПРОВЕРКА ПО ЭТАЛОНУ (СРАВНЕНИЕ С ОБЫЧНЫМ МАССИВОМ)
// ============================================================================

func demoCheck() {
	fmt.Println(boldC + "\n═══ 2. Проверка по эталону ═══" + rst)
	fmt.Println("Рядом с деревом — обычный массив: по")
	fmt.Println("значению на каждый час. Каждую операцию")
	fmt.Println("делаем с обоими и сравниваем.")
	fmt.Println()

	vals := []string{free, "Анна", "Борис", "Вера"}
	rng := rand.New(rand.NewSource(1))
	im := makeIntervalMap[int, string](free)
	oracle := make([]string, slots)
	for p := range oracle {
		oracle[p] = free
	}

	draw := func(first bool, op string) bool {
		same := true
		for p := 0; p < slots; p++ {
			if im.At(dayFrom+p) != oracle[p] {
				same = false
			}
		}
		status := grn + "✔ совпало" + rst
		if err := im.validate(); err != nil {
			same = false
			status = red + "✘ " + err.Error() + rst
		} else if !same {
			status = red + "✘ РАСХОЖДЕНИЕ" + rst
		}
		lines := []string{
			op,
			dimC + "  дерево (только точки смены):" + rst,
			"  " + strip(func(p int) string { return im.At(dayFrom + p) }),
			dimC + "  эталон (массив, все часы):" + rst,
			"  " + strip(func(p int) string { return oracle[p] }),
			"  " + status + fmt.Sprintf(" · узлов: %d", im.m.Len()),
		}
		if !first {
			fmt.Printf("\x1b[%dA", len(lines))
		}
		for _, l := range lines {
			fmt.Print("\r\x1b[2K" + l + "\n")
		}
		return same
	}

	for i := 1; i <= 25; i++ {
		a := rng.Intn(slots)
		b := a + 1 + rng.Intn(slots-a)
		who := vals[rng.Intn(len(vals))]
		im.Assign(dayFrom+a, dayFrom+b, who)
		for p := a; p < b; p++ {
			oracle[p] = who
		}
		op := fmt.Sprintf("  операция %d: %s → [%d,%d)", i, who, dayFrom+a, dayFrom+b)
		if !draw(i == 1, op) {
			os.Exit(1)
		}
		time.Sleep(350 * time.Millisecond)
	}

	fmt.Println()
	fmt.Println("Теперь то же самое, но без показа:")
	const ops, size = 20000, 300
	big := makeIntervalMap[int, int](0)
	ref := make([]int, size)
	for i := 1; i <= ops; i++ {
		a := rng.Intn(size - 1)
		b := a + 1 + rng.Intn(size-a)
		v := rng.Intn(5)
		big.Assign(a, b, v)
		for p := a; p < b; p++ {
			ref[p] = v
		}
		if err := big.validate(); err != nil {
			fmt.Printf("\n%s✘ шаг %d: %v%s\n", red, i, err, rst)
			os.Exit(1)
		}
		for p := 0; p < size; p++ {
			if big.At(p) != ref[p] {
				fmt.Printf("\n%s✘ шаг %d: точка %d: дерево=%d, эталон=%d%s\n", red, i, p, big.At(p), ref[p], rst)
				os.Exit(1)
			}
		}
		if i%400 == 0 {
			f := 20 * i / ops
			fmt.Printf("\r  [%s%s] %3d%%", strings.Repeat("█", f), strings.Repeat("░", 20-f), 100*i/ops)
		}
	}
	fmt.Printf("\n  %s✔ %d операций, %d сверок точек — всё совпало%s\n\n", grn, ops, ops*size, rst)

	fmt.Println("Крайние случаи и Set/Map:")
	testCorrectness()
	pause()
}

// ============================================================================
// 5. СТРЕСС-ТЕСТ ПРОИЗВОДИТЕЛЬНОСТИ И ВЫСОТЫ ДЕРЕВА
// ============================================================================

func fmtDur(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%dмкс", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%dмс", d.Milliseconds())
	default:
		return fmt.Sprintf("%.1fс", d.Seconds())
	}
}

func idealHeight(nodes int) int {
	return int(math.Ceil(math.Log2(float64(nodes + 1))))
}

func demoLoad() {
	const (
		totalSec = 26
		span     = 1_000_000_000
	)

	fmt.Printf("%s\n═══ 3. Стресс-тест (%s) ═══%s\n", boldC, treeName(), rst)
	fmt.Println("Сыплем в IntervalMap случайные отрезки")
	fmt.Println("(случайное место, длина 1..1000) до")
	fmt.Println("26 секунд и смотрим, как растут")
	fmt.Println("дерево и как меняется скорость.")
	fmt.Println()
	fmt.Println(dimC + "  с  отрезков     узлов  выс.  скорость" + rst)

	rng := rand.New(rand.NewSource(7))
	im := makeIntervalMap[int, int](0)

	var (
		workTime, prevWork          time.Duration
		sec, ops, lastOps           int
		firstRate, lastRate, halfAt int
		wallStart                   = time.Now()
		turnOn                      = wallStart
	)

	for sec < totalSec {
		if ops&255 == 0 {
			now := time.Now()
			workTime += now.Sub(turnOn)

			if now.Sub(wallStart) >= time.Duration(sec+1)*time.Second {
				sec++
				pureDt := (workTime - prevWork).Seconds()
				if pureDt > 0 {
					lastRate = int(float64(ops-lastOps) / pureDt)
				}
				if firstRate == 0 {
					firstRate = lastRate
				}
				if halfAt == 0 && lastRate*2 < firstRate {
					halfAt = ops
				}

				fmt.Printf("%3d %9d %9d %5d %7d/с\n",
					sec, ops, im.m.Len(), im.m.height(), lastRate)

				prevWork = workTime
				lastOps = ops
			}

			turnOn = time.Now()
		}

		a := rng.Intn(span)
		im.Assign(a, a+1+rng.Intn(1000), 1+rng.Intn(9))
		ops++
	}
	wallTotal := time.Since(wallStart)

	if err := im.validate(); err != nil {
		fmt.Printf("%s✘ дерево испорчено: %v%s\n", red, err, rst)
		os.Exit(1)
	}

	nodes, h := im.m.Len(), im.m.height()
	fmt.Println()
	fmt.Printf("%sИтог (%s)%s: %d отрезков за %s\n (чистое время вставки/удаления: %s)\n",
		boldC, treeName(), rst, ops, fmtDur(wallTotal), fmtDur(workTime))
	fmt.Printf("  узлов в дереве: %d\n", nodes)
	fmt.Printf("  высота: %d, у сбалансированного\n  было бы ≈ %d\n", h, idealHeight(nodes))
	if firstRate > 0 && lastRate > 0 {
		fmt.Printf("  скорость: %d → %d/с (в %.1f раза)\n", firstRate, lastRate, float64(firstRate)/float64(lastRate))
	}
	if halfAt > 0 {
		fmt.Printf("  скорость упала вдвое на ~%d отрезках\n", halfAt)
	}
	fmt.Printf("  %s✔ структура дерева после теста цела%s\n", grn, rst)
}

func main() {
	if slices.ContainsFunc(os.Args[1:], func(arg string) bool { return strings.EqualFold(arg, "bst") }) {
		useBST = true
	}

	fmt.Printf("%sИспользуемое дерево:%s %s%s%s\n\n", boldC, rst, cyan, treeName(), rst)
	demoRoom()
	demoCheck()
	demoLoad()
}
