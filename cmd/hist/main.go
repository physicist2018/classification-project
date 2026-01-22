package main

import (
	"classification-project/internal/models"
	"classification-project/pkg/math/statistics"
	"fmt"
	"math"
)

// Пример использования
func main() {
	// Создаем тестовые данные
	solutions := []models.OutputSolution{
		{
			Cv:          []float64{1.0, 2.0, 3.0},
			Discrepancy: 0.1,
		},
		{
			Cv:          []float64{1.1, 2.1, 3.1},
			Discrepancy: 0.2,
		},
		{
			Cv:          []float64{0.9, 1.9, 2.9},
			Discrepancy: 0.15,
		},
		// Добавьте больше данных для реалистичной гистограммы
	}

	// Добавим больше данных для демонстрации
	for i := 0; i < 100; i++ {
		solutions = append(solutions, models.OutputSolution{
			Cv: []float64{
				1.0 + 0.5*math.Sin(float64(i)*0.1),
				2.0 + 0.3*math.Cos(float64(i)*0.2),
				3.0 + 0.2*math.Sin(float64(i)*0.3),
			},
			Discrepancy: 0.1 + 0.05*math.Abs(math.Sin(float64(i)*0.05)),
		})
	}

	// Простой расчет гистограмм
	fmt.Println("=== Простой расчет гистограмм ===")
	simpleResults := statistics.CalculateHistograms(solutions, 10)
	statistics.PrintAllHistograms(simpleResults, 50)

	// Вывод статистики для Discrepancy
	fmt.Println("\n=== Статистика для Discrepancy ===")
	stats := statistics.GetHistogramStatistics(simpleResults["Discrepancy"])
	for key, value := range stats {
		fmt.Printf("%s: %.4f\n", key, value)
	}

	// Экспорт в CSV
	fmt.Println("\n=== CSV экспорт для Discrepancy ===")
	csv := statistics.ExportHistogramToCSV(simpleResults["Discrepancy"])
	fmt.Println(csv[:200] + "...") // Показываем только начало

	// Продвинутый расчет с опциями
	fmt.Println("\n=== Продвинутый расчет гистограмм ===")
	options := statistics.AdvancedHistogramOptions{
		NumBins:           15,
		AutoBinWidth:      false,
		CustomRange:       false,
		IncludeStatistics: true,
		ExportFormat:      "text",
	}

	advancedResults := statistics.CalculateAdvancedHistograms(solutions, options)

	// Выводим только Discrepancy для примера
	statistics.PrintHistogram(advancedResults["Discrepancy"], 50)

	// Получаем статистику для Cv[0]
	if cv0Result, ok := advancedResults["Cv[0]"]; ok {
		fmt.Println("\n=== Статистика для Cv[0] ===")
		cvStats := statistics.GetHistogramStatistics(cv0Result)
		for key, value := range cvStats {
			fmt.Printf("%s: %.4f\n", key, value)
		}
	}
}
