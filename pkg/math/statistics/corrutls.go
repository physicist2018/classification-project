package statistics

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

// Corr2Submatrix вычисляет корреляцию для подматриц
func Corr2Submatrix(A, B *mat.Dense, r1, c1, r2, c2 int) (float64, error) {
	rows := r2 - r1 + 1
	cols := c2 - c1 + 1
	n := rows * cols

	if n <= 1 {
		return 0, fmt.Errorf("подматрица слишком мала")
	}

	// Собираем элементы подматриц
	var sumA, sumB, sumAA, sumBB, sumAB float64

	for i := r1; i <= r2; i++ {
		for j := c1; j <= c2; j++ {
			a := A.At(i, j)
			b := B.At(i, j)
			sumA += a
			sumB += b
			sumAA += a * a
			sumBB += b * b
			sumAB += a * b
		}
	}

	nFloat := float64(n)
	cov := sumAB/nFloat - (sumA/nFloat)*(sumB/nFloat)
	stdA := math.Sqrt(sumAA/nFloat - (sumA/nFloat)*(sumA/nFloat))
	stdB := math.Sqrt(sumBB/nFloat - (sumB/nFloat)*(sumB/nFloat))

	if stdA == 0 || stdB == 0 {
		if stdA == 0 && stdB == 0 {
			return 1.0, nil
		}
		return 0, nil
	}

	return cov / (stdA * stdB), nil
}

// FindMaxAreaMinCorrelation находит максимальную область с минимальной корреляцией
func FindMaxAreaMinCorrelation(A, B *mat.Dense, minSize int) (int, int, int, int, float64, error) {
	rows, cols := A.Dims()

	if minSize < 1 {
		minSize = 1
	}

	var bestR1, bestC1, bestR2, bestC2 int
	bestCorr := math.MaxFloat64 // ищем минимальную по модулю
	bestSize := -1

	// Перебираем все возможные прямоугольные области
	for r1 := 0; r1 < rows; r1++ {
		for c1 := 0; c1 < cols; c1++ {
			// Максимально возможные размеры из начальной точки
			maxRows := rows - r1
			maxCols := cols - c1

			// Перебираем размеры области
			for h := minSize; h <= maxRows; h++ {
				for w := minSize; w <= maxCols; w++ {
					r2 := r1 + h - 1
					c2 := c1 + w - 1

					// Пропускаем слишком маленькие области
					if h*w < minSize {
						continue
					}

					corr, err := Corr2Submatrix(A, B, r1, c1, r2, c2)
					if err != nil {
						continue
					}

					absCorr := math.Abs(corr)
					area := h * w

					// Критерий: сначала минимальный |corr|, затем максимальная площадь
					if (absCorr < bestCorr) ||
						(math.Abs(absCorr-bestCorr) < 1e-10 && area > bestSize) {
						bestCorr = absCorr
						bestSize = area
						bestR1, bestC1 = r1, c1
						bestR2, bestC2 = r2, c2
					}
				}
			}
		}
	}

	if bestSize == -1 {
		return -1, -1, -1, -1, 0, fmt.Errorf("не найдено подходящих областей")
	}

	return bestR1, bestC1, bestR2, bestC2, bestCorr, nil
}

// FindMaxAreaMinCorrelationOptimized - оптимизированная версия с предвычислениями
func FindMaxAreaMinCorrelationOptimized(A, B *mat.Dense, minSize int) (int, int, int, int, float64, error) {
	rows, cols := A.Dims()

	if minSize < 1 {
		minSize = 1
	}

	// Предвычисление префиксных сумм для быстрого расчета статистик
	prefixA := mat.NewDense(rows+1, cols+1, nil)
	prefixB := mat.NewDense(rows+1, cols+1, nil)
	prefixAA := mat.NewDense(rows+1, cols+1, nil)
	prefixBB := mat.NewDense(rows+1, cols+1, nil)
	prefixAB := mat.NewDense(rows+1, cols+1, nil)

	// Заполняем префиксные суммы
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			a := A.At(i, j)
			b := B.At(i, j)
			a2 := a * a
			b2 := b * b
			ab := a * b

			prefixA.Set(i+1, j+1, a+prefixA.At(i+1, j)+prefixA.At(i, j+1)-prefixA.At(i, j))
			prefixB.Set(i+1, j+1, b+prefixB.At(i+1, j)+prefixB.At(i, j+1)-prefixB.At(i, j))
			prefixAA.Set(i+1, j+1, a2+prefixAA.At(i+1, j)+prefixAA.At(i, j+1)-prefixAA.At(i, j))
			prefixBB.Set(i+1, j+1, b2+prefixBB.At(i+1, j)+prefixBB.At(i, j+1)-prefixBB.At(i, j))
			prefixAB.Set(i+1, j+1, ab+prefixAB.At(i+1, j)+prefixAB.At(i, j+1)-prefixAB.At(i, j))
		}
	}

	// Функция быстрого получения суммы в прямоугольнике
	getSum := func(prefix *mat.Dense, r1, c1, r2, c2 int) float64 {
		return prefix.At(r2+1, c2+1) - prefix.At(r1, c2+1) -
			prefix.At(r2+1, c1) + prefix.At(r1, c1)
	}

	var bestR1, bestC1, bestR2, bestC2 int
	bestCorr := math.MaxFloat64
	bestSize := -1

	// Перебираем все возможные прямоугольные области
	for r1 := 0; r1 < rows; r1++ {
		for r2 := r1 + minSize - 1; r2 < rows; r2++ {
			height := r2 - r1 + 1

			for c1 := 0; c1 < cols; c1++ {
				for c2 := c1 + minSize - 1; c2 < cols; c2++ {
					width := c2 - c1 + 1
					area := height * width

					if area < minSize {
						continue
					}

					// Быстро вычисляем статистики через префиксные суммы
					sumA := getSum(prefixA, r1, c1, r2, c2)
					sumB := getSum(prefixB, r1, c1, r2, c2)
					sumAA := getSum(prefixAA, r1, c1, r2, c2)
					sumBB := getSum(prefixBB, r1, c1, r2, c2)
					sumAB := getSum(prefixAB, r1, c1, r2, c2)

					nFloat := float64(area)
					cov := sumAB/nFloat - (sumA/nFloat)*(sumB/nFloat)
					stdA := math.Sqrt(sumAA/nFloat - (sumA/nFloat)*(sumA/nFloat))
					stdB := math.Sqrt(sumBB/nFloat - (sumB/nFloat)*(sumB/nFloat))

					if stdA == 0 || stdB == 0 {
						continue
					}

					corr := cov / (stdA * stdB)
					absCorr := math.Abs(corr)

					// Обновляем лучший результат
					if (absCorr < bestCorr) ||
						(math.Abs(absCorr-bestCorr) < 1e-10 && area > bestSize) {
						bestCorr = absCorr
						bestSize = area
						bestR1, bestC1, bestR2, bestC2 = r1, c1, r2, c2
					}
				}
			}
		}
	}

	if bestSize == -1 {
		return -1, -1, -1, -1, 0, fmt.Errorf("не найдено подходящих областей")
	}

	return bestR1, bestC1, bestR2, bestC2, bestCorr, nil
}

// FindMaxSquareMinCorrelation - поиск квадратной области (упрощенная задача)
func FindMaxSquareMinCorrelation(A, B *mat.Dense, minSize int) (int, int, int, float64, error) {
	rows, cols := A.Dims()

	var bestI, bestJ, bestSize int
	bestCorr := math.MaxFloat64

	// Перебираем все возможные квадратные области
	for size := minSize; size <= min(rows, cols); size++ {
		for i := 0; i <= rows-size; i++ {
			for j := 0; j <= cols-size; j++ {
				corr, err := Corr2Submatrix(A, B, i, j, i+size-1, j+size-1)
				if err != nil {
					continue
				}

				absCorr := math.Abs(corr)
				if absCorr < bestCorr {
					bestCorr = absCorr
					bestI, bestJ = i, j
					bestSize = size
				}
			}
		}
	}

	if bestSize == 0 {
		return -1, -1, -1, 0, fmt.Errorf("не найдено подходящих квадратных областей")
	}

	return bestI, bestJ, bestSize, bestCorr, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
