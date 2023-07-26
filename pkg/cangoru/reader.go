package cangoru

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/alecthomas/repr"
	"github.com/squadracorsepolito/jsondbc/pkg/cangoru/dbc"
)

func ReadFromDBC(dbcFilename string) error {
	ext := path.Ext(dbcFilename)
	if ext != ".dbc" {
		return fmt.Errorf("file %s: extension must be .dbc; got %s", dbcFilename, ext)
	}

	file, err := ioutil.ReadFile(dbcFilename)
	if err != nil {
		return err
	}

	fileMime := http.DetectContentType(file)
	if fileMime != "text/plain; charset=utf-8" {
		return fmt.Errorf("file %s: content type must be text/plain; got %s", dbcFilename, fileMime)
	}

	parser := dbc.NewParser(file)
	dbcAST, err := parser.Parse()
	if err != nil {
		return err
	}

	repr.Print(dbcAST)

	return nil
}
