package main

import (
	"fmt"
	"runtime"
	"time"
)

// Порог переключения на сортировку вставками для малых подмассивов
const mergeThresh = 128

// Гибридная сортировка слиянием с чередованием буфера
func mergeSort(arr []int) {
	if len(arr) <= 1 {
		return
	}
	mergeSortR(arr, nil, false)
}

// Рекурсивно делит массив и сливает части, чередуя основной слайс и буфер, чтобы избежать лишнего копирования
func mergeSortR(orig, buf []int, toBuf bool) {
	if len(orig) <= mergeThresh && (!toBuf || buf == nil) {
		insertionSort(orig)
		return
	}

	// Ленивое выделение памяти: создаём буфер только при первом реальном слиянии
	if buf == nil {
		buf = make([]int, len(orig))
		toBuf = false
	}

	mid := len(orig) / 2
	origLeft, origRight := orig[:mid], orig[mid:]
	bufLeft, bufRight := buf[:mid], buf[mid:]

	mergeSortR(origLeft, bufLeft, !toBuf)
	mergeSortR(origRight, bufRight, !toBuf)

	if toBuf {
		merge(origLeft, origRight, buf)
	} else {
		merge(bufLeft, bufRight, orig)
	}
}

// Сливает два отсортированных слайса в целевой массив dst
func merge(left, right, dst []int) {
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			dst[i+j] = left[i]
			i++
		} else {
			dst[i+j] = right[j]
			j++
		}
	}
	copy(dst[i+j:], left[i:])
	copy(dst[i+j:], right[j:])
}

// Классическая быстрая сортировка Хоара с выбором опорного элемента по центру
func quickSort(arr []int) {
	if len(arr) < 2 {
		return
	}

	left, right := 0, len(arr)-1
	pivot := arr[len(arr)/2]

	for left <= right {
		for arr[left] < pivot {
			left++
		}
		for arr[right] > pivot {
			right--
		}
		if left <= right {
			arr[left], arr[right] = arr[right], arr[left]
			left++
			right--
		}
	}

	quickSort(arr[:right+1])
	quickSort(arr[left:])
}

// Пирамидальная сортировка на максимальной двоичной куче
func heapSort(arr []int) {
	n := len(arr)
	for i := n/2 - 1; i >= 0; i-- {
		heapify(arr, n, i)
	}
	for i := n - 1; i > 0; i-- {
		arr[0], arr[i] = arr[i], arr[0]
		heapify(arr, i, 0)
	}
}

// Просеивает элемент вниз, восстанавливая свойства максимальной кучи
func heapify(arr []int, size, root int) {
	largest := root
	left := 2*root + 1
	right := 2*root + 2

	if left < size && arr[left] > arr[largest] {
		largest = left
	}
	if right < size && arr[right] > arr[largest] {
		largest = right
	}
	if largest != root {
		arr[root], arr[largest] = arr[largest], arr[root]
		heapify(arr, size, largest)
	}
}

// Сортировка вставками
func insertionSort(arr []int) {
	for i := 1; i < len(arr); i++ {
		temp := arr[i]
		j := i - 1
		for j >= 0 && arr[j] > temp {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = temp
	}
}

// Выводит в консоль наглядную столбчатую диаграмму значений небольшого массива
func drawArray(arr []int) {
	if len(arr) == 0 {
		return
	}

	min, max := arr[0], arr[0]
	for _, v := range arr {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}
	if max == min {
		max++
	}

	maxHeight := 6
	for h := maxHeight; h >= 1; h-- {
		fmt.Print(" │ ")
		for _, v := range arr {
			normalized := 1 + int(float64(v-min)/float64(max-min)*float64(maxHeight-1))
			if normalized >= h {
				fmt.Print("█  ")
			} else {
				fmt.Print("   ")
			}
		}
		fmt.Println()
	}

	fmt.Print(" └─")
	for i := 0; i < len(arr); i++ {
		fmt.Print("───")
	}
	fmt.Println("▶")

	fmt.Print("  ")
	for _, v := range arr {
		fmt.Printf("%2d ", v)
	}
	fmt.Println()
}

// Проверяет неубывающий порядок элементов после сортировки
func checkSorted(arr []int) string {
	for i := 0; i < len(arr)-1; i++ {
		if arr[i] > arr[i+1] {
			return "❌ СЛОМАЛСЯ"
		}
	}
	return "🟢 ВЕРНО"
}

// Форматирует время выполнения в секунды или миллисекунды с точностью до 3 знаков
func formatDuration(d time.Duration) string {
	if d >= time.Second {
		return fmt.Sprintf("%.3f s", d.Seconds())
	}
	ms := float64(d.Nanoseconds()) / 1e6
	return fmt.Sprintf("%.3f ms", ms)
}

// Быстрый линейный конгруэнтный генератор псевдослучайных чисел
func randomInt(seed *uint64, max int) int {
	*seed = *seed*6364136223846793005 + 1442695040888963407
	return int(*seed % uint64(max))
}

// Конструирует худший входной массив для quickSort: на каждом шаге в центр текущего
// отрезка [i:] подставляется минимальное число, ломая баланс рекурсии до O(N²).
func generateAntiQuickSort(size int) []int {
	arr := make([]int, size)
	pos := make([]int, size)
	for i := range pos {
		pos[i] = i
	}

	for i := 0; i < size; i++ {
		mid := i + (size-i)/2               // Центр оставшегося несортированного участка
		arr[pos[mid]] = i                   // Кладём в этот будущий pivot минимальное значение
		pos[i], pos[mid] = pos[mid], pos[i] // Повторяем обмен из схемы Хоара
	}
	return arr
}

// Хранит результаты замеров времени и памяти
type Metrics struct {
	Name       string
	Duration   time.Duration
	AllocMemKB uint64
}

// Замеряет время и аллокации памяти. При copyData == true работает с копией, сохраняя исходный массив.
func profile(name string, sortFunc func([]int), source []int, copyData bool) Metrics {
	data := source
	if copyData {
		data = make([]int, len(source))
		copy(data, source)
	}

	fmt.Printf("-> Запускается %s...\n", name)

	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)

	start := time.Now()
	sortFunc(data)
	duration := time.Since(start)

	runtime.ReadMemStats(&mAfter)
	allocKB := (mAfter.TotalAlloc - mBefore.TotalAlloc) / 1024

	return Metrics{Name: name, Duration: duration, AllocMemKB: allocKB}
}

// 1. Проверка корректности всех алгоритмов на небольшом массиве с отрицательными числами и дубликатами
func testSanity() {
	perverted := []int{9, -9, 9, 0, -5, 3, 3, -5, 0, 7, -9}

	fmt.Println("=== 1. ПРОВЕРКА КОРРЕКТНОСТИ (11 ЭЛ.) ===")
	fmt.Println("ДО СОРТИРОВКИ:")
	drawArray(perverted)

	pMerge := make([]int, len(perverted))
	pQuick := make([]int, len(perverted))
	pHeap := make([]int, len(perverted))
	pIns := make([]int, len(perverted))
	copy(pMerge, perverted)
	copy(pQuick, perverted)
	copy(pHeap, perverted)
	copy(pIns, perverted)

	mergeSort(pMerge)
	quickSort(pQuick)
	heapSort(pHeap)
	insertionSort(pIns)

	fmt.Println("\nПОСЛЕ СОРТИРОВКИ:")
	drawArray(pMerge)

	fmt.Println()
	fmt.Printf("Merge Sort:     [%s]\n", checkSorted(pMerge))
	fmt.Printf("Quick Sort:     [%s]\n", checkSorted(pQuick))
	fmt.Printf("Heap Sort:      [%s]\n", checkSorted(pHeap))
	fmt.Printf("Insertion Sort: [%s]\n", checkSorted(pIns))
}

// 2. Сравнение на средних объёмах: демонстрация квадратичного отставания Insertion Sort
func testCompareInsertion() {
	size := 210000
	fmt.Printf("\n=== 2. ГОНКА (%d ЭЛ.) ===\n", size)

	var seed uint64 = 42
	source := make([]int, size)
	for i := range source {
		source[i] = randomInt(&seed, 1000000)
	}

	m1 := profile("Merge Sort", mergeSort, source, true)
	m2 := profile("Quick Sort", quickSort, source, true)
	m3 := profile("Heap Sort", heapSort, source, true)
	m4 := profile("Insertion Sort", insertionSort, source, true)

	fmt.Println()
	fmt.Printf("%-14s | %-10s | %-8s\n", "Алгоритм", "Время", "Память")
	fmt.Println("--------------------------------------")
	fmt.Printf("%-14s | %-10s | %d KB\n", m1.Name, formatDuration(m1.Duration), m1.AllocMemKB)
	fmt.Printf("%-14s | %-10s | %d KB\n", m2.Name, formatDuration(m2.Duration), m2.AllocMemKB)
	fmt.Printf("%-14s | %-10s | %d KB\n", m3.Name, formatDuration(m3.Duration), m3.AllocMemKB)
	fmt.Printf("%-14s | %-10s | %d KB\n", m4.Name, formatDuration(m4.Duration), m4.AllocMemKB)
}

// 3. Проверка на большом масштабе: просадка Heap Sort из-за промахов кэша процессора
func testCompareHeap() {
	size := 17000000
	fmt.Printf("\n=== 3. БОЛЬШАЯ ГОНКА (%d ЭЛ.) ===\n", size)

	var seed uint64 = 42
	source := make([]int, size)
	for i := range source {
		source[i] = randomInt(&seed, 1000000)
	}

	b1 := profile("Merge Sort", mergeSort, source, true)
	b2 := profile("Quick Sort", quickSort, source, true)
	b3 := profile("Heap Sort", heapSort, source, true)

	fmt.Println()
	fmt.Printf("%-14s | %-10s | %-8s\n", "Алгоритм", "Время", "Память")
	fmt.Println("--------------------------------------")
	fmt.Printf("%-14s | %-10s | %d KB\n", b1.Name, formatDuration(b1.Duration), b1.AllocMemKB)
	fmt.Printf("%-14s | %-10s | %d KB\n", b2.Name, formatDuration(b2.Duration), b2.AllocMemKB)
	fmt.Printf("%-14s | %-10s | %d KB\n", b3.Name, formatDuration(b3.Duration), b3.AllocMemKB)
}

// 4. Финальное сравнение двух лидеров (Merge Sort и Quick Sort) на сверхбольшом массиве
func testCompareQuickest() {
	size := 65000000
	fmt.Printf("\n=== 4. ФИНАЛ (%d ЭЛ.) ===\n", size)

	var seed1 uint64 = 42
	data := make([]int, size)
	for i := range data {
		data[i] = randomInt(&seed1, 1000000)
	}
	f1 := profile("Merge Sort", mergeSort, data, false)

	var seed2 uint64 = 42
	for i := range data {
		data[i] = randomInt(&seed2, 1000000)
	}
	f2 := profile("Quick Sort", quickSort, data, false)

	fmt.Println()
	fmt.Printf("%-14s | %-10s | %-8s\n", "Алгоритм", "Время", "Память")
	fmt.Println("--------------------------------------")
	fmt.Printf("%-14s | %-10s | %d KB\n", f1.Name, formatDuration(f1.Duration), f1.AllocMemKB)
	fmt.Printf("%-14s | %-10s | %d KB\n", f2.Name, formatDuration(f2.Duration), f2.AllocMemKB)
}

// 5. Демонстрация уязвимости Quick Sort на специально подобранном худшем массиве
func testWinnerVulnerability() {
	size := 210000
	fmt.Printf("\n=== 5. УЯЗВИМОСТЬ ПОБЕДИТЕЛЯ (%d ЭЛ.) ===\n", size)
	fmt.Println("Искусственно созданный «извращённый» массив для Quick Sort: на каждом шаге в центр")
	fmt.Println("попадает минимальный элемент, из-за чего алгоритм вырождается в свой худший случай.")

	source := generateAntiQuickSort(size)

	w1 := profile("Merge Sort", mergeSort, source, true)
	w2 := profile("Quick Sort", quickSort, source, true)
	w3 := profile("Heap Sort", heapSort, source, true)
	w4 := profile("Insertion Sort", insertionSort, source, true)

	fmt.Println()
	fmt.Printf("%-14s | %-10s | %-8s\n", "Алгоритм", "Время", "Память")
	fmt.Println("--------------------------------------")
	fmt.Printf("%-14s | %-10s | %d KB\n", w1.Name, formatDuration(w1.Duration), w1.AllocMemKB)
	fmt.Printf("%-14s | %-10s | %d KB\n", w2.Name, formatDuration(w2.Duration), w2.AllocMemKB)
	fmt.Printf("%-14s | %-10s | %d KB\n", w3.Name, formatDuration(w3.Duration), w3.AllocMemKB)
	fmt.Printf("%-14s | %-10s | %d KB\n", w4.Name, formatDuration(w4.Duration), w4.AllocMemKB)
}

func main() {
	testSanity()
	testCompareInsertion()
	testCompareHeap()
	testCompareQuickest()
	testWinnerVulnerability()
}
