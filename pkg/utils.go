// Package pkg provides utilities for the cli
package pkg

import (
	"os"
	"strconv"
)

func formatFloat(val float64) string {
	return strconv.FormatFloat(val, 'f', -1, 64)
}

func formatString(val string) string {
	return "\"" + val + "\""
}

func formatUint(val uint32) string {
	return strconv.FormatUint(uint64(val), 10)
}

func parseUint(val string) (uint32, error) {
	res, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(res), nil
}

func parseFloat(val string) (float64, error) {
	return strconv.ParseFloat(val, 64)
}

type file struct {
	f *os.File
}

func newFile(f *os.File) *file {
	return &file{f: f}
}

func (f *file) print(str ...string) {
	tmp := ""
	for i, s := range str {
		if i == 0 {
			tmp += s
			continue
		}
		if len(s) == 0 {
			continue
		}
		if str[i-1] == "\t" {
			tmp += s
			continue
		}
		tmp += " " + s
	}
	_, err := f.f.WriteString(tmp + "\n")
	if err != nil {
		panic(err)
	}
}
