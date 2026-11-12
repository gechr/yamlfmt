# internal/yaml

A copy of [`go.yaml.in/yaml/v4`](https://github.com/yaml/go-yaml) v4.0.0, limited to the packages yamlfmt uses, with the patches below. Import paths are rewritten to `github.com/gechr/yamlfmt/internal/yaml`. Everything else matches upstream, and `LICENSE` and `NOTICE` apply.

The repository's lint, format and `go fix` steps skip this directory, so that it stays comparable with upstream.

## Patches

Each fix has its regression tests in a `patch_*_test.go` file beside the code it changes.

1. `node_test.go`: print `Kind` and `Style` with `%v`, which Go 1.27's vet requires.
2. `emitter.go`: a document that ends in a keep-chomped block scalar (`|+`) and a comment kept gaining a newline in the scalar's value. The emitter wrote a blank line before the document's foot comment, which became part of the scalar. Test: `patch_emitter_test.go`.
3. `options.go`: `WithLineWidth` wrote to its captured width each time it was applied, so sharing the option between goroutines was a data race. Test: `patch_options_test.go`.
4. `representer.go`: dumping a `Node` changed it, removing tags and changing quoting styles, because later stages modified the caller's tree. The representer now copies caller-supplied nodes, keeping shared and cyclic aliases. Test: `patch_dumper_test.go`.
5. `parser.go`: a comment on the `---` line of a document with no content was lost when a footer followed `...`. The empty content now takes the comments that come before it. Test: `patch_parser_test.go`.
6. `parser.go`, `emitter.go`: accept `%YAML 1.2` as well as `%YAML 1.1`. Other versions are still rejected, and loading still follows the library's existing schema. Test: `patch_parser_test.go`.
