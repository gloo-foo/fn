package fn

import (
	"context"

	gloo "github.com/gloo-foo/framework"
)

// Pipeline is a gloo command — one command or several composed — in ordinary
// data-function form. Its method values ([Pipeline.Bytes], [Pipeline.String],
// [Pipeline.Lines], [Pipeline.Reader]) are plain, directly-callable functions
// over standard byte and reader data.
//
// It is an immutable value with value-receiver methods: safe to copy, reuse
// across calls, and share across goroutines. Each call runs the command over its
// own fresh input, so one Pipeline serves many invocations.
type Pipeline struct {
	cmd gloo.Command[[]byte, []byte]
}

// Of adapts a single command into a [Pipeline].
func Of(cmd gloo.Command[[]byte, []byte]) Pipeline {
	return Pipeline{cmd: cmd}
}

// Chain composes commands left to right into a single [Pipeline], feeding each
// command's output stream to the next. With no arguments it is the identity
// pipeline (input passes through unchanged).
func Chain(cmds ...gloo.Command[[]byte, []byte]) Pipeline {
	if len(cmds) == 0 {
		return Pipeline{cmd: identity()}
	}
	composed := gloo.Compose(cmds[0])
	for _, cmd := range cmds[1:] {
		composed = composed.To(cmd)
	}
	return Pipeline{cmd: composed}
}

// To returns a new [Pipeline] with next appended as the final stage. The
// receiver is unchanged.
func (p Pipeline) To(next gloo.Command[[]byte, []byte]) Pipeline {
	return Pipeline{cmd: gloo.Pipe(p.cmd, next)}
}

// identity is the no-op command: it returns its input stream unchanged. It backs
// the zero-argument [Chain] so an empty composition is a valid pipeline rather
// than an error.
func identity() gloo.Command[[]byte, []byte] {
	return gloo.FuncCommand[[]byte, []byte](
		func(_ context.Context, input gloo.Stream[[]byte]) gloo.Stream[[]byte] {
			return input
		},
	)
}
