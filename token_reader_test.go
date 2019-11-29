package sec

import (
	"io"
	"testing"
)

func TestReadToken(t *testing.T) {
	var r tokenReader
	r.load((`gtmdc3p(114514) ** 1919810 // 5`))
	for {
		tk, err := r.read()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Log(err)
			return
		}
		t.Log(tk)
	}
}
