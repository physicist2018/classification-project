package models

const (
	Dust int = iota
	Smoke
	Urban
	Beta
	Volume
	Total
)

const (
	TotalCv = 3
)

var ClassificationName = [...]string{"d", "s", "u"}

type InputParameters struct {
	N              [Total]*Table // Доли вкладов
	NPoints        int           // Число точек для составления системы уравнений
	NIters         int           // Число итераций Монте-Карло
	NWorkers       int           // Число потоков для параллельной обработки
	NumPointsToAvg int           // количество решений для усреднения
	Lambda         float64       // Параметр регуляризации
	Debug          bool          // Флаг отладки
	MinSize        int           // Минимальный размер области
}

type DataPacket struct {
	Row  int
	Col  int
	Data []float64
}

type ProcessResult struct {
	Cv [TotalCv]float64
}

type Index struct {
	Row int
	Col int
}

type OutputSolution struct {
	Cv          []float64
	Discrepancy float64
}
