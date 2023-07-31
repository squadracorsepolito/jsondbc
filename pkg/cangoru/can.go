package cangoru

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/squadracorsepolito/jsondbc/pkg/cangoru/dbc"
)

type CAN struct {
	tmp *dbc.DBC
}

func NewCANFromDBC(dbcFilename string) (*CAN, error) {
	ext := path.Ext(dbcFilename)
	if ext != ".dbc" {
		return nil, fmt.Errorf("file %s: extension must be .dbc; got %s", dbcFilename, ext)
	}

	file, err := ioutil.ReadFile(dbcFilename)
	if err != nil {
		return nil, err
	}

	fileMime := http.DetectContentType(file)
	if fileMime != "text/plain; charset=utf-8" {
		return nil, fmt.Errorf("file %s: content type must be text/plain; got %s", dbcFilename, fileMime)
	}

	parser := dbc.NewParser(file)
	dbcAST, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	return &CAN{tmp: dbcAST}, nil
}

func (c *CAN) ToDBC(dbcFilename string) error {
	writer := dbc.NewWriter()

	wFile, err := os.Create(dbcFilename)
	if err != nil {
		return err
	}
	defer wFile.Close()

	_, err = wFile.WriteString(writer.Write(c.tmp))
	return err
}
