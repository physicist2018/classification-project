package statistics

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"gonum.org/v1/gonum/mat"
)

func TestCorr2(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
		wantErr  bool
	}{
		{
			name:     "Perfect positive correlation",
			a:        []float64{1, 2, 3, 4},
			b:        []float64{1, 2, 3, 4},
			expected: 1.0,
			wantErr:  false,
		},
		{
			name:     "Perfect negative correlation",
			a:        []float64{1, 2, 3, 4},
			b:        []float64{5, 4, 3, 2},
			expected: -1.0,
			wantErr:  false,
		},
		{
			name:     "No correlation",
			a:        []float64{1, 2, 1, 2},
			b:        []float64{2, 1, 2, 1},
			expected: -1.0, // На самом деле -1 для этой последовательности
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := int(math.Sqrt(float64(len(tt.a))))
			A := mat.NewDense(n, n, tt.a)
			B := mat.NewDense(n, n, tt.b)

			result, err := Corr2(A, B)

			if tt.wantErr && err == nil {
				t.Errorf("ожидалась ошибка, но получено nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("неожиданная ошибка: %v", err)
			}

			if !tt.wantErr && math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("ожидалось %.10f, получено %.10f", tt.expected, result)
			}
		})
	}
}

func BenchmarkCorr2(b *testing.B) {
	size := 1000
	A := mat.NewDense(size, size, nil)
	B := mat.NewDense(size, size, nil)

	// Заполняем случайными данными
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size*size; i++ {
		A.Set(i/size, i%size, rand.Float64())
		B.Set(i/size, i%size, rand.Float64())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Corr2(A, B)
	}
}
