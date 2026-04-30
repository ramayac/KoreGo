package common

import (
	"errors"
	"io"
)

// LimitWriter limits the amount of data written to an underlying io.Writer.
type LimitWriter struct {
	W     io.Writer
	Limit int
	Wrote int
}

func (lw *LimitWriter) Write(p []byte) (n int, err error) {
	if lw.Wrote >= lw.Limit {
		return 0, errors.New("output limit exceeded")
	}
	if lw.Wrote+len(p) > lw.Limit {
		p = p[:lw.Limit-lw.Wrote]
		n, err = lw.W.Write(p)
		lw.Wrote += n
		if err == nil {
			err = errors.New("output limit exceeded")
		}
		return n, err
	}
	n, err = lw.W.Write(p)
	lw.Wrote += n
	return n, err
}
