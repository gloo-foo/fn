package fn_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/framework/patterns"

	"github.com/gloo-foo/fn"
)

// upper is a stand-in command; a real program passes a cmd-* command (Cat, Grep,
// Sort, …) or any Command[[]byte, []byte].
func upper() gloo.Command[[]byte, []byte] {
	return patterns.Map(func(line []byte) ([]byte, error) {
		return bytes.ToUpper(line), nil
	})
}

func keepNonEmpty() gloo.Command[[]byte, []byte] {
	return patterns.Filter(func(line []byte) (bool, error) {
		return len(line) > 0, nil
	})
}

// A composed pipeline is called like an ordinary function over string data.
func Example_normalFunction() {
	pipeline := fn.Chain(upper(), keepNonEmpty())

	out, err := pipeline.String("hello\n\nworld")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%q\n", out)
	// Output: "HELLO\nWORLD\n"
}

// Bytes runs a command over a whole buffer.
func ExamplePipeline_Bytes() {
	out, _ := fn.Of(upper()).Bytes([]byte("abc"))
	fmt.Printf("%q\n", out)
	// Output: "ABC\n"
}

// Reader streams output lazily from a reader, so unbounded input never buffers.
func ExamplePipeline_Reader() {
	r := fn.Of(upper()).Reader(strings.NewReader("one\ntwo"))
	out, _ := io.ReadAll(r)
	fmt.Printf("%q\n", out)
	// Output: "ONE\nTWO\n"
}
