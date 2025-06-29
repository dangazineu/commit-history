package internal

import (
	"encoding/csv"
	"os"
	"testing"

	"github.com/google/go-github/v62/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteRecords(t *testing.T) {
	file, err := os.CreateTemp("", "test.csv")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	writer, err := NewCSVWriter(file.Name())
	require.NoError(t, err)

	pr := &github.PullRequest{
		Number: github.Int(1),
		Title:  github.String("test title"),
		Body:   github.String("test body"),
		Head:   &github.PullRequestBranch{SHA: github.String("head_sha")},
	}
	err = writer.Write(pr, true, "merge title", "merge body", "source_link", "resolved_source_link", "unidiff")
	require.NoError(t, err)
	writer.Close()

	f, err := os.Open(file.Name())
	require.NoError(t, err)
	defer f.Close()

	r := csv.NewReader(f)
	// Read header
	_, err = r.Read()
	require.NoError(t, err)

	record, err := r.Read()
	require.NoError(t, err)

	assert.Equal(t, "1", record[0])
	assert.Equal(t, "head_sha", record[1])
	assert.Equal(t, "", record[2])
	assert.Equal(t, "test title", record[3])
	assert.Equal(t, "test body", record[4])
	assert.Equal(t, "true", record[5])
	assert.Equal(t, "merge title", record[6])
	assert.Equal(t, "merge body", record[7])
	assert.Equal(t, "source_link", record[8])
	assert.Equal(t, "resolved_source_link", record[9])
	assert.Equal(t, "unidiff", record[10])
}


