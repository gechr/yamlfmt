package main

import (
	"io"
	"os"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/chroma/styles"
)

type Highlighter struct {
	reader io.Reader
	writer io.Writer
}

func shouldHighlight() bool {
	// Highlighting does not make sense when writing to a file.
	if flagWrite {
		return false
	}
	// If `NO_COLOR` is set, do not highlight (https://no-color.org/).
	if _, nc := os.LookupEnv("NO_COLOR"); nc {
		return false
	}
	// If not attached to a terminal, do not highlight.
	if !isTerminal() {
		return false
	}
	return true
}

func registerStyle() {
	styles.Fallback = styles.Register(
		chroma.MustNewStyle(
			"yamlfmt",
			chroma.StyleEntries{
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
	)
}

// nolint:golint
func NewHighlighter() *Highlighter {
	if !shouldHighlight() {
		return nil
	}

	registerStyle()

	return &Highlighter{
		reader: os.Stdin,
		writer: os.Stdout,
	}
}

func (h *Highlighter) SetReader(r io.Reader) {
	h.reader = r
}

func (h *Highlighter) SetWriter(w io.Writer) {
	h.writer = w
}

func (h *Highlighter) Highlight() error {
	b, err := io.ReadAll(h.reader)
	if err != nil {
		return err
	}
	return quick.Highlight(
		h.writer,
		string(b),
		"yaml",
		"terminal16m",
		"",
	)
}
