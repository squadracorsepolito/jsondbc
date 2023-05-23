// Package main
package main

import (
	"log"
	"os"

	"github.com/FerroO2000/canconv/pkg"
)

func main() {
	f, err := os.Open("./examples/simple.dbc")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	g := &pkg.DBCGenerator{}

	log.Print(*g.Read(f))

	//cmd.Execute()
}
