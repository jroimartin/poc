// Command loxlex demos an experimental lexer for the Lox programming
// language.
//
// Lox is the programming language that drives the amazing book
// [Crafting Interpreters] by Robert Nystrom.
//
// This implementation of the lexer is based on the also amazing talk
// [Lexical Scanning in Go] by Rob Pike.
//
// [Crafting Interpreters]: https://craftinginterpreters.com/
// [Lexical Scanning in Go]: https://youtu.be/HxaD_trXwRE
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
	for it := range lex(string(input)) {
		fmt.Printf("%-10s %s\n", it.typ, it.val)
	}
}
