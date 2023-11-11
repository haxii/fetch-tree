package fetch_tree

import (
	"fmt"
	"io"
)

type ByteVarWriter struct {
	io.Writer
	c int
}

func newByteVarWriter(w io.Writer) *ByteVarWriter {
	return &ByteVarWriter{Writer: w}
}

func (w *ByteVarWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	for n = range p {
		if w.c%18 == 0 {
			w.Writer.Write([]byte("\n"))
		}
		fmt.Fprintf(w.Writer, "0x%02x,", p[n])
		w.c++
	}

	n++

	return
}
