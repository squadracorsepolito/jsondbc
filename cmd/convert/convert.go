// Package convert contains the convert command
package convert

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/FerroO2000/canconv/pkg"
	"github.com/spf13/cobra"
)

var (
	extension   string
	inFileName  string
	outFileName string
)

const (
	dbcExt  = ".dbc"
	jsonExt = ".json"
)

var validInExt = []string{jsonExt, dbcExt}
var validOutExt = []string{jsonExt, dbcExt}

// convert is the handler for the convert command.
// It opens the input file, reads, converts and write into the output file.
func convert() error {
	var reader pkg.Reader
	var writer pkg.Writer

	inExt := filepath.Ext(inFileName)
	switch inExt {
	case jsonExt:
		reader = pkg.NewJsonReader()
	case dbcExt:
		reader = pkg.NewDBCReader()

	default:
		return fmt.Errorf("%s extension is not supported as input file", inExt)
	}

	if outFileName == "" {
		outFileName = inFileName[:len(inFileName)-len(inExt)] + extension
	}
	outExt := filepath.Ext(outFileName)
	switch outExt {
	case jsonExt:
		writer = pkg.NewJsonWriter()
	case dbcExt:
		writer = pkg.NewDBCWriter()

	default:
		return fmt.Errorf("%s extension is not supported as output file", outExt)
	}

	inFile, err := os.Open(inFileName)
	if err != nil {
		return err
	}
	defer inFile.Close()

	canModel, err := reader.Read(inFile)
	if err != nil {
		return err
	}

	if err := canModel.Validate(); err != nil {
		return err
	}

	outFile, err := os.Create(outFileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err := writer.Write(outFile, canModel); err != nil {
		return err
	}

	return nil
}

// ConvertCmd represents the convert command
var ConvertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Converts the CAM model defined in the input file to the specified extension",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return convert()
	},
}

// init initializes the flags for the convert command.
func init() {
	ConvertCmd.Flags().StringVar(&inFileName, "in", "", "Sets the input file")
	if err := ConvertCmd.MarkFlagFilename("in", validInExt...); err != nil {
		log.Fatal(err)
	}
	if err := ConvertCmd.MarkFlagRequired("in"); err != nil {
		log.Fatal(err)
	}

	ConvertCmd.Flags().StringVarP(&extension, "ext", "e", dbcExt, "Sets the output file extension")

	ConvertCmd.Flags().StringVar(&outFileName, "out", "", "Sets the output file")
	if err := ConvertCmd.MarkFlagFilename("out", validOutExt...); err != nil {
		log.Fatal(err)
	}
}
