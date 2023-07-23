package dbc

type syntaxKind uint

const (
	syntaxColon syntaxKind = iota
	syntaxComma
	syntaxLeftParen
	syntaxRightParen
	syntaxLeftSquareBrace
	syntaxRightSquareBrace
	syntaxPipe
	syntaxSemicolon
	syntaxAt
	syntaxPlus
	syntaxMinus
)

var syntaxKeywords = map[rune]syntaxKind{
	':': syntaxColon,
	',': syntaxComma,
	'(': syntaxLeftParen,
	')': syntaxRightParen,
	'[': syntaxLeftSquareBrace,
	']': syntaxRightSquareBrace,
	'|': syntaxPipe,
	';': syntaxSemicolon,
	'@': syntaxAt,
	'+': syntaxPlus,
	'-': syntaxMinus,
}

func isSyntaxKeyword(r rune) bool {
	_, ok := syntaxKeywords[r]
	return ok
}
