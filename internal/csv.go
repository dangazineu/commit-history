package internal

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/google/go-github/v62/github"
)

type CSVWriter struct {
	writer *csv.Writer
	file   *os.File
}

func NewCSVWriter(filename string) (*CSVWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)
	// Write header
	header := []string{
		"pr_number",
		"before_merge_commit_hash",
		"after_merge_commit_hash",
		"pr_title",
		"pr_body",
		"is_squash_merge",
		"merge_commit_title",
		"merge_commit_body",
		"source_link",
		"resolved_source_link",
		"source_link_unidiff",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	return &CSVWriter{
		writer: writer,
		file:   file,
	}, nil
}

func (w *CSVWriter) Write(pr *github.PullRequest, isSquash bool, title, body, sourceLink, resolvedSourceLink, unidiff string) error {
	record := []string{
		strconv.Itoa(*pr.Number),
		*pr.Head.SHA,
		*pr.MergeCommitSHA,
		*pr.Title,
		*pr.Body,
		strconv.FormatBool(isSquash),
		title,
		body,
		sourceLink,
		resolvedSourceLink,
		unidiff,
	}
	return w.writer.Write(record)
}

func (w *CSVWriter) Close() {
	w.writer.Flush()
	w.file.Close()
}

func NewAugmentedCSVWriter(filename string) (*CSVWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)
	// Write header
	header := []string{
		"pr_number",
		"before_merge_commit_hash",
		"after_merge_commit_hash",
		"pr_title",
		"pr_body",
		"is_squash_merge",
		"merge_commit_title",
		"merge_commit_body",
		"source_link",
		"resolved_source_link",
		"source_link_unidiff",
		"gemini_proposed_title",
		"gemini_proposed_body",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	return &CSVWriter{
		writer: writer,
		file:   file,
	}, nil
}

func (w *CSVWriter) WriteAugmented(record []string, geminiTitle, geminiBody string) error {
	newRecord := append(record, geminiTitle, geminiBody)
	return w.writer.Write(newRecord)
}

func NewAnalyzedCSVWriter(filename string) (*CSVWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)
	// Write header
	header := []string{
		"pr_number",
		"before_merge_commit_hash",
		"after_merge_commit_hash",
		"pr_title",
		"pr_body",
		"is_squash_merge",
		"merge_commit_title",
		"merge_commit_body",
		"source_link",
		"resolved_source_link",
		"source_link_unidiff",
		"gemini_proposed_title",
		"gemini_proposed_body",
		"gemini_analysis",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	return &CSVWriter{
		writer: writer,
		file:   file,
	}, nil
}

func (w *CSVWriter) WriteAnalyzed(record []string, analysis string) error {
	newRecord := append(record, analysis)
	return w.writer.Write(newRecord)
}
