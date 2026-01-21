package models

type MatrixData struct {
	Rows    int       `json:"rows"`
	Columns int       `json:"columns"`
	Data    []float64 `json:"data"`
}

type Table struct {
	MatrixData
	ColumnLabels []string `json:"column_labels"`
	RowLabels    []string `json:"row_labels"`
}

func NewMatrix(rows, columns int, data []float64) *MatrixData {
	if rows <= 0 || columns <= 0 {
		panic("invalid dimensions")
	}

	if data == nil {
		return &MatrixData{
			Rows:    rows,
			Columns: columns,
			Data:    make([]float64, rows*columns),
		}
	}
	if len(data) != rows*columns {
		panic("invalid data size")
	}

	return &MatrixData{
		Rows:    rows,
		Columns: columns,
		Data:    data,
	}
}

func NewTable(rows, columns int, data []float64, columnLabels, rowLabels []string) *Table {
	if len(columnLabels) != columns || len(rowLabels) != rows {
		panic("invalid labels size")
	}
	matrix := NewMatrix(rows, columns, data)
	return &Table{
		MatrixData:   *matrix,
		ColumnLabels: columnLabels,
		RowLabels:    rowLabels,
	}
}

func (m *MatrixData) Get(row, col int) float64 {
	if row < 0 || row >= m.Rows || col < 0 || col >= m.Columns {
		panic("index out of bounds")
	}
	return m.Data[row*m.Columns+col]
}

func (m *MatrixData) Set(row, col int, val float64) {
	if row < 0 || row >= m.Rows || col < 0 || col >= m.Columns {
		panic("index out of bounds")
	}
	m.Data[row*m.Columns+col] = val
}

func (m *Table) SetWithLabels(row, col int, val float64, rowlabel string, collabel string) {
	if row < 0 || row >= m.Rows || col < 0 || col >= m.Columns {
		panic("index out of bounds")
	}
	m.Data[row*m.Columns+col] = val
	m.RowLabels[row] = rowlabel
	m.ColumnLabels[col] = collabel
}
