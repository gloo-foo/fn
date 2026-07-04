# fn

[![CI](https://github.com/gloo-foo/fn/actions/workflows/go.yml/badge.svg)](https://github.com/gloo-foo/fn/actions/workflows/go.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/gloo-foo/fn.svg)](https://pkg.go.dev/github.com/gloo-foo/fn) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Use any [gloo-foo](https://github.com/gloo-foo) command as an ordinary Go data function.

A gloo command is a `framework.Command[[]byte, []byte]` — a transform over a stream of input lines. That shape is ideal for wiring pipelines and Unix executables ([cli](https://github.com/gloo-foo/cli)), but awkward to call from ordinary code, where you have a buffer or a reader and want a buffer or a reader back. This package closes that gap: it adapts any command — one command, or several composed — into a `Pipeline` whose method values are plain, directly-callable functions over standard `[]byte`, `string`, and `io.Reader` data.

```go
pipeline := fn.Chain(grep, sort, uniq) // fn.Pipeline
run := pipeline.String                 // func(string) (string, error)
out, err := run("my input")            // call it like any function
```

## Data shapes

All three run the same command; pick the one that fits your data.

| Method | Signature | Use |
| --- | --- | --- |
| `Bytes` | `func([]byte) ([]byte, error)` | whole buffer in, whole buffer out |
| `String` | `func(string) (string, error)` | whole string in, whole string out |
| `Lines` | `func([]byte) ([]string, error)` | output as lines, terminators stripped |
| `Reader` | `func(io.Reader) io.ReadCloser` | lazy streaming — unbounded input never buffers |

Each has a `…Context` variant that takes a `context.Context` so a caller can cancel a run.

## Composition

- `fn.Of(cmd)` — adapt a single command.
- `fn.Chain(cmds…)` — compose commands left to right into one pipeline (no arguments is the identity pipeline).
- `pipeline.To(next)` — append a stage, returning a new pipeline (the receiver is immutable).

A `Pipeline` is an immutable value: safe to copy, reuse across calls, and share across goroutines.

## Streaming

`Reader` returns a lazy `io.ReadCloser`: the command advances only as the reader is consumed, so unbounded input (`yes | head`) never buffers and backpressure flows naturally. Read to `io.EOF` or `Close` the reader — one or the other — to release the producing goroutine.

## Output framing

Output matches a shell filter: each result line is written followed by `\n`, so `fn.Of(cat).String("abc")` yields `"abc\n"` — identical to `printf 'abc' | cat`. `Lines` returns the lines without terminators.

## Package documentation

Runnable, compiler-checked usage lives in the package examples (`go test`-verified, rendered by `go doc`). Run `go doc github.com/gloo-foo/fn` for the API and example source.
