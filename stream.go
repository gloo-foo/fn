package fn

import (
	"context"
	"io"

	gloo "github.com/gloo-foo/framework"
)

// Reader runs the pipeline over input and returns its output as a lazy reader:
// output is produced only as the returned reader is read, so unbounded input
// (yes | head) never buffers. Each result line is emitted followed by '\n'. Any
// error the pipeline raises surfaces from Read. Close abandons an unfinished
// read, tearing the pipeline down; reading to io.EOF releases it too — the
// caller MUST do one or the other, or the producing goroutine blocks forever
// (the io.Pipe contract). It uses [context.Background]; see
// [Pipeline.ReaderContext] to pass a context.
func (p Pipeline) Reader(input io.Reader) io.ReadCloser {
	return p.ReaderContext(context.Background(), input)
}

// ReaderContext is [Pipeline.Reader] with an explicit context: cancelling ctx
// stops the pipeline and the next Read reports the cancellation.
//
// The command runs in a goroutine that pumps its output through an [io.Pipe];
// the returned reader is the pipe's read half. Because io.Pipe is synchronous,
// the command only advances as the reader consumes it — the source of the
// laziness and backpressure.
func (p Pipeline) ReaderContext(ctx context.Context, input io.Reader) io.ReadCloser {
	pr, pw := io.Pipe()
	go func() {
		_, err := gloo.PumpBytes(ctx, p.cmd, input, pw)
		_ = pw.CloseWithError(err)
	}()
	return pr
}
