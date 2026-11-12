# yamlfmt

`yamlfmt` is a YAML file formatter.

Data can be read from standard input or a list of file names.

## Installation

```sh
$ go install github.com/gechr/yamlfmt@latest
```

## Usage

```sh
# Read from files
$ yamlfmt 1.yaml 2.yaml 3.yaml

# Read from stdin
$ yamlfmt < 1.yaml
$ echo "foo: bar" | yamlfmt

# Rewrite files in place
$ yamlfmt --write 1.yaml 2.yaml

# Show what would change, without writing
$ yamlfmt --diff 1.yaml 2.yaml
```

Output to a terminal is highlighted unless `NO_COLOR` is set. With several files, each one is introduced by a `# <file>` comment.

## Behaviour

`yamlfmt` keeps comments, key order, anchors, aliases, tags and the text of every scalar (`1.10` stays `1.10`). It normalises spacing and indentation, and indents sequences under their keys. It checks every document it writes, and refuses to change a file if the formatted document would hold different data.

Known limitations:

- Folded block scalars (`>`) are re-flowed onto one line.
- Blank lines between entries are removed, and line endings become LF.
- A comment on a `---` line moves to the line below it.

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
