package solver

import (
	"classification-project/internal/models"
	"fmt"
	"log/slog"
	"math"

	"gonum.org/v1/gonum/mat"
)

type CholProblem struct {
	logger *slog.Logger
	A      *mat.Dense
	b      *mat.VecDense
	labmda float64
}

func NewCholProblem(matrix *mat.Dense, vector *mat.VecDense, lambda float64, logger *slog.Logger) *CholProblem {
	A, b := BalanceProblem(matrix, vector)
	return &CholProblem{
		A:      A,
		b:      b,
		labmda: lambda,
		logger: logger,
	}
}

func (p *CholProblem) Solve(xinit []float64) (models.OutputSolution, error) {

	p.logger.Debug("", slog.Float64("Condition number", mat.Cond(p.A, 2)))
	//fmt.Printf("Condition number: %f\n", mat.Cond(p.A, 2))

	x, err := solveRegularizedLS(p.A, p.b, p.labmda)

	if err != nil {
		return models.OutputSolution{}, err
	}

	result := x.RawVector().Data

	p.logger.Debug("Solution", slog.Any("x", result),
		slog.Any("A", p.A),
		slog.Any("b", p.b),
		slog.Any("x", x))

	// fmt.Println("Матрица A:")
	// fmt.Printf(" %+.3f\n", mat.Formatted(p.A, mat.Prefix(" ")))

	// fmt.Println("Вектор b:")
	// fmt.Printf(" %+.3f\n", mat.Formatted(p.b, mat.Prefix(" ")))

	// fmt.Println("Вектор x:")
	// fmt.Printf(" %+.3f\n", mat.Formatted(x, mat.Prefix(" ")))
	// Remove negative values from the solution vector
	// for i := range result {
	// 	if result[i] < 0 {
	// 		result[i] = 0
	// 	}
	// }

	F := mat.NewDense(p.A.RawMatrix().Rows, 1, nil)
	F.Mul(p.A, x)

	norm := 0.0
	residual := mat.NewDense(p.A.RawMatrix().Rows, 1, nil)
	residual.Sub(p.b, F)

	for i := range 3 {
		norm += math.Pow(residual.At(i, 0)/p.b.AtVec(i), 2)
	}

	return models.OutputSolution{
		Cv:          result,
		Discrepancy: math.Sqrt(norm),
	}, nil
}

// BalanceProblem balances the problem by scaling rows to have unit L2 norm.
// It returns a new matrix B and a new vector b
func BalanceProblem(A *mat.Dense, b *mat.VecDense) (*mat.Dense, *mat.VecDense) {
	m, n := A.Dims()
	B := mat.NewDense(m, n, nil)
	v := mat.NewVecDense(m, nil)

	for i := range m {
		// Получаем строку как срез
		row := mat.Row(nil, i, A)
		// Вычисляем L2-норму строки
		norm := mat.Norm(mat.NewVecDense(n, row), 2)

		if norm > 1e-12 { // Защита от нулевых строк
			scale := 1.0 / norm
			for j := range n {
				B.Set(i, j, A.At(i, j)*scale)
			}
			v.SetVec(i, b.AtVec(i)*scale)
		} else {
			// Если строка нулевая — копируем как есть
			for j := range n {
				B.Set(i, j, A.At(i, j))
			}
			v.SetVec(i, b.AtVec(i))
		}
	}
	return B, v
}

// Решает переопределённую СЛАУ с регуляризацией Тихонова
func solveRegularizedLS(A *mat.Dense, b *mat.VecDense, lambda float64) (*mat.VecDense, error) {
	m, n := A.Dims()
	if m <= n {
		return nil, fmt.Errorf("ожидается переопределённая система (m > n), но m=%d, n=%d", m, n)
	}

	// 1. Создаём симметричную матрицу сразу
	symATA := mat.NewSymDense(n, nil)

	// 2. Вычисляем A^T * A + λI напрямую в symATA
	// Вычисляем произведение A^T * A
	temp := mat.NewDense(n, n, nil)
	temp.Mul(A.T(), A)

	// Копируем в симметричную матрицу (только верхний треугольник)
	for i := 0; i < n; i++ {
		for j := i; j < n; j++ {
			val := temp.At(i, j)
			if i == j {
				val += lambda // добавляем регуляризацию на диагонали
			}
			symATA.SetSym(i, j, val)
		}
	}

	// 3. Вычисляем A^T * b
	ATb := mat.NewVecDense(n, nil)
	ATb.MulVec(A.T(), b)

	// 4. Решаем через Холецкого
	var chol mat.Cholesky
	ok := chol.Factorize(symATA)
	if !ok {
		return nil, fmt.Errorf("ATA + λI не положительно определена (λ=%g)", lambda)
	}

	x := mat.NewVecDense(n, nil)
	if err := chol.SolveVecTo(x, ATb); err != nil {
		return nil, fmt.Errorf("ошибка при решении: %v", err)
	}

	return x, nil
}

func solveRegularizedLS1(A *mat.Dense, b *mat.VecDense, lambda float64) (*mat.VecDense, error) {
	m, n := A.Dims()
	if m <= n {
		return nil, fmt.Errorf("ожидается переопределённая система (m > n), но m=%d, n=%d", m, n)
	}

	// Вычисляем A^T * A + λI
	ATA := mat.NewDense(n, n, nil)
	ATA.Mul(A.T(), A)
	for i := 0; i < n; i++ {
		ATA.Set(i, i, ATA.At(i, i)+lambda)
	}

	// Вычисляем A^T * b
	ATb := mat.NewVecDense(n, nil)
	ATb.MulVec(A.T(), b)

	// LU разложение
	var lu mat.LU
	lu.Factorize(ATA)

	// Решаем систему: ATA * x = ATb
	// trans = false, так как решаем ATA * x = ATb, а не ATA^T * x = ATb
	x := mat.NewVecDense(n, nil)
	if err := lu.SolveVecTo(x, false, ATb); err != nil {
		return nil, fmt.Errorf("ошибка при решении: %v (λ=%g)", err, lambda)
	}

	return x, nil
}
