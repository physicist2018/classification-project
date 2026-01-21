package main

import (
	"classification-project/internal/interface/reader"
	"classification-project/internal/models"
	"classification-project/pkg/math/statistics"
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

func main() {

	var N [models.Total]*models.Table
	for i := range models.TotalCv {
		N[i] = reader.ReadTableOrPanic(fmt.Sprintf("%s.txt", models.ClassificationName[i]))
	}

	rows, cols := N[models.Dust].Rows, N[models.Dust].Columns
	A := mat.NewDense(rows, cols, N[models.Dust].Data)
	B := mat.NewDense(rows, cols, N[models.Urban].Data)

	fmt.Print("Матрица D: ")
	fa := mat.Formatted(A, mat.Prefix("           "), mat.Squeeze())
	fmt.Printf("%.2f\n", fa)

	fmt.Print("\nМатрица U: ")
	fb := mat.Formatted(B, mat.Prefix("           "), mat.Squeeze())
	fmt.Printf("%.2f\n", fb)

	// Поиск максимальной области с минимальной корреляцией
	minSize := 4
	r1, c1, r2, c2, corr, err := statistics.FindMaxAreaMinCorrelation(A, B, minSize)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	fmt.Printf("\nРезультат поиска (минимальный размер: %dx%d):\n", minSize, minSize)
	fmt.Printf("Область: [%d:%d, %d:%d] (размер: %dx%d = %d элементов)\n",
		r1, r2, c1, c2,
		r2-r1+1, c2-c1+1, (r2-r1+1)*(c2-c1+1))
	fmt.Printf("Коэффициент корреляции: %.6f (|corr| = %.6f)\n", corr, math.Abs(corr))

	// Пример для квадратной области
	fmt.Println("\n--- Поиск квадратной области ---")
	i, j, size, corrSq, err := statistics.FindMaxSquareMinCorrelation(A, B, 2)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	fmt.Printf("Квадратная область: [%d:%d, %d:%d] (размер: %dx%d)\n",
		i, i+size-1, j, j+size-1, size, size)
	fmt.Printf("Коэффициент корреляции: %.6f\n", corrSq)

	// Демонстрация корреляции для всей матрицы
	fullCorr, _ := statistics.Corr2Submatrix(A, B, 0, 0, rows-1, cols-1)
	fmt.Printf("\nКорреляция для всей матрицы: %.6f\n", fullCorr)
}
