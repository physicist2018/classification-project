package main

import (
	"classification-project/internal/interface/reader"
	"classification-project/internal/models"
	"classification-project/pkg/math/statistics"
	"classification-project/pkg/solver"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"gonum.org/v1/gonum/mat"
)

func main() {

	params := models.InputParameters{}
	ParseFlags(&params)

	rand.NewSource(time.Now().UnixNano())
	DustData := reader.ReadTableOrPanic("d.txt")
	UrbanData := reader.ReadTableOrPanic("u.txt")
	SmokeData := reader.ReadTableOrPanic("s.txt")
	VolumeData := reader.ReadTableOrPanic("Vol.txt")
	BetaData := reader.ReadTableOrPanic("beta.txt")

	rows, cols := DustData.Rows, DustData.Columns
	matDust := mat.NewDense(rows, cols, DustData.Data)
	matUrban := mat.NewDense(rows, cols, UrbanData.Data)
	matSmoke := mat.NewDense(rows, cols, SmokeData.Data)
	matVolume := mat.NewDense(rows, cols, VolumeData.Data)
	matBeta := mat.NewDense(rows, cols, BetaData.Data)

	// Поиск максимальной области с минимальной корреляцией
	minSize := params.MinSize
	r1, c1, r2, c2, _, err := statistics.FindMaxAreaMinCorrelation(matDust, matUrban, minSize)
	fmt.Println(r1, r2, c1, c2)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}
	matDust = matDust.Slice(r1, r2+1, c1, c2+1).(*mat.Dense)
	matUrban = matUrban.Slice(r1, r2+1, c1, c2+1).(*mat.Dense)
	matSmoke = matSmoke.Slice(r1, r2+1, c1, c2+1).(*mat.Dense)
	matVolume = matVolume.Slice(r1, r2+1, c1, c2+1).(*mat.Dense)
	matBeta = matBeta.Slice(r1, r2+1, c1, c2+1).(*mat.Dense)

	DustData = models.NewTable(r2-r1+1, c2-c1+1, nil, DustData.ColumnLabels[c1:c2+1], DustData.RowLabels[r1:r2+1])
	UrbanData = models.NewTable(r2-r1+1, c2-c1+1, nil, UrbanData.ColumnLabels[c1:c2+1], UrbanData.RowLabels[r1:r2+1])
	SmokeData = models.NewTable(r2-r1+1, c2-c1+1, nil, SmokeData.ColumnLabels[c1:c2+1], SmokeData.RowLabels[r1:r2+1])
	VolumeData = models.NewTable(r2-r1+1, c2-c1+1, nil, VolumeData.ColumnLabels[c1:c2+1], VolumeData.RowLabels[r1:r2+1])
	BetaData = models.NewTable(r2-r1+1, c2-c1+1, nil, BetaData.ColumnLabels[c1:c2+1], BetaData.RowLabels[r1:r2+1])

	rows, cols = matDust.Dims()
	for i := range rows {
		for j := range cols {
			DustData.SetWithLabels(i, j, matDust.At(i, j), DustData.RowLabels[i], DustData.ColumnLabels[j])
			UrbanData.SetWithLabels(i, j, matUrban.At(i, j), UrbanData.RowLabels[i], UrbanData.ColumnLabels[j])
			SmokeData.SetWithLabels(i, j, matSmoke.At(i, j), SmokeData.RowLabels[i], SmokeData.ColumnLabels[j])
			VolumeData.SetWithLabels(i, j, matVolume.At(i, j), VolumeData.RowLabels[i], VolumeData.ColumnLabels[j])
			BetaData.SetWithLabels(i, j, matBeta.At(i, j), BetaData.RowLabels[i], BetaData.ColumnLabels[j])
		}
	}

	params.N[models.Dust] = DustData
	params.N[models.Urban] = UrbanData
	params.N[models.Smoke] = SmokeData
	params.N[models.Volume] = VolumeData
	params.N[models.Beta] = BetaData

	loglevel := slog.LevelInfo
	if params.Debug {
		loglevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: loglevel,
	}))

	cls := solver.NewSolver(logger)
	res, err := cls.Solve(params)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Cv: %.3e\n", res.Cv)
	fmt.Printf("Discrepancy: %.2e\n", res.Discrepancy)

	rows, cols = params.N[models.Urban].Rows, params.N[models.Urban].Columns
	r := mat.NewDense(rows, cols, nil)

	for i := range rows {
		for j := range cols {
			tmpR := params.N[models.Volume].Get(i, j) / params.N[models.Beta].Get(i, j)
			r.Set(i, j, (params.N[models.Dust].Get(i, j)*res.Cv[models.Dust]+
				params.N[models.Urban].Get(i, j)*res.Cv[models.Urban]+
				params.N[models.Smoke].Get(i, j)*res.Cv[models.Smoke]-
				tmpR)/tmpR)

		}
	}

	fmt.Printf("Relative Discrepancy Matrix:\n")
	for i := range rows {
		for j := range cols {
			fmt.Printf("%+.2e  ", r.At(i, j))
		}
		fmt.Println()
	}
}

func ParseFlags(params *models.InputParameters) {
	flag.IntVar(&params.NPoints, "npoints", 4, "Число точек для матрицы")
	flag.IntVar(&params.NIters, "niters", 400, "Число повторений Монте-Карло")
	flag.IntVar(&params.NumPointsToAvg, "navg", 10, "Количество решений для усреднения")
	flag.Float64Var(&params.Lambda, "lambda", 0.01, "Параметр регуляризации")
	flag.BoolVar(&params.Debug, "debug", false, "Флаг отладки")
	flag.IntVar(&params.MinSize, "min-size", 5, "Минимальный размер области")
	flag.Parse()
}
