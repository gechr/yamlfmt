# yamlfmt

`yamlfmt` is a YAML file formatter. If [`bat`](https://github.com/sharkdp/bat) is installed, output will be piped through it for syntax highlighting.

## Installation

```sh
$ go get -u github.com/gechr/yamlfmt
```

## Usage

```sh
# Read from stdin
$ yamlfmt < FILE
$ cat FILE | yamlfmt

# Read from files
$ yamlfmt FILE
```
