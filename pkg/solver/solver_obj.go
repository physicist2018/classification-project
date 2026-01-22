package solver

import (
	"classification-project/internal/models"
	"classification-project/pkg/math/statistics"
	"fmt"
	"log/slog"
	"math/rand"
	"sort"

	"gonum.org/v1/gonum/mat"
)

type Solver struct {
	// Define fields here
	logger *slog.Logger
}

func NewSolver(logger *slog.Logger) *Solver {
	return &Solver{
		logger: logger,
	}
}

func (s *Solver) Solve(p models.InputParameters) (models.OutputSolution, error) {
	//mkm2cm3Tom3m3 := 1.0 //1e-12
	scaleFactor := 1.0e-6
	solutions := make([]models.OutputSolution, 0, p.NIters)
	for _ = range p.NIters {
		indices := s.generateIndices(p.N[0].Rows, p.N[0].Columns, p.NPoints)
		tmpA := mat.NewDense(p.NPoints, 3, nil)
		tmpb := mat.NewVecDense(p.NPoints, nil)

		for j := range p.NPoints {
			tmpA.Set(j, 0, p.N[models.Dust].Get(indices[j].Row, indices[j].Col))
			tmpA.Set(j, 1, p.N[models.Urban].Get(indices[j].Row, indices[j].Col))
			tmpA.Set(j, 2, p.N[models.Smoke].Get(indices[j].Row, indices[j].Col))
			tmpb.SetVec(j, p.N[models.Volume].Get(indices[j].Row, indices[j].Col)/p.N[models.Beta].Get(indices[j].Row, indices[j].Col)*scaleFactor)
		}

		//m := NewProblem(tmpA, tmpb)
		m := NewCholProblem(tmpA, tmpb, p.Lambda, s.logger)
		sol, err := m.Solve(nil)
		if err == nil {
			solutions = append(solutions, sol)
		}

	}
	nValid := len(solutions)
	fmt.Printf("Num Valid Solutions: %d\n", nValid)
	sort.Slice(solutions, func(i, j int) bool {
		return solutions[i].Discrepancy < solutions[j].Discrepancy
	})

	// Простой расчет гистограмм
	fmt.Println("=== Простой расчет гистограмм ===")
	fmt.Printf("=== Масштаб для Cv  x10¹² ===\n")
	simpleResults := statistics.CalculateHistograms(solutions, 10)
	statistics.PrintAllHistograms(simpleResults, 50)

	numPtsToAvg := min(nValid, p.NumPointsToAvg)
	scale := 1.0 / float64(numPtsToAvg)
	cfinal := make([]float64, models.TotalCv)
	Discr := 0.0
	for i := range numPtsToAvg {
		cfinal[0] += solutions[i].Cv[0] * scale / scaleFactor
		cfinal[1] += solutions[i].Cv[1] * scale / scaleFactor
		cfinal[2] += solutions[i].Cv[2] * scale / scaleFactor
		Discr += solutions[i].Discrepancy * scale
	}

	return models.OutputSolution{
		Cv:          cfinal,
		Discrepancy: Discr,
	}, nil
}

func (s *Solver) generateIndices(rows, cols, nPoints int) []models.Index {
	indices := make([]models.Index, nPoints)
	for i := range indices {
		indices[i] = models.Index{
			Row: rand.Intn(rows),
			Col: rand.Intn(cols),
		}
	}
	return indices
}
