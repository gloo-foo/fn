package fn

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"

	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/framework/patterns"
)

// errBoom is the sentinel a failing test command emits, matched with errors.Is.
var errBoom = errors.New("boom")

// upper upper-cases each input line.
func upper() gloo.Command[[]byte, []byte] {
	return patterns.Map(func(line []byte) ([]byte, error) {
		return bytes.ToUpper(line), nil
	})
}

// nonEmpty drops blank lines.
func nonEmpty() gloo.Command[[]byte, []byte] {
	return patterns.Filter(func(line []byte) (bool, error) {
		return len(line) > 0, nil
	})
}

// failing fails on the first line.
func failing() gloo.Command[[]byte, []byte] {
	return patterns.Map(func([]byte) ([]byte, error) {
		return nil, errBoom
	})
}

// failAfter passes n lines through, then fails; each Execute gets a fresh count.
func failAfter(n int) gloo.Command[[]byte, []byte] {
	return patterns.StatefulMap(func() func([]byte) ([]byte, error) {
		seen := 0
		return func(line []byte) ([]byte, error) {
			seen++
			if seen > n {
				return nil, errBoom
			}
			return line, nil
		}
	})
}

func TestOf_Bytes(t *testing.T) {
	out, err := Of(upper()).Bytes([]byte("abc\ndef"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "ABC\nDEF\n" {
		t.Fatalf("got %q, want %q", out, "ABC\nDEF\n")
	}
}

func TestString_MatchesNormalFunction(t *testing.T) {
	run := Of(upper()).String
	out, err := run("my input")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "MY INPUT\n" {
		t.Fatalf("got %q, want %q", out, "MY INPUT\n")
	}
}

func TestLines(t *testing.T) {
	out, err := Of(upper()).Lines([]byte("a\nb"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(out, []string{"A", "B"}) {
		t.Fatalf("got %v, want [A B]", out)
	}
}

func TestChain_ComposesInOrder(t *testing.T) {
	out, err := Chain(upper(), nonEmpty()).String("a\n\nb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "A\nB\n" {
		t.Fatalf("got %q, want %q", out, "A\nB\n")
	}
}

func TestChain_EmptyIsIdentity(t *testing.T) {
	out, err := Chain().String("x\ny")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "x\ny\n" {
		t.Fatalf("got %q, want %q", out, "x\ny\n")
	}
}

func TestTo_AppendsStage(t *testing.T) {
	out, err := Of(upper()).To(nonEmpty()).String("a\n\nb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "A\nB\n" {
		t.Fatalf("got %q, want %q", out, "A\nB\n")
	}
}

func TestBytes_Empty(t *testing.T) {
	out, err := Of(upper()).Bytes([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("got %q, want empty", out)
	}
}

func TestLines_Empty(t *testing.T) {
	out, err := Of(upper()).Lines([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("got %v, want empty", out)
	}
}

func TestBytes_PropagatesError(t *testing.T) {
	out, err := Of(failing()).Bytes([]byte("x"))
	if !errors.Is(err, errBoom) {
		t.Fatalf("got err %v, want errBoom", err)
	}
	if out != nil {
		t.Fatalf("got %q, want nil output", out)
	}
}

func TestString_PropagatesError(t *testing.T) {
	out, err := Of(failing()).String("x")
	if !errors.Is(err, errBoom) {
		t.Fatalf("got err %v, want errBoom", err)
	}
	if out != "" {
		t.Fatalf("got %q, want empty output", out)
	}
}

func TestLines_PropagatesError(t *testing.T) {
	out, err := Of(failing()).Lines([]byte("x"))
	if !errors.Is(err, errBoom) {
		t.Fatalf("got err %v, want errBoom", err)
	}
	if out != nil {
		t.Fatalf("got %v, want nil output", out)
	}
}

func TestPipeline_IsReusable(t *testing.T) {
	p := Chain(upper())
	first, _ := p.String("a")
	second, _ := p.String("b")
	if first != "A\n" || second != "B\n" {
		t.Fatalf("got %q and %q, want A\\n and B\\n", first, second)
	}
}

// upperThenFail composes into a real pipeline whose second stage fails midway,
// proving Chain's composed command surfaces a downstream stage's error.
func TestChain_SurfacesDownstreamError(t *testing.T) {
	_, err := Chain(upper(), failAfter(1)).Lines([]byte("a\nb\nc"))
	if !errors.Is(err, errBoom) {
		t.Fatalf("got err %v, want errBoom", err)
	}
	if !strings.Contains(errBoom.Error(), "boom") {
		t.Fatalf("sentinel changed")
	}
}
