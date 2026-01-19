package reader

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"

	"classification-project/internal/models"
)

// ReadTableFromFile читает таблицу из текстового файла и возвращает *models.Table
func ReadTableFromFile(filename string) (*models.Table, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string

	// Читаем все строки, игнорируя комментарии и пустые строки
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}

	if len(lines) < 2 {
		return nil, errors.New("недостаточно данных в файле: нужна хотя бы одна строка заголовков и одна строка данных")
	}

	// Парсим заголовки столбцов (первая строка)
	columnLabels, err := parseQuotedStrings(lines[0])
	if err != nil {
		return nil, err
	}
	cols := len(columnLabels)

	// Парсим строки данных
	var rowLabels []string
	var data []float64

	for i, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) != cols+1 {
			return nil, errors.New(
				"неверное количество полей в строке " + strconv.Itoa(i+2) +
					": ожидается " + strconv.Itoa(cols+1) + ", получено " + strconv.Itoa(len(fields)))
		}

		// Первая часть — метка строки (в кавычках)
		rowLabel, err := unquote(fields[0])
		if err != nil {
			return nil, err
		}
		rowLabels = append(rowLabels, rowLabel)

		// Остальные — числовые значения
		for j, field := range fields[1:] {
			val, err := strconv.ParseFloat(field, 64)
			if err != nil {
				return nil, errors.New(
					"ошибка парсинга числа в строке " + strconv.Itoa(i+2) +
						", столбец " + strconv.Itoa(j+2) + ": " + err.Error())
			}
			data = append(data, val)
		}
	}

	rows := len(rowLabels)

	// Создаём и возвращаем Table
	table := models.NewTable(rows, cols, data, columnLabels, rowLabels)
	return table, nil
}

// parseQuotedStrings разбирает строку с quoted-строками (в "кавычках")
func parseQuotedStrings(line string) ([]string, error) {
	var result []string
	fields := strings.Fields(line)
	for _, field := range fields {
		unquoted, err := unquote(field)
		if err != nil {
			return nil, err
		}
		result = append(result, unquoted)
	}
	return result, nil
}

// unquote убирает кавычки с начала и конца строки, если они есть
func unquote(s string) (string, error) {
	if len(s) < 2 {
		return "", errors.New("строка слишком короткая для кавычек: " + s)
	}
	if s[0] != '"' || s[len(s)-1] != '"' {
		return "", errors.New("строка не заключена в кавычки: " + s)
	}
	return s[1 : len(s)-1], nil
}
