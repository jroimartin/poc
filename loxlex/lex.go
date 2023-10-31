package main

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// item represents a token returned by the scanner.
type item struct {
	typ itemType // Type, such as itemNumber.
	val string   // Value, such as "23.2".
}

// itemType identifies the type of lex items.
type itemType int

// Lex item types.
const (
	// Error occurred; value is text of error.
	itemError itemType = iota

	// Single-character tokens.
	itemLeftParen
	itemRightParen
	itemLeftBrace
	itemRightBrace
	itemComma
	itemDot
	itemMinus
	itemPlus
	itemSemicolon
	itemSlash
	itemStar

	// One or two character tokens.
	itemBang
	itemBangEqual
	itemEqual
	itemEqualEqual
	itemGreater
	itemGreaterEqual
	itemLess
	itemLessEqual

	// Literals.
	itemIdentifier
	itemString
	itemNumber

	// Keywords.
	itemAnd
	itemClass
	itemElse
	itemFalse
	itemFun
	itemFor
	itemIf
	itemNil
	itemOr
	itemPrint
	itemReturn
	itemSuper
	itemThis
	itemTrue
	itemVar
	itemWhile

	// End of file.
	itemEOF
)

// itemNames associates item types with the corresponding string
// representations.
var itemNames = map[itemType]string{
	itemError:        "Error",
	itemLeftParen:    "LeftParen",
	itemRightParen:   "RightParen",
	itemLeftBrace:    "LeftBrace",
	itemRightBrace:   "RightBrace",
	itemComma:        "Comma",
	itemDot:          "Dot",
	itemMinus:        "Minus",
	itemPlus:         "Plus",
	itemSemicolon:    "Semicolon",
	itemSlash:        "Slash",
	itemStar:         "Star",
	itemBang:         "Bang",
	itemBangEqual:    "BangEqual",
	itemEqual:        "Equal",
	itemEqualEqual:   "EqualEqual",
	itemGreater:      "Greater",
	itemGreaterEqual: "GreaterEqual",
	itemLess:         "Less",
	itemLessEqual:    "LessEqual",
	itemIdentifier:   "Identifier",
	itemString:       "String",
	itemNumber:       "Number",
	itemAnd:          "And",
	itemClass:        "Class",
	itemElse:         "Else",
	itemFalse:        "False",
	itemFun:          "Fun",
	itemFor:          "For",
	itemIf:           "If",
	itemNil:          "Nil",
	itemOr:           "Or",
	itemPrint:        "Print",
	itemReturn:       "Return",
	itemSuper:        "Super",
	itemThis:         "This",
	itemTrue:         "True",
	itemVar:          "Var",
	itemWhile:        "While",
	itemEOF:          "EOF",
}

func (t itemType) String() string {
	if s, ok := itemNames[t]; ok {
		return s
	}
	return "unknown"
}

// key associates keywords with the corresponding item types.
var key = map[string]itemType{
	"and":    itemAnd,
	"class":  itemClass,
	"else":   itemElse,
	"false":  itemFalse,
	"fun":    itemFun,
	"for":    itemFor,
	"if":     itemIf,
	"nil":    itemNil,
	"or":     itemOr,
	"print":  itemPrint,
	"return": itemReturn,
	"super":  itemSuper,
	"this":   itemThis,
	"true":   itemTrue,
	"var":    itemVar,
	"while":  itemWhile,
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	return fmt.Sprintf("%q", i.val)
}

// stateFn represents the state of the scanner as a function that
// returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	input string    // the string being scanned.
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of last rune read from input.
	items chan item // channel of scanned items.
}

// lex initializes the lexer to lex an input string and launches the
// state machine as a goroutine. It returns a channel of scanned
// items.
func lex(input string) <-chan item {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l.items
}

// run lexes the input by executing state functions until the state is
// nil.
func (l *lexer) run() {
	for state := lexCode; state != nil; {
		state = state(l)
	}
	close(l.items) // No more tokens will be delivered.
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// eof represents end of file.
const eof = -1

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune. Can be called only once per call of
// next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// accept consumes the next rune if it is r.
func (l *lexer) accept(r rune) bool {
	if l.next() == r {
		return true
	}
	l.backup()
	return false
}

// condFn is a function that returns whether a rune meets a given
// condition.
type condFn func(rune) bool

// not inverts the provided condition.
func not(cond condFn) condFn {
	return func(r rune) bool {
		return !cond(r)
	}
}

// acceptRun consumes a run of runes that meet the specified
// condition.
func (l *lexer) acceptRun(f condFn) {
	for f(l.next()) {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating
// [*lexer.run].
func (l *lexer) errorf(format string, args ...any) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}

// lexCode scans the elements in a piece of Lox code.
func lexCode(l *lexer) stateFn {
	switch r := l.next(); {
	case r == eof:
		l.emit(itemEOF)
		return nil
	case r == '(':
		l.emit(itemLeftParen)
	case r == ')':
		l.emit(itemRightParen)
	case r == '{':
		l.emit(itemLeftBrace)
	case r == '}':
		l.emit(itemRightBrace)
	case r == ',':
		l.emit(itemComma)
	case r == '.':
		l.emit(itemDot)
	case r == '-':
		l.emit(itemMinus)
	case r == '+':
		l.emit(itemPlus)
	case r == ';':
		l.emit(itemSemicolon)
	case r == '*':
		l.emit(itemStar)
	case r == '!':
		if l.accept('=') {
			l.emit(itemBangEqual)
			break
		}
		l.emit(itemBang)
	case r == '=':
		if l.accept('=') {
			l.emit(itemEqualEqual)
			break
		}
		l.emit(itemEqual)
	case r == '<':
		if l.accept('=') {
			l.emit(itemLessEqual)
			break
		}
		l.emit(itemLess)
	case r == '>':
		if l.accept('=') {
			l.emit(itemGreaterEqual)
			break
		}
		l.emit(itemGreater)
	case r == '/':
		if l.accept('/') {
			return lexComment
		}
		l.emit(itemSlash)
	case r == '"':
		return lexQuote
	case isSpace(r):
		l.ignore()
	case unicode.IsDigit(r):
		l.backup()
		return lexNumber
	case isAlpha(r):
		l.backup()
		return lexIdentifier
	default:
		return l.errorf("unexpected character: %c", r)
	}
	return lexCode
}

// lexComment scans a comment.
func lexComment(l *lexer) stateFn {
	l.acceptRun(not(isEOL))
	l.ignore()
	return lexCode
}

// lexQuote scans a string.
func lexQuote(l *lexer) stateFn {
	switch l.next() {
	case eof:
		return l.errorf("unclosed string")
	case '"':
		l.emit(itemString)
		return lexCode
	default:
		return lexQuote
	}
}

// lexNumber scans a number.
func lexNumber(l *lexer) stateFn {
	l.acceptRun(unicode.IsDigit)

	if l.accept('.') {
		l.acceptRun(unicode.IsDigit)
	}

	l.emit(itemNumber)
	return lexCode
}

// lexIdentifier scans an identifier.
func lexIdentifier(l *lexer) stateFn {
	l.acceptRun(isAlphaNumeric)

	word := l.input[l.start:l.pos]
	if kw, ok := key[word]; ok {
		l.emit(kw)
	} else {
		l.emit(itemIdentifier)
	}
	return lexCode
}

// isAlpha returns whether r is a letter or underscore.
func isAlpha(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

// isAlphaNumeric returns whether r is alphanumeric.
func isAlphaNumeric(r rune) bool {
	return isAlpha(r) || unicode.IsDigit(r)
}

// isSpace returns whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\r' || r == '\t' || r == '\n'
}

// isEOL returns whether r is a newline or eof.
func isEOL(r rune) bool {
	return r == '\n' || r == eof
}
