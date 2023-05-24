// Package main
package main

import (
	"encoding/json"
	"os"

	"github.com/FerroO2000/canconv/pkg"
)

func main() {
	f, err := os.Open("./data/Model3CAN.dbc")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	g := &pkg.DBCReader{}

	m, _ := g.Read(f)

	// for _, msg := range m.Messages {
	// 	log.Print(msg)
	// 	for _, sig := range msg.Signals {
	// 		log.Print(sig)
	// 	}
	// }

	jsonFile, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("./data/model3.json", jsonFile, 0644); err != nil {
		panic(err)
	}

	//cmd.Execute()
}
