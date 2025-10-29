// Package markdown provides Markdown parsing functionality for gopdf.
package markdown

import (
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// Parser represents a Markdown parser.
type Parser struct {
	parser *parser.Parser
}

// NewParser creates a new Markdown parser with CommonMark and GFM extensions.
func NewParser() *Parser {
	// Enable CommonMark extensions and GitHub Flavored Markdown
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	return &Parser{
		parser: p,
	}
}

// Parse parses Markdown text and returns the abstract syntax tree (AST).
func (p *Parser) Parse(markdown []byte) ast.Node {
	return p.parser.Parse(markdown)
}

// ParseString parses Markdown text from a string and returns the AST.
func (p *Parser) ParseString(markdown string) ast.Node {
	return p.Parse([]byte(markdown))
}
