package solver

import (
	"classification-project/internal/models"
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/optimize"
)

type Problem struct {
	A *mat.Dense
	b *mat.VecDense
}

type Problemer interface {
	Func(x []float64) float64
}

func NewProblem(matrix *mat.Dense, vector *mat.VecDense) *Problem {
	A, b := BalanceProblem(matrix, vector)
	return &Problem{
		A: A,
		b: b,
	}
}

// Для вектора решений вычисляем невязку по решаемой задаче
func (p *Problem) Func(x []float64) float64 {
	rows, cols := p.A.Dims()
	tmpb := mat.NewVecDense(rows, nil)
	tmpb.MulVec(p.A, mat.NewVecDense(cols, x))
	norm := 0.0
	for i := range tmpb.Len() {
		norm += math.Pow(math.Abs(tmpb.AtVec(i)-p.b.AtVec(i))/p.b.AtVec(i), 2)
	}
	penalty := 0.0
	for i := range len(x) {
		if x[i] < 0 {
			penalty += 1000 * math.Pow(x[i], 2)
		}
	}
	return norm
}

func (p *Problem) Solve(xinit []float64) (models.OutputSolution, error) {
	pp := optimize.Problem{
		Func: p.Func,
	}
	if xinit == nil {
		xinit = make([]float64, p.A.RawMatrix().Cols)
		for i := range xinit {
			xinit[i] = 1.0 / float64(p.A.RawMatrix().Cols)
		}
	}

	result, err := optimize.Minimize(pp, xinit, &optimize.Settings{
		MajorIterations: 1000}, &optimize.NelderMead{})
	if err != nil {
		return models.OutputSolution{}, err
	}

	fmt.Println("Матрица A:")
	fmt.Printf("%v\n", mat.Formatted(p.A, mat.Prefix(" ")))

	fmt.Println("Вектор b:")
	fmt.Printf("%v\n", mat.Formatted(p.b, mat.Prefix(" ")))

	fmt.Println("Вектор x:")
	fmt.Printf("%v\n", mat.Formatted(mat.NewVecDense(len(result.X), result.X), mat.Prefix(" ")))

	// Remove negative values from the solution vector
	// for i := range result.X {
	// 	if result.X[i] < 0 {
	// 		result.X[i] = 0
	// 	}
	// }

	return models.OutputSolution{
		Cv:          result.X,
		Discrepancy: math.Sqrt(result.F),
	}, nil
}
