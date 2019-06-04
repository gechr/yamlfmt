# yamlfmt

`yamlfmt` is a YAML file formatter.

Data can be read from standard input or a list of file names.

If [`bat`](https://github.com/sharkdp/bat) is installed, it will automatically be used for syntax highlighting.

## Installation

```sh
$ go get -u github.com/gechr/yamlfmt
```

## Usage

```sh
# Read from files
$ yamlfmt example_1.yaml example_2.yaml example_3.yaml

# Read from stdin
$ yamlfmt < example_1.yaml
$ echo "foo: bar" | yamlfmt
```

## TODO

- Add `-w/--write` flag to write formatted YAML back to the source file (instead of stdout)
