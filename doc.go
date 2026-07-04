// Package fn uses a gloo-foo command as an ordinary Go data function.
//
// A gloo command is a [github.com/gloo-foo/framework.Command][[]byte, []byte]:
// a transform over a stream of input lines. That shape is ideal for wiring
// pipelines and Unix executables, but awkward to call from ordinary code, where
// you have a buffer or a reader and want a buffer or a reader back. This package
// closes that gap: it adapts any command — one command, or several composed —
// into a [Pipeline] whose method values are plain, directly-callable functions
// over standard data types.
//
//	pipeline := fn.Chain(grep, sort, uniq) // fn.Pipeline
//	run := pipeline.String                 // func(string) (string, error)
//	out, err := run("my input")            // call it like any function
//
// Three data shapes are offered, all from the same command:
//
//   - Buffered: [Pipeline.Bytes], [Pipeline.String], [Pipeline.Lines] read the
//     whole input and return the whole output — for finite data.
//   - Streaming: [Pipeline.Reader] returns a lazy [io.ReadCloser] that pulls
//     output as it is read, so unbounded input (yes | head) never buffers.
//
// Output framing matches a shell filter: each result line is written followed by
// '\n', so fn.Of(cat).String("abc") yields "abc\n" — identical to
// printf 'abc' | cat. [Pipeline.Lines] returns the lines without terminators.
//
// A [Pipeline] is an immutable value: safe to copy, reuse, and share.
package fn
