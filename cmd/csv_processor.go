package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/danielgazineu/commit-history/internal"
)

// CSVProcessor defines the interface for processing a single CSV record.
type CSVProcessor interface {
	ProcessRecord(record []string, geminiService internal.GeminiService) ([]string, error)
	GetOutputHeaders() []string
	ShouldSkip(record []string) bool
}

// processCSV provides a generic function for reading, processing, and writing CSV files.
func processCSV(processor CSVProcessor, geminiService internal.GeminiService, inputPath, outputPath string) error {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	reader := csv.NewReader(inputFile)

	// Read header
	_, err = reader.Read() // Discard the input header, as the processor will provide the output header
	if err != nil {
		return fmt.Errorf("failed to read header from input file: %w", err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write output headers
	if err := writer.Write(processor.GetOutputHeaders()); err != nil {
		return fmt.Errorf("failed to write output headers: %w", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read record from input file: %w", err)
		}

		if processor.ShouldSkip(record) {
			log.Printf("Skipping already processed record: %v", record)
			if err := writer.Write(record); err != nil { // Write original record if skipped
				log.Printf("Error writing skipped record to CSV: %v", err)
			}
			continue
		}

		processedRecord, err := processor.ProcessRecord(record, geminiService)
		if err != nil {
			log.Printf("Error processing record: %v. Writing original record with error.", err)
			// Decide how to handle errors: write original record, or a modified one with error status
			// For now, let's write the original record and log the error.
			if err := writer.Write(record); err != nil {
				log.Printf("Error writing original record after processing error to CSV: %v", err)
			}
			continue
		}

		if err := writer.Write(processedRecord); err != nil {
			return fmt.Errorf("failed to write processed record to output file: %w", err)
		}
	}
	return nil
}
