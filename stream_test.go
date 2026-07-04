package fn

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/gloo-foo/framework/patterns"
)

// repeater yields chunk endlessly, an unbounded input source for laziness tests.
type repeater struct {
	chunk []byte
	off   int
}

func (r *repeater) Read(p []byte) (int, error) {
	n := 0
	for n < len(p) {
		n += copy(p[n:], r.chunk[r.off:])
		r.off = (r.off + len(r.chunk[r.off:])) % len(r.chunk)
	}
	return n, nil
}

func TestReader_ReadAll(t *testing.T) {
	out, err := io.ReadAll(Of(upper()).Reader(strings.NewReader("a\nb")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "A\nB\n" {
		t.Fatalf("got %q, want %q", out, "A\nB\n")
	}
}

func TestReader_SmallBufferReassembles(t *testing.T) {
	r := Of(upper()).Reader(strings.NewReader("ab\ncd"))
	var got []byte
	buf := make([]byte, 1)
	for {
		n, err := r.Read(buf)
		got = append(got, buf[:n]...)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if string(got) != "AB\nCD\n" {
		t.Fatalf("got %q, want %q", got, "AB\nCD\n")
	}
}

func TestReader_Empty(t *testing.T) {
	out, err := io.ReadAll(Of(upper()).Reader(strings.NewReader("")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("got %q, want empty", out)
	}
}

func TestReader_LazyOverUnboundedInput(t *testing.T) {
	// An infinite source with a Take(2) stage must terminate and never buffer
	// the whole (endless) input — proving the reader is lazy.
	pipeline := Chain(upper(), patterns.Take[[]byte](2))
	out, err := io.ReadAll(pipeline.Reader(&repeater{chunk: []byte("y\n")}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "Y\nY\n" {
		t.Fatalf("got %q, want %q", out, "Y\nY\n")
	}
}

func TestReader_PropagatesError(t *testing.T) {
	_, err := io.ReadAll(Of(failing()).Reader(strings.NewReader("x")))
	if !errors.Is(err, errBoom) {
		t.Fatalf("got err %v, want errBoom", err)
	}
}

func TestReader_ErrorMidStreamKeepsEarlierData(t *testing.T) {
	r := Of(failAfter(1)).Reader(strings.NewReader("a\nb\nc"))
	out, err := io.ReadAll(r)
	if !errors.Is(err, errBoom) {
		t.Fatalf("got err %v, want errBoom", err)
	}
	if !strings.Contains(string(out), "a") {
		t.Fatalf("got %q, want it to contain earlier line a", out)
	}
	// A read after the error keeps reporting it.
	if _, err := r.Read(make([]byte, 1)); !errors.Is(err, errBoom) {
		t.Fatalf("post-error read got %v, want errBoom", err)
	}
}

func TestReader_CloseBeforeEOF(t *testing.T) {
	r := Of(upper()).Reader(&repeater{chunk: []byte("y\n")})
	if _, err := r.Read(make([]byte, 2)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}
	// After Close, reads report the pipe-closed error and never block on the
	// unbounded upstream.
	if _, err := r.Read(make([]byte, 8)); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("post-close read got %v, want ErrClosedPipe", err)
	}
	// Close is idempotent.
	if err := r.Close(); err != nil {
		t.Fatalf("second close error: %v", err)
	}
}

func TestReader_ReadAfterEOF(t *testing.T) {
	r := Of(upper()).Reader(strings.NewReader("a"))
	if _, err := io.ReadAll(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.Read(make([]byte, 4)); !errors.Is(err, io.EOF) {
		t.Fatalf("second EOF read got %v, want EOF", err)
	}
}

func TestReaderContext_CancelStopsUnboundedStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := Of(upper()).ReaderContext(ctx, &repeater{chunk: []byte("y\n")})
	_, err := io.ReadAll(r)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got err %v, want context.Canceled", err)
	}
}

func TestBytesContext_PassesContext(t *testing.T) {
	out, err := Of(upper()).BytesContext(context.Background(), []byte("a"))
	if err != nil || string(out) != "A\n" {
		t.Fatalf("got (%q, %v), want (A\\n, nil)", out, err)
	}
}

func TestLinesContext_PassesContext(t *testing.T) {
	out, err := Of(upper()).LinesContext(context.Background(), []byte("a"))
	if err != nil || len(out) != 1 || out[0] != "A" {
		t.Fatalf("got (%v, %v), want ([A], nil)", out, err)
	}
}
