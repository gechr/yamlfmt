package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	indent = 2
)

type Formatter struct {
	data       yaml.Node `yaml:",inline,omitempty"`
	decodeFunc func(interface{}) error
	encodeFunc func(interface{}) error
}

func NewFormatter() *Formatter {
	f := &Formatter{
		data: yaml.Node{},
	}
	// By default, read from stdin and write to stdout.
	f.SetReader(os.Stdin)
	f.SetWriter(os.Stdout)
	return f
}

func (f *Formatter) SetReader(r io.Reader) {
	dec := yaml.NewDecoder(r)
	f.decodeFunc = dec.Decode
}

func (f *Formatter) SetWriter(w io.Writer) {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(indent)
	f.encodeFunc = enc.Encode
}

func (f *Formatter) Format() error {
	var err error
	for {
		err = f.decode()
		if errors.Is(err, io.EOF) {
			// End of input stream.
			break
		}
		if err != nil {
			return fmt.Errorf("failed to decode: %w", err)
		}
		err = f.encode()
		if err != nil {
			// Given that the decode will have been successful at this point, we
			// would never expect an error here, but who knows!
			return fmt.Errorf("failed to encode: %w", err)
		}
	}
	return nil
}

func (f *Formatter) decode() error {
	return f.decodeFunc(&f.data)
}

func (f *Formatter) encode() error {
	return f.encodeFunc(&f.data)
}
