// Package main
package main

import "github.com/squadracorsepolito/jsondbc/cmd"

func main() {
	cmd.Execute()
	/*can, err := cangoru.NewCANFromDBC("examples/simple.dbc")
	if err != nil {
		panic(err)
	}

	if err := can.ToDBC("res_simple.dbc"); err != nil {
		panic(err)
	}

	can, err = cangoru.NewCANFromDBC("examples/multiplexed_signal.dbc")
	if err != nil {
		panic(err)
	}

	if err := can.ToDBC("res_multiplexed_signal.dbc"); err != nil {
		panic(err)
	}*/
}
