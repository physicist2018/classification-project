package statistics

import (
	"fmt"
	"math"
	"gonum.org/v1/gonum/mat"
)

// Corr2 вычисляет коэффициент корреляции между двумя матрицами одинакового размера.
// Возвращает значение корреляции и ошибку, если матрицы имеют разные размеры.
func Corr2(A, B *mat.Dense) (float64, error) {
	// Проверяем, что матрицы имеют одинаковые размеры
	ra, ca := A.Dims()
	rb, cb := B.Dims()
	if ra != rb || ca != cb {
		return 0, fmt.Errorf("матрицы должны иметь одинаковые размеры: A(%dx%d), B(%dx%d)", ra, ca, rb, cb)
	}

	// Получаем плоские представления матриц
	aFlat := A.RawMatrix().Data
	bFlat := B.RawMatrix().Data
	n := float64(len(aFlat))

	// Вычисляем средние значения
	var sumA, sumB float64
	for i := 0; i < len(aFlat); i++ {
		sumA += aFlat[i]
		sumB += bFlat[i]
	}
	meanA := sumA / n
	meanB := sumB / n

	// Вычисляем числитель и знаменатель для формулы корреляции
	var numerator float64
	var sumSqA, sumSqB float64

	for i := 0; i < len(aFlat); i++ {
		diffA := aFlat[i] - meanA
		diffB := bFlat[i] - meanB

		numerator += diffA * diffB
		sumSqA += diffA * diffA
		sumSqB += diffB * diffB
	}

	// Проверяем случай, когда знаменатель равен нулю
	if sumSqA == 0 || sumSqB == 0 {
		if sumSqA == 0 && sumSqB == 0 {
			// Если оба набора данных постоянны, корреляция считается равной 1
			return 1.0, nil
		}
		// Если один из наборов данных постоянный, корреляция не определена
		return 0, fmt.Errorf("один из наборов данных постоянный, корреляция не определена")
	}

	// Вычисляем коэффициент корреляции
	correlation := numerator / math.Sqrt(sumSqA*sumSqB)
	return correlation, nil
}

// Corr2WithNaNHandling - версия с обработкой NaN значений
func Corr2WithNaNHandling(A, B *mat.Dense) (float64, error) {
	ra, ca := A.Dims()
	rb, cb := B.Dims()
	if ra != rb || ca != cb {
		return 0, fmt.Errorf("матрицы должны иметь одинаковые размеры: A(%dx%d), B(%dx%d)", ra, ca, rb, cb)
	}

	aFlat := A.RawMatrix().Data
	bFlat := B.RawMatrix().Data

	// Собираем пары значений без NaN
	var validPairs [][2]float64
	for i := 0; i < len(aFlat); i++ {
		a, b := aFlat[i], bFlat[i]
		if !math.IsNaN(a) && !math.IsNaN(b) {
			validPairs = append(validPairs, [2]float64{a, b})
		}
	}

	if len(validPairs) == 0 {
		return math.NaN(), fmt.Errorf("нет валидных пар значений для расчета корреляции")
	}

	n := float64(len(validPairs))

	// Вычисляем средние
	var sumA, sumB float64
	for _, pair := range validPairs {
		sumA += pair[0]
		sumB += pair[1]
	}
	meanA := sumA / n
	meanB := sumB / n

	// Вычисляем корреляцию
	var numerator float64
	var sumSqA, sumSqB float64

	for _, pair := range validPairs {
		diffA := pair[0] - meanA
		diffB := pair[1] - meanB

		numerator += diffA * diffB
		sumSqA += diffA * diffA
		sumSqB += diffB * diffB
	}

	if sumSqA == 0 || sumSqB == 0 {
		if sumSqA == 0 && sumSqB == 0 {
			return 1.0, nil
		}
		return 0, fmt.Errorf("один из наборов данных постоянный")
	}

	return numerator / math.Sqrt(sumSqA*sumSqB), nil
}
