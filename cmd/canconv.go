package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/FerroO2000/canconv"
)

func main() {
	jsonFile, err := os.Open("model.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteFile, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	m := &canconv.Model{}
	if err := json.Unmarshal(byteFile, m); err != nil {
		panic(err)
	}

	f, err := os.Create("pippo.dbc")
	if err != nil {
		panic(err)
	}

	dbcGen := canconv.NewDBCGenerator()
	dbcGen.Generate(m, f)

	defer f.Close()
}
