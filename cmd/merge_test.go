package cmd

import (
	"encoding/csv"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	file1, err := os.CreateTemp("", "file1.csv")
	require.NoError(t, err)
	defer os.Remove(file1.Name())

	w1 := csv.NewWriter(file1)
	require.NoError(t, w1.WriteAll([][]string{
		{"pr_number", "title"},
		{"1", "title1"},
		{"2", "title2"},
	}))
	w1.Flush()

	file2, err := os.CreateTemp("", "file2.csv")
	require.NoError(t, err)
	defer os.Remove(file2.Name())

	w2 := csv.NewWriter(file2)
	require.NoError(t, w2.WriteAll([][]string{
		{"pr_number", "author"},
		{"1", "author1"},
		{"3", "author3"},
	}))
	w2.Flush()

	outFile, err := os.CreateTemp("", "out.csv")
	require.NoError(t, err)
	defer os.Remove(outFile.Name())

	err = runMerge(file1.Name(), file2.Name(), outFile.Name())
	require.NoError(t, err)

	f, err := os.Open(outFile.Name())
	require.NoError(t, err)
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	require.NoError(t, err)

	expected := [][]string{
		{"pr_number", "title", "author"},
		{"1", "title1", "author1"},
	}
	assert.Equal(t, expected, records)
}
