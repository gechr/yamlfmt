package main

import (
	"bytes"
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/gechr/x/terminal"
)

// highlighter colours YAML for a terminal. It is safe for concurrent use.
type highlighter struct {
	formatter chroma.Formatter
	lexer     chroma.Lexer
	style     *chroma.Style
}

func newHighlighter() *highlighter {
	// Terminals without true-colour support show 24-bit escapes in the
	// wrong colours, or not at all.
	formatter := formatters.TTY256
	if terminal.SupportsTrueColor() {
		formatter = formatters.TTY16m
	}
	return &highlighter{
		formatter: formatter,
		lexer:     chroma.Coalesce(lexers.Get("yaml")),
		style: chroma.MustNewStyle(
			"yamlfmt",
			chroma.StyleEntries{ //nolint:exhaustive // unlisted token types inherit from their parent
				chroma.Text:                "#f8f8f2",
				chroma.Error:               "#960050 bg:#1e0010",
				chroma.Comment:             "#75715e",
				chroma.Keyword:             "#66d9ef",
				chroma.KeywordNamespace:    "#f92672",
				chroma.Operator:            "#f92672",
				chroma.Punctuation:         "#f8f8f2",
				chroma.Name:                "#f8f8f2",
				chroma.NameAttribute:       "#a6e22e",
				chroma.NameClass:           "#a6e22e",
				chroma.NameConstant:        "#66d9ef",
				chroma.NameDecorator:       "#a6e22e",
				chroma.NameException:       "#a6e22e",
				chroma.NameFunction:        "#a6e22e",
				chroma.NameOther:           "#a6e22e",
				chroma.NameTag:             "#f92672",
				chroma.LiteralNumber:       "#ae81ff",
				chroma.Literal:             "#e6db74",
				chroma.LiteralDate:         "#e6db74",
				chroma.LiteralString:       "#e6db74",
				chroma.LiteralStringEscape: "#ae81ff",
				chroma.GenericDeleted:      "#f92672",
				chroma.GenericEmph:         "italic",
				chroma.GenericInserted:     "#a6e22e",
				chroma.GenericStrong:       "bold",
				chroma.GenericSubheading:   "#75715e",
				chroma.Background:          "bg:#272822",
			},
		),
	}
}

// write highlights src and writes it to w. The formatter writes one escape
// sequence per token, so the result is collected first and written once.
func (h *highlighter) write(w io.Writer, src []byte) error {
	tokens, err := h.lexer.Tokenise(nil, string(src))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err = h.formatter.Format(&buf, h.style, tokens); err != nil {
		return err
	}
	_, err = w.Write(buf.Bytes())
	return err
}
