// Package main
package main

import (
	"github.com/squadracorsepolito/jsondbc/pkg/cangoru"
)

func main() {
	//cmd.Execute()
	if err := cangoru.ReadFromDBC("examples/multiplexed_signal.dbc"); err != nil {
		panic(err)
	}
}
