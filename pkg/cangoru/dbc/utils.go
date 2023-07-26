package dbc

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"unicode"
)

const eof = rune(0)

func isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isNumber(ch rune) bool {
	return unicode.IsDigit(ch)
}

func isHexNumber(ch rune) bool {
	return isNumber(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isAlphaNumeric(ch rune) bool {
	return isLetter(ch) || isNumber(ch) || ch == '_' || ch == '-'
}

func isEOF(ch rune) bool {
	return ch == eof
}

func isSyntaxKeyword(r rune) bool {
	_, ok := syntaxKeywords[r]
	return ok
}

func parseUint(val string) (uint32, error) {
	res, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(res), nil
}

func parseHexInt(val string) (int, error) {
	if !strings.HasPrefix(val, "0x") && !strings.HasPrefix(val, "0X") {
		return 0, errors.New("invalid hex number")
	}
	res, err := strconv.ParseUint(val[2:], 16, 32)
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func parseInt(val string) (int, error) {
	res, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Print(err)
		return 0, err
	}
	return int(res), nil
}

func parseDouble(val string) (float64, error) {
	return strconv.ParseFloat(val, 64)
}
