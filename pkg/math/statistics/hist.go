package statistics

import (
	"classification-project/internal/models"
	"fmt"
	"math"
)

// HistogramBin представляет один бин гистограммы
type HistogramBin struct {
	LowerBound float64
	UpperBound float64
	Count      int
	Frequency  float64 // относительная частота
}

// HistogramResult содержит результаты гистограммы
type HistogramResult struct {
	Min        float64
	Max        float64
	BinWidth   float64
	TotalCount int
	Bins       []HistogramBin
	DataName   string // название данных (например, "Discrepancy", "Cv[0]")
}

// CalculateHistograms вычисляет гистограммы для всех данных
func CalculateHistograms(solutions []models.OutputSolution, numBins int) map[string]HistogramResult {
	results := make(map[string]HistogramResult)

	if len(solutions) == 0 {
		return results
	}

	// 1. Гистограмма для Discrepancy
	discrepancyData := make([]float64, len(solutions))
	for i, sol := range solutions {
		discrepancyData[i] = sol.Discrepancy
	}

	results["Discrepancy"] = calculateSingleHistogram(
		discrepancyData,
		numBins,
		"Discrepancy",
	)

	// 2. Гистограммы для каждого элемента Cv
	if len(solutions) > 0 && len(solutions[0].Cv) > 0 {
		cvLength := len(solutions[0].Cv)

		// Проверяем, что все Cv имеют одинаковую длину
		for i := 1; i < len(solutions); i++ {
			if len(solutions[i].Cv) != cvLength {
				panic("Все Cv должны иметь одинаковую длину")
			}
		}

		// Для каждого индекса Cv создаем отдельную гистограмму
		for cvIndex := 0; cvIndex < cvLength; cvIndex++ {
			idxName := models.ClassificationName[cvIndex]
			cvData := make([]float64, len(solutions))
			for i, sol := range solutions {
				cvData[i] = sol.Cv[cvIndex]
			}

			results[fmt.Sprintf("Cv[%s]", idxName)] = calculateSingleHistogram(
				cvData,
				numBins,
				fmt.Sprintf("Cv[%s]", idxName),
			)
		}
	}

	return results
}

// calculateSingleHistogram вычисляет гистограмму для одного набора данных
func calculateSingleHistogram(data []float64, numBins int, name string) HistogramResult {
	if len(data) == 0 {
		return HistogramResult{DataName: name}
	}

	// Находим минимум и максимум
	minVal, maxVal := findMinMax(data)

	// Добавляем небольшую эпсилон для обработки граничных значений
	epsilon := 1e-10

	// Вычисляем ширину бина
	binWidth := (maxVal - minVal + 2*epsilon) / float64(numBins)

	// Создаем бины
	bins := make([]HistogramBin, numBins)

	// Инициализируем бины
	for i := 0; i < numBins; i++ {
		lower := minVal + float64(i)*binWidth
		upper := minVal + float64(i+1)*binWidth

		// Для последнего бина включаем верхнюю границу
		if i == numBins-1 {
			upper += epsilon
		}

		bins[i] = HistogramBin{
			LowerBound: lower,
			UpperBound: upper,
			Count:      0,
		}
	}

	// Распределяем данные по бинам
	for _, value := range data {
		// Определяем индекс бина
		binIndex := int(math.Floor((value - minVal) / binWidth))

		// Обработка граничного случая, когда value == maxVal
		if binIndex >= numBins {
			binIndex = numBins - 1
		}
		if binIndex < 0 {
			binIndex = 0
		}

		bins[binIndex].Count++
	}

	// Вычисляем относительные частоты
	total := float64(len(data))
	for i := range bins {
		bins[i].Frequency = float64(bins[i].Count) / total
	}

	return HistogramResult{
		Min:        minVal,
		Max:        maxVal,
		BinWidth:   binWidth,
		TotalCount: len(data),
		Bins:       bins,
		DataName:   name,
	}
}

// findMinMax находит минимум и максимум в срезе
func findMinMax(data []float64) (min, max float64) {
	if len(data) == 0 {
		return 0, 0
	}

	min = data[0]
	max = data[0]

	for _, val := range data[1:] {
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
	}

	return min, max
}

// PrintHistogram выводит гистограмму в консоль
func PrintHistogram(result HistogramResult, maxBarWidth int) {
	fmt.Printf("\n=== Гистограмма для %s ===\n", result.DataName)
	fmt.Printf("Диапазон: [%.4f, %.4f]\n", result.Min, result.Max)
	fmt.Printf("Ширина бина: %.4f\n", result.BinWidth)
	fmt.Printf("Всего значений: %d\n\n", result.TotalCount)

	// Находим максимальное количество для масштабирования
	maxCount := 0
	for _, bin := range result.Bins {
		if bin.Count > maxCount {
			maxCount = bin.Count
		}
	}

	// Выводим каждый бин
	for i, bin := range result.Bins {
		// Масштабируем длину бара
		barLength := 0
		if maxCount > 0 {
			barLength = int(float64(bin.Count) / float64(maxCount) * float64(maxBarWidth))
		}

		// Создаем строку бара
		bar := ""
		for j := 0; j < barLength; j++ {
			bar += "█"
		}

		fmt.Printf("[%+7.4f - %+7.4f] %8d (%6.1f%%) %s\n",
			bin.LowerBound,
			bin.UpperBound,
			bin.Count,
			bin.Frequency*100,
			bar,
		)

		_ = i
		// // Добавляем пустую строку для разделения каждых 5 бинов
		// if (i+1)%5 == 0 && i != len(result.Bins)-1 {
		// 	fmt.Println()
		// }
	}
}

// PrintAllHistograms выводит все гистограммы
func PrintAllHistograms(results map[string]HistogramResult, maxBarWidth int) {
	for _, result := range results {
		PrintHistogram(result, maxBarWidth)
		fmt.Println()
	}
}

// GetHistogramStatistics возвращает статистику по гистограмме
func GetHistogramStatistics(result HistogramResult) map[string]float64 {
	stats := make(map[string]float64)

	// Вычисляем среднее значение (приближенно)
	//var sum float64
	var weightedSum float64

	for _, bin := range result.Bins {
		binCenter := (bin.LowerBound + bin.UpperBound) / 2
		weightedSum += binCenter * float64(bin.Count)
	}

	stats["mean"] = weightedSum / float64(result.TotalCount)

	// Находим медиану (приближенно)
	medianBin := 0
	cumulativeCount := 0
	target := result.TotalCount / 2

	for i, bin := range result.Bins {
		cumulativeCount += bin.Count
		if cumulativeCount >= target {
			medianBin = i
			break
		}
	}

	if medianBin < len(result.Bins) {
		stats["median"] = (result.Bins[medianBin].LowerBound + result.Bins[medianBin].UpperBound) / 2
	}

	// Мода (бин с максимальным количеством)
	maxCount := 0
	modeBin := 0
	for i, bin := range result.Bins {
		if bin.Count > maxCount {
			maxCount = bin.Count
			modeBin = i
		}
	}

	if modeBin < len(result.Bins) {
		stats["mode"] = (result.Bins[modeBin].LowerBound + result.Bins[modeBin].UpperBound) / 2
	}

	// Стандартное отклонение (приближенно)
	var sumSquaredDiff float64
	mean := stats["mean"]

	for _, bin := range result.Bins {
		binCenter := (bin.LowerBound + bin.UpperBound) / 2
		diff := binCenter - mean
		sumSquaredDiff += diff * diff * float64(bin.Count)
	}

	stats["stddev"] = math.Sqrt(sumSquaredDiff / float64(result.TotalCount))

	return stats
}

// ExportHistogramToCSV экспортирует гистограмму в CSV формат
func ExportHistogramToCSV(result HistogramResult) string {
	csv := fmt.Sprintf("Bin,LowerBound,UpperBound,Count,Frequency,FrequencyPercent\n")

	for i, bin := range result.Bins {
		csv += fmt.Sprintf("%d,%.6f,%.6f,%d,%.6f,%.2f\n",
			i+1,
			bin.LowerBound,
			bin.UpperBound,
			bin.Count,
			bin.Frequency,
			bin.Frequency*100,
		)
	}

	return csv
}

// AdvancedHistogramOptions опции для продвинутого расчета гистограмм
type AdvancedHistogramOptions struct {
	NumBins           int
	AutoBinWidth      bool    // Автоматически определять ширину бина по правилу Стёрджеса
	CustomRange       bool    // Использовать пользовательский диапазон
	CustomMin         float64 // Пользовательский минимум
	CustomMax         float64 // Пользовательский максимум
	IncludeStatistics bool    // Включать статистику
	ExportFormat      string  // Формат экспорта ("csv", "json", "text")
}

// CalculateAdvancedHistograms продвинутая версия с опциями
func CalculateAdvancedHistograms(
	solutions []models.OutputSolution,
	options AdvancedHistogramOptions,
) map[string]HistogramResult {

	results := make(map[string]HistogramResult)

	if len(solutions) == 0 {
		return results
	}

	// Определяем количество бинов, если включена автонастройка
	numBins := options.NumBins
	if options.AutoBinWidth && numBins <= 0 {
		// Правило Стёрджеса: k = 1 + log2(n)
		n := len(solutions)
		numBins = int(1 + math.Log2(float64(n)))
	}

	if numBins <= 0 {
		numBins = 10 // значение по умолчанию
	}

	// 1. Гистограмма для Discrepancy
	discrepancyData := extractDiscrepancyData(solutions)

	results["Discrepancy"] = calculateSingleHistogramWithOptions(
		discrepancyData,
		numBins,
		"Discrepancy",
		options,
	)

	// 2. Гистограммы для каждого элемента Cv
	if len(solutions) > 0 && len(solutions[0].Cv) > 0 {
		cvLength := len(solutions[0].Cv)

		for cvIndex := 0; cvIndex < cvLength; cvIndex++ {
			cvData := extractCvData(solutions, cvIndex)

			results[fmt.Sprintf("Cv[%d]", cvIndex)] = calculateSingleHistogramWithOptions(
				cvData,
				numBins,
				fmt.Sprintf("Cv[%d]", cvIndex),
				options,
			)
		}
	}

	return results
}

// extractDiscrepancyData извлекает данные Discrepancy
func extractDiscrepancyData(solutions []models.OutputSolution) []float64 {
	data := make([]float64, len(solutions))
	for i, sol := range solutions {
		data[i] = sol.Discrepancy
	}
	return data
}

// extractCvData извлекает данные для конкретного индекса Cv
func extractCvData(solutions []models.OutputSolution, index int) []float64 {
	data := make([]float64, len(solutions))
	for i, sol := range solutions {
		if index < len(sol.Cv) {
			data[i] = sol.Cv[index]
		} else {
			data[i] = math.NaN()
		}
	}
	return data
}

// calculateSingleHistogramWithOptions вычисляет гистограмму с учетом опций
func calculateSingleHistogramWithOptions(
	data []float64,
	numBins int,
	name string,
	options AdvancedHistogramOptions,
) HistogramResult {

	// Фильтруем NaN значения
	filteredData := make([]float64, 0, len(data))
	for _, val := range data {
		if !math.IsNaN(val) {
			filteredData = append(filteredData, val)
		}
	}

	if len(filteredData) == 0 {
		return HistogramResult{DataName: name}
	}

	// Определяем диапазон
	var minVal, maxVal float64

	if options.CustomRange {
		minVal = options.CustomMin
		maxVal = options.CustomMax
	} else {
		minVal, maxVal = findMinMax(filteredData)
	}

	// Добавляем небольшую эпсилон для обработки граничных значений
	epsilon := 1e-10

	// Вычисляем ширину бина
	binWidth := (maxVal - minVal + 2*epsilon) / float64(numBins)

	// Создаем бины
	bins := make([]HistogramBin, numBins)

	// Инициализируем бины
	for i := 0; i < numBins; i++ {
		lower := minVal + float64(i)*binWidth
		upper := minVal + float64(i+1)*binWidth

		// Для последнего бина включаем верхнюю границу
		if i == numBins-1 {
			upper += epsilon
		}

		bins[i] = HistogramBin{
			LowerBound: lower,
			UpperBound: upper,
			Count:      0,
		}
	}

	// Распределяем данные по бинам
	for _, value := range filteredData {
		// Определяем индекс бина
		binIndex := int(math.Floor((value - minVal) / binWidth))

		// Обработка граничного случая
		if binIndex >= numBins {
			binIndex = numBins - 1
		}
		if binIndex < 0 {
			binIndex = 0
		}

		bins[binIndex].Count++
	}

	// Вычисляем относительные частоты
	total := float64(len(filteredData))
	for i := range bins {
		bins[i].Frequency = float64(bins[i].Count) / total
	}

	return HistogramResult{
		Min:        minVal,
		Max:        maxVal,
		BinWidth:   binWidth,
		TotalCount: len(filteredData),
		Bins:       bins,
		DataName:   name,
	}
}
