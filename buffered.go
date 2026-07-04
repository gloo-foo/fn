package fn

import (
	"bytes"
	"context"
	"io"

	gloo "github.com/gloo-foo/framework"
)

// stream wires input through the pipeline's command, returning the output line
// stream. The caller owns the stream: consume it to completion (Collect drains)
// or hand it to a reader that Discards it.
func (p Pipeline) stream(ctx context.Context, input io.Reader) gloo.Stream[[]byte] {
	return gloo.From(ctx, gloo.ByteReaderSource([]io.Reader{input}), p.cmd)
}

// Lines runs the pipeline over input and returns the output as lines, each
// without its terminator. It uses [context.Background]; see [Pipeline.LinesContext]
// to pass a context.
func (p Pipeline) Lines(input []byte) ([]string, error) {
	return p.LinesContext(context.Background(), input)
}

// LinesContext is [Pipeline.Lines] with an explicit context: cancelling ctx stops
// the pipeline.
func (p Pipeline) LinesContext(ctx context.Context, input []byte) ([]string, error) {
	lines, err := p.stream(ctx, bytes.NewReader(input)).Collect()
	if err != nil {
		return nil, err
	}
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = string(line)
	}
	return out, nil
}

// Bytes runs the pipeline over input and returns the whole output, each result
// line terminated by '\n' — matching a shell filter. It uses [context.Background];
// see [Pipeline.BytesContext] to pass a context.
func (p Pipeline) Bytes(input []byte) ([]byte, error) {
	return p.BytesContext(context.Background(), input)
}

// BytesContext is [Pipeline.Bytes] with an explicit context: cancelling ctx stops
// the pipeline.
func (p Pipeline) BytesContext(ctx context.Context, input []byte) ([]byte, error) {
	lines, err := p.stream(ctx, bytes.NewReader(input)).Collect()
	if err != nil {
		return nil, err
	}
	return joinLines(lines), nil
}

// String is [Pipeline.Bytes] over string data: it runs the pipeline over input and
// returns the whole output as a string. It uses [context.Background]; see
// [Pipeline.StringContext] to pass a context.
func (p Pipeline) String(input string) (string, error) {
	return p.StringContext(context.Background(), input)
}

// StringContext is [Pipeline.String] with an explicit context: cancelling ctx stops
// the pipeline.
func (p Pipeline) StringContext(ctx context.Context, input string) (string, error) {
	out, err := p.BytesContext(ctx, []byte(input))
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// joinLines renders each line followed by '\n', the framing a shell filter emits.
func joinLines(lines [][]byte) []byte {
	var out []byte
	for _, line := range lines {
		out = append(out, line...)
		out = append(out, '\n')
	}
	return out
}
