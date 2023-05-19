// Package convert contains the convert command
package convert

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/FerroO2000/canconv/pkg"
	"github.com/spf13/cobra"
)

var (
	extension  string
	inputFile  string
	outputFile string
)

const (
	fileDBC = "dbc"
)

// convert is the handler for the convert command.
// It opens the input file, reads it and converts it to the specified extension.
func convert() error {
	jsonFile, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteFile, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	model := &pkg.CanModel{}
	if err := json.Unmarshal(byteFile, model); err != nil {
		return err
	}

	if outputFile == "" {
		outputFile = inputFile[:len(inputFile)-len("json")-1] + "." + extension
	}
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	switch extension {
	case fileDBC:
		generator := pkg.NewDBCGenerator()
		generator.Generate(model, outFile)

	default:
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
	ConvertCmd.Flags().StringVar(&inputFile, "in", "", "Sets the input file")
	if err := ConvertCmd.MarkFlagFilename("in", "json"); err != nil {
		log.Fatal(err)
	}
	if err := ConvertCmd.MarkFlagRequired("in"); err != nil {
		log.Fatal(err)
	}

	ConvertCmd.Flags().StringVar(&extension, "ext", "dbc", "Sets the output file extension")

	ConvertCmd.Flags().StringVar(&outputFile, "out", "", "Sets the output file")
	if err := ConvertCmd.MarkFlagFilename("out", extension); err != nil {
		log.Fatal(err)
	}
}
