package internal

import (
	"encoding/csv"
	"os"
)

// CSVWriter provides a generic CSV writing functionality.
type CSVWriter struct {
	writer *csv.Writer
	file   *os.File
}

// NewCSVWriter creates a new CSVWriter with the given filename and headers.
func NewCSVWriter(filename string, headers []string) (*CSVWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	return &CSVWriter{
		writer: writer,
		file:   file,
	}, nil
}

// Write writes a single record to the CSV file.
func (w *CSVWriter) Write(record []string) error {
	return w.writer.Write(record)
}

// Close flushes any buffered data to the underlying file and closes the file.
func (w *CSVWriter) Close() {
	w.writer.Flush()
	w.file.Close()
}
