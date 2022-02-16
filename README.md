# yamlfmt

`yamlfmt` is a YAML file formatter.

Data can be read from standard input or a list of file names.

## Installation

```sh
$ go get -u github.com/gechr/yamlfmt
```

## Usage

```sh
# Read from files
$ yamlfmt 1.yaml 2.yaml 3.yaml

# Read from stdin
$ yamlfmt < 1.yaml
$ echo "foo: bar" | yamlfmt
```

## Example

```sh
$ cat example.yaml
abc:      def
def:  hij   # foobar
k:
# n
  - l
# m
  - op
q:
  # baz
  r: s
  t:   u
  v       : w
  x:
    - "y"

    - z

$ yamlfmt example.yaml
abc: def
def: hij # foobar
k:
# n
- l
# m
- op
q:
  # baz
  r: s
  t: u
  v: w
  x:
  - "y"
  - z
```
