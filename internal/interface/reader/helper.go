package reader

import "classification-project/internal/models"

func ReadTableOrPanic(filename string) *models.Table {
	table, err := ReadTableFromFile(filename)
	if err != nil {
		panic(err)
	}
	return table
}
