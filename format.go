package main

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	defaultReader = os.Stdin
	defaultWriter = os.Stdout
)

type yml struct {
	Data map[string]interface{} `yaml:",inline,omitempty"`
}

type formatter struct {
	data       yml
	decodeFunc func(interface{}) error
	encodeFunc func(interface{}) error
}

func NewFormatter() *formatter {
	f := &formatter{
		data: yml{},
	}
	// By default, read from stdin and write to stdout
	f.SetReader(defaultReader)
	f.SetWriter(defaultWriter)
	return f
}

func (f *formatter) SetReader(r io.Reader) {
	dec := yaml.NewDecoder(r)
	f.decodeFunc = dec.Decode
}

func (f *formatter) SetWriter(w io.Writer) {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	f.encodeFunc = enc.Encode
}

func (f *formatter) Format() error {
	var err error
	for {
		err = f.decode()
		if err == io.EOF {
			// End of input stream
			break
		}
		if err != nil {
			return errors.Wrap(err, "failed to decode")
		}
		err = f.encode()
		if err != nil {
			// Given that the decode will have been successful at this point, we
			// would never expect an error here, but who knows!
			return errors.Wrap(err, "failed to encode")
		}
	}
	return nil
}

func (f *formatter) decode() error {
	err := f.decodeFunc(&f.data)
	if err != nil {
		return err
	}
	return nil
}

func (f *formatter) encode() error {
	err := f.encodeFunc(&f.data)
	if err != nil {
		return err
	}
	return nil
}
