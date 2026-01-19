package reader

import (
	"classification-project/internal/models"
	"fmt"
	"io"
)

type TXTMatrixReader struct {
	reader io.Reader
}

func NewTXTMatrixReader(reader io.Reader) *TXTMatrixReader {
	return &TXTMatrixReader{reader: reader}
}

func (r *TXTMatrixReader) ReadMatrix() (*models.MatrixData, error) {
	var rows, columns int
	var data []float64
	var err error

	if _, err = fmt.Fscanf(r.reader, "%d %d\n", &rows, &columns); err != nil {
		return nil, err
	}

	data = make([]float64, rows*columns)
	for i := range rows {
		for j := range columns {
			if _, err = fmt.Fscanf(r.reader, "%f ", &data[i*columns+j]); err != nil {
				return nil, err
			}
		}
	}

	return models.NewMatrix(rows, columns, data), nil
}
