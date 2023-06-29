package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type JsonWriter struct{}

func NewJsonWriter() *JsonWriter {
	return &JsonWriter{}
}

func (w *JsonWriter) Write(file *os.File, canModel *CanModel) error {
	jsonFile, err := json.MarshalIndent(canModel, "", "\t")
	if err != nil {
		return err
	}

	_, err = file.Write(jsonFile)
	return err
}

type JsonReader struct{}

func NewJsonReader() *JsonReader {
	return &JsonReader{}
}

func (r *JsonReader) getLineErr(input []byte, offset int, jsonErr error) error {
	lf := rune(0x0A)

	if offset > len(input) || offset < 0 {
		return fmt.Errorf("couldn't find offset %d within the input", offset)
	}

	line := 1
	for i, b := range input {
		if b == byte(lf) {
			line++
		}
		if i == offset {
			break
		}
	}

	return fmt.Errorf("line %d: %v", line, jsonErr)
}

func (r *JsonReader) Read(file *os.File) (*CanModel, error) {
	jsonFile, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	canModel := &CanModel{}
	err = json.Unmarshal(jsonFile, canModel)
	if err != nil {
		switch jsonErr := err.(type) {
		case *json.UnmarshalTypeError:
			return nil, r.getLineErr(jsonFile, int(jsonErr.Offset), jsonErr)
		case *json.SyntaxError:
			return nil, r.getLineErr(jsonFile, int(jsonErr.Offset), jsonErr)
		}

		return nil, err
	}

	return canModel, nil
}
