package cmd

import (
	"encoding/csv"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	file1 string
	file2 string
	out   string
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge two CSV files.",
	Long:  `Merge two CSV files based on a common column.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runMerge(file1, file2, out); err != nil {
			log.Fatal(err)
		}
	},
}

func runMerge(file1, file2, out string) error {
	f1, err := os.Open(file1)
	if err != nil {
		return err
	}
	defer f1.Close()

	r1 := csv.NewReader(f1)
	records1, err := r1.ReadAll()
	if err != nil {
		return err
	}

	f2, err := os.Open(file2)
	if err != nil {
		return err
	}
	defer f2.Close()

	r2 := csv.NewReader(f2)
	records2, err := r2.ReadAll()
	if err != nil {
		return err
	}

	header1 := records1[0]
	header2 := records2[0]
	mergedHeader := append(header1, header2[1:]...)

	recordsMap := make(map[string][]string)
	for _, record := range records1[1:] {
		recordsMap[record[0]] = record
	}

	var mergedRecords [][]string
	mergedRecords = append(mergedRecords, mergedHeader)

	for _, record2 := range records2[1:] {
		prNumber := record2[0]
		if record1, ok := recordsMap[prNumber]; ok {
			mergedRecord := append(record1, record2[1:]...)
			mergedRecords = append(mergedRecords, mergedRecord)
		}
	}

	outFile, err := os.Create(out)
	if err != nil {
		return err
	}
	defer outFile.Close()

	w := csv.NewWriter(outFile)
	return w.WriteAll(mergedRecords)
}

func init() {
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().StringVar(&file1, "file1", "", "First CSV file")
	mergeCmd.Flags().StringVar(&file2, "file2", "", "Second CSV file")
	mergeCmd.Flags().StringVar(&out, "out", "merged.csv", "Output CSV file")
	mergeCmd.MarkFlagRequired("file1")
	mergeCmd.MarkFlagRequired("file2")
}
