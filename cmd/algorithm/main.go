package main

import (
	"classification-project/internal/interface/reader"
	"classification-project/internal/models"
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
	var N [models.Total]*models.Table
	for i := range models.TotalCv {
		N[i] = reader.ReadTableOrPanic(fmt.Sprintf("%s.txt", models.ClassificationName[i]))
	}
	N[models.Volume] = reader.ReadTableOrPanic("Vol.txt")
	N[models.Beta] = reader.ReadTableOrPanic("beta.txt")
	params.N = N

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

	rows, cols := N[models.Urban].Rows, N[models.Urban].Columns
	r := mat.NewDense(rows, cols, nil)

	for i := range rows {
		for j := range cols {
			tmpR := N[models.Volume].Get(i, j) / N[models.Beta].Get(i, j)
			r.Set(i, j, (N[models.Dust].Get(i, j)*res.Cv[models.Dust]+
				N[models.Urban].Get(i, j)*res.Cv[models.Urban]+
				N[models.Smoke].Get(i, j)*res.Cv[models.Smoke]-
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
	flag.Parse()
}
