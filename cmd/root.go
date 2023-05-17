// Package cmd contains all the commands of the application
package cmd

import (
	"os"

	"github.com/FerroO2000/canconv/cmd/convert"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "canconv",
	Short: "A tool to convert CAN models definded in JSON",
	Long:  ``,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(convert.ConvertCmd)
}
