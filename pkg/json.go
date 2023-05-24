package pkg

import (
	"encoding/json"
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

func (r *JsonReader) Read(file *os.File) (*CanModel, error) {
	jsonFile, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	canModel := &CanModel{}
	if err := json.Unmarshal(jsonFile, canModel); err != nil {
		return nil, err
	}

	return canModel, nil
}
