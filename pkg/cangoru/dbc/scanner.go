package dbc

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type tokenKind uint

const (
	tokenError tokenKind = iota
	tokenEOF
	tokenSpace

	tokenIdent
	tokenNumber
	tokenNumberRange
	tokenString
	tokenKeyword
	tokenSyntax
)

var tokenNames = map[tokenKind]string{
	tokenError: "error",
	tokenEOF:   "eof",
	tokenSpace: "space",

	tokenIdent:       "ident",
	tokenNumber:      "number",
	tokenNumberRange: "number_range",
	tokenString:      "string",
	tokenKeyword:     "keyword",
	tokenSyntax:      "syntax",
}

const eof = rune(0)

type token struct {
	kind     tokenKind
	kindName string
	value    string
	start    int
	col      int
	line     int
}

func isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isNumber(ch rune) bool {
	return unicode.IsDigit(ch)
}

func isAlphaNumeric(ch rune) bool {
	return isLetter(ch) || isNumber(ch) || ch == '_' || ch == '-'
}

func isEOF(ch rune) bool {
	return ch == eof
}

type scanner struct {
	r     *bufio.Reader
	input string

	pos   int
	value string
}

func newScanner(r io.Reader, str string) *scanner {
	return &scanner{
		r:     bufio.NewReader(r),
		input: str,

		pos:   0,
		value: "",
	}
}

func (s *scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}

	s.pos++

	s.value += string(ch)

	return ch
}

func (s *scanner) unread() {
	_ = s.r.UnreadRune()

	s.pos--

	s.value = s.value[:len(s.value)-1]
}

func (s *scanner) getPosition() (int, int) {
	col := 1
	line := 1

	for _, ch := range s.input[:s.pos-len(s.value)] {
		if ch == '\n' {
			col = 1
			line++
			continue
		}
		col++
	}

	return col, line
}

func (s *scanner) emitToken(kind tokenKind) token {
	val := s.value
	if kind == tokenString {
		val = s.value[1 : len(s.value)-1]
	}

	col, line := s.getPosition()

	t := token{
		kind:     kind,
		kindName: tokenNames[kind],
		value:    val,
		col:      col,
		line:     line,
	}

	s.value = ""

	return t
}

const maxErrorValueLength = 20

func (s *scanner) emitErrorToken(msg string) token {
	val := ""
	if len(s.value) > maxErrorValueLength {
		val = fmt.Sprintf("%s : %s", msg, s.value[:maxErrorValueLength])
	} else {
		val = fmt.Sprintf("%s : %s", msg, s.value)
	}

	col, line := s.getPosition()

	t := token{
		kind:     tokenError,
		kindName: tokenNames[tokenError],
		value:    val,
		col:      col,
		line:     line,
	}

	s.value = ""

	return t
}

func (s *scanner) scan() token {
	switch ch := s.read(); {
	case isEOF(ch):
		return s.emitToken(tokenEOF)

	case isSpace(ch):
		return s.emitToken(tokenSpace)

	case isLetter(ch):
		s.unread()
		return s.scanText()

	case isNumber(ch) || ch == '-' || ch == '+':
		s.unread()
		return s.scanNumber()

	case ch == '"':
		return s.scanString()

	case isSyntaxKeyword(ch):
		return s.emitToken(tokenSyntax)
	}

	return s.emitErrorToken("unrecognized symbol")
}

func (s *scanner) scanText() token {
	buf := new(strings.Builder)
	buf.WriteRune(s.read())

loop:
	for {
		switch ch := s.read(); {
		case isEOF(ch):
			break loop

		case !isAlphaNumeric(ch):
			s.unread()
			break loop

		default:
			buf.WriteRune(ch)
		}
	}

	if _, ok := keywords[buf.String()]; ok {
		return s.emitToken(tokenKeyword)
	}

	return s.emitToken(tokenIdent)
}

func (s *scanner) scanNumber() token {
	firstCh := s.read()
	hasMore := false
	isRange := false

loop:
	for {
		switch ch := s.read(); {
		case isEOF(ch):
			break loop

		case !isNumber(ch) && ch != '.':
			if ch == '-' && isNumber(firstCh) && !isRange {
				isRange = true
				continue
			}
			s.unread()
			break loop

		default:
			hasMore = true
		}
	}

	if !hasMore {
		if firstCh == '-' || firstCh == '+' {
			return s.emitToken(tokenSyntax)
		}
	}

	if isRange {
		return s.emitToken(tokenNumberRange)
	}

	return s.emitToken(tokenNumber)
}

func (s *scanner) scanString() token {
	for {
		switch ch := s.read(); {
		case isEOF(ch):
			return s.emitErrorToken(`unclosed string, missing closing "`)

		case ch == '"':
			return s.emitToken(tokenString)
		}
	}
}
